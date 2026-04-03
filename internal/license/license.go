package license

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	secretName      = "kymaros-license"
	secretNamespace = "kymaros-system"
)

// verifyKey is the HMAC signing key injected at build time via ldflags:
//
//	-ldflags "-X github.com/kymaroshq/kymaros/internal/license.verifyKey=..."
//
// In development, leave empty and set KYMAROS_LICENSE_KEY env var instead.
var verifyKey string

// hmacKeyBytes returns the license validation key. It prefers the build-time
// ldflags value and falls back to the KYMAROS_LICENSE_KEY environment variable.
// If neither is set, paid license keys cannot be verified (Community only).
func hmacKeyBytes() []byte {
	if verifyKey != "" {
		return []byte(verifyKey)
	}
	if k := os.Getenv("KYMAROS_LICENSE_KEY"); k != "" {
		return []byte(k)
	}
	return nil
}

type Tier string

const (
	TierCommunity  Tier = "community"
	TierTeam       Tier = "team"
	TierEnterprise Tier = "enterprise"
)

type FeatureFlags struct {
	MaxHistoryDays   int  `json:"maxHistoryDays"`
	CompliancePage   bool `json:"compliancePage"`
	PDFExport        bool `json:"pdfExport"`
	CSVExport        bool `json:"csvExport"`
	MultiBackup      bool `json:"multiBackup"`
	MultiCluster     bool `json:"multiCluster"`
	SSO              bool `json:"sso"`
	RTOAnalytics     bool `json:"rtoAnalytics"`
	RegressionAlerts bool `json:"regressionAlerts"`
	Timeline         bool `json:"timeline"`
	ScoreBreakdown   bool `json:"scoreBreakdown"`
}

type License struct {
	Tier        Tier         `json:"tier"`
	Key         string       `json:"-"`
	ExpiresAt   *time.Time   `json:"expiresAt,omitempty"`
	TrialEndsAt *time.Time   `json:"trialEndsAt,omitempty"`
	Features    FeatureFlags `json:"features"`
}

func CommunityFeatures() FeatureFlags {
	return FeatureFlags{
		MaxHistoryDays: 7,
	}
}

func TeamFeatures() FeatureFlags {
	return FeatureFlags{
		MaxHistoryDays:   90,
		CompliancePage:   true,
		CSVExport:        true,
		MultiBackup:      true,
		RTOAnalytics:     true,
		RegressionAlerts: true,
		Timeline:         true,
		ScoreBreakdown:   true,
	}
}

func EnterpriseFeatures() FeatureFlags {
	return FeatureFlags{
		MaxHistoryDays:   0, // unlimited
		CompliancePage:   true,
		PDFExport:        true,
		CSVExport:        true,
		MultiBackup:      true,
		MultiCluster:     true,
		SSO:              true,
		RTOAnalytics:     true,
		RegressionAlerts: true,
		Timeline:         true,
		ScoreBreakdown:   true,
	}
}

func Community() *License {
	return &License{
		Tier:     TierCommunity,
		Features: CommunityFeatures(),
	}
}

// FeaturesForTier returns the feature flags for the given tier.
func FeaturesForTier(t Tier) FeatureFlags {
	switch t {
	case TierTeam:
		return TeamFeatures()
	case TierEnterprise:
		return EnterpriseFeatures()
	default:
		return CommunityFeatures()
	}
}

// LoadFromSecret reads the kymaros-license Secret and returns the active License.
// If the Secret is missing or invalid, it returns a Community license.
func LoadFromSecret(ctx context.Context, c client.Client) *License {
	var secret corev1.Secret
	key := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
	if err := c.Get(ctx, key, &secret); err != nil {
		slog.Info("no license secret found, using community tier", "error", err)
		return Community()
	}

	tierStr := string(secret.Data["tier"])
	licenseKey := string(secret.Data["key"])
	expiresStr := string(secret.Data["expires"])

	tier := Tier(strings.ToLower(tierStr))
	if tier != TierTeam && tier != TierEnterprise {
		slog.Warn("unknown license tier, falling back to community", "tier", tierStr)
		return Community()
	}

	// Validate key
	if !validateKey(licenseKey, tierStr, expiresStr) {
		slog.Warn("invalid license key, falling back to community")
		return Community()
	}

	lic := &License{
		Tier:     tier,
		Key:      licenseKey,
		Features: FeaturesForTier(tier),
	}

	// Parse expiration
	if expiresStr != "" {
		t, err := time.Parse("2006-01-02", expiresStr)
		if err == nil {
			lic.ExpiresAt = &t
			if time.Now().UTC().After(t) {
				slog.Warn("license expired, falling back to community", "expired", expiresStr)
				return Community()
			}
		}
	}

	// Detect trial (key starts with KYM-TRIAL)
	if strings.HasPrefix(licenseKey, "KYM-TRIAL") && lic.ExpiresAt != nil {
		lic.TrialEndsAt = lic.ExpiresAt
	}

	slog.Debug("license loaded", "tier", lic.Tier, "expires", expiresStr)
	return lic
}

// IsExpired returns true if the license has an expiration date in the past.
func (l *License) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return time.Now().UTC().After(*l.ExpiresAt)
}

// ClampDays limits the requested days to the tier's maximum.
func (l *License) ClampDays(requested int) int {
	max := l.Features.MaxHistoryDays
	if max == 0 { // unlimited
		return requested
	}
	if requested > max {
		return max
	}
	return requested
}

func validateKey(key, tier, expires string) bool {
	if key == "" {
		return false
	}
	// Trial keys bypass HMAC for MVP
	if strings.HasPrefix(key, "KYM-TRIAL-") {
		return true
	}
	// HMAC validation: key format is KYM-<tier>-<hmac>
	parts := strings.SplitN(key, "-", 3)
	if len(parts) < 3 {
		return false
	}
	signingKey := hmacKeyBytes()
	if signingKey == nil {
		slog.Warn("KYMAROS_LICENSE_KEY not set, cannot validate paid license")
		return false
	}
	payload := fmt.Sprintf("%s:%s", tier, expires)
	mac := hmac.New(sha256.New, signingKey)
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))[:16]
	return hmac.Equal([]byte(parts[2]), []byte(expected))
}
