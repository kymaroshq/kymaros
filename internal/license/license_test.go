package license

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// --- FeaturesForTier ---

func TestFeaturesForTierCommunity(t *testing.T) {
	f := FeaturesForTier(TierCommunity)
	assert.Equal(t, 7, f.MaxHistoryDays)
	assert.False(t, f.CompliancePage)
	assert.False(t, f.PDFExport)
	assert.False(t, f.CSVExport)
	assert.False(t, f.MultiBackup)
	assert.False(t, f.MultiCluster)
	assert.False(t, f.SSO)
	assert.False(t, f.RTOAnalytics)
	assert.False(t, f.RegressionAlerts)
	assert.False(t, f.Timeline)
	assert.False(t, f.ScoreBreakdown)
}

func TestFeaturesForTierTeam(t *testing.T) {
	f := FeaturesForTier(TierTeam)
	assert.Equal(t, 90, f.MaxHistoryDays)
	assert.True(t, f.CompliancePage)
	assert.False(t, f.PDFExport, "Team should not have PDF export")
	assert.True(t, f.CSVExport)
	assert.True(t, f.MultiBackup)
	assert.False(t, f.MultiCluster, "Team should not have multi-cluster")
	assert.False(t, f.SSO, "Team should not have SSO")
	assert.True(t, f.RTOAnalytics)
	assert.True(t, f.RegressionAlerts)
	assert.True(t, f.Timeline)
	assert.True(t, f.ScoreBreakdown)
}

func TestFeaturesForTierEnterprise(t *testing.T) {
	f := FeaturesForTier(TierEnterprise)
	assert.Equal(t, 0, f.MaxHistoryDays, "enterprise should be unlimited (0)")
	assert.True(t, f.CompliancePage)
	assert.True(t, f.PDFExport)
	assert.True(t, f.CSVExport)
	assert.True(t, f.MultiBackup)
	assert.True(t, f.MultiCluster)
	assert.True(t, f.SSO)
	assert.True(t, f.RTOAnalytics)
	assert.True(t, f.RegressionAlerts)
	assert.True(t, f.Timeline)
	assert.True(t, f.ScoreBreakdown)
}

func TestFeaturesForTierUnknown(t *testing.T) {
	f := FeaturesForTier(Tier("unknown"))
	assert.Equal(t, CommunityFeatures(), f, "unknown tier should return community features")
}

// --- ClampDays ---

func TestClampDays(t *testing.T) {
	tests := []struct {
		name      string
		tier      Tier
		requested int
		expected  int
	}{
		{"community clamps 30 to 7", TierCommunity, 30, 7},
		{"community allows 5", TierCommunity, 5, 5},
		{"community clamps exactly 7", TierCommunity, 7, 7},
		{"team clamps 120 to 90", TierTeam, 120, 90},
		{"team allows 30", TierTeam, 30, 30},
		{"enterprise unlimited", TierEnterprise, 365, 365},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lic := &License{
				Tier:     tt.tier,
				Features: FeaturesForTier(tt.tier),
			}
			assert.Equal(t, tt.expected, lic.ClampDays(tt.requested))
		})
	}
}

// --- IsExpired ---

func TestIsExpired(t *testing.T) {
	past := time.Now().UTC().Add(-24 * time.Hour)
	future := time.Now().UTC().Add(24 * time.Hour)

	tests := []struct {
		name     string
		expires  *time.Time
		expected bool
	}{
		{"nil expires is not expired", nil, false},
		{"past date is expired", &past, true},
		{"future date is not expired", &future, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lic := &License{ExpiresAt: tt.expires}
			assert.Equal(t, tt.expected, lic.IsExpired())
		})
	}
}

// --- Community defaults ---

func TestCommunityDefaults(t *testing.T) {
	lic := Community()
	assert.Equal(t, TierCommunity, lic.Tier)
	assert.Equal(t, CommunityFeatures(), lic.Features)
	assert.Nil(t, lic.ExpiresAt)
	assert.Nil(t, lic.TrialEndsAt)
	assert.Empty(t, lic.Key)
}

const testSigningKey = "test-signing-key-12345"

// --- validateKey ---

func TestValidateKeyEmpty(t *testing.T) {
	assert.False(t, validateKey("", "team", "2027-12-31"))
}

func TestValidateKeyTrialBypass(t *testing.T) {
	assert.True(t, validateKey("KYM-TRIAL-abc123", "team", "2027-12-31"))
}

func TestValidateKeyValidHMAC(t *testing.T) {
	// Set the verifyKey for this test.
	oldKey := verifyKey
	verifyKey = testSigningKey
	defer func() { verifyKey = oldKey }()

	tier := "team"
	expires := "2027-12-31"
	payload := fmt.Sprintf("%s:%s", tier, expires)
	mac := hmac.New(sha256.New, []byte(verifyKey))
	mac.Write([]byte(payload))
	hmacHex := hex.EncodeToString(mac.Sum(nil))[:16]
	key := fmt.Sprintf("KYM-%s-%s", tier, hmacHex)

	assert.True(t, validateKey(key, tier, expires))
}

func TestValidateKeyInvalidHMAC(t *testing.T) {
	oldKey := verifyKey
	verifyKey = testSigningKey
	defer func() { verifyKey = oldKey }()

	assert.False(t, validateKey("KYM-team-0000000000000000", "team", "2027-12-31"))
}

func TestValidateKeyMalformed(t *testing.T) {
	oldKey := verifyKey
	verifyKey = testSigningKey
	defer func() { verifyKey = oldKey }()

	assert.False(t, validateKey("garbage", "team", "2027-12-31"))
	assert.False(t, validateKey("KYM-team", "team", "2027-12-31"))
}

func TestValidateKeyNoSigningKey(t *testing.T) {
	oldKey := verifyKey
	verifyKey = ""
	defer func() { verifyKey = oldKey }()

	// Without verifyKey set, paid keys cannot be validated.
	assert.False(t, validateKey("KYM-team-abcdef1234567890", "team", "2027-12-31"))
}

// --- tierRank ---

func TestTierRank(t *testing.T) {
	assert.Equal(t, 0, tierRank(TierCommunity))
	assert.Equal(t, 1, tierRank(TierTeam))
	assert.Equal(t, 2, tierRank(TierEnterprise))
	assert.Equal(t, 0, tierRank(Tier("unknown")))
}
