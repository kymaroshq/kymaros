package license

import (
	"context"
	"log/slog"
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// LicenseManager maintains a thread-safe cache of the active license.
// It is shared between the operator (refreshed by the reconciler) and
// the dashboard (refreshed by a periodic goroutine).
type LicenseManager struct {
	client client.Client
	mu     sync.RWMutex
	cached *License
}

// NewManager creates a LicenseManager initialised with a Community license.
func NewManager(c client.Client) *LicenseManager {
	return &LicenseManager{
		client: c,
		cached: Community(),
	}
}

// Refresh re-reads the kymaros-license Secret and updates the cache.
// Safe to call from multiple goroutines.
func (m *LicenseManager) Refresh(ctx context.Context) {
	newLic := LoadFromSecret(ctx, m.client)

	m.mu.Lock()
	oldTier := m.cached.Tier
	m.cached = newLic
	m.mu.Unlock()

	if oldTier != newLic.Tier {
		slog.Info("license tier changed", "from", oldTier, "to", newLic.Tier)
	}
}

// Current returns a snapshot of the active license.
func (m *LicenseManager) Current() *License {
	m.mu.RLock()
	defer m.mu.RUnlock()
	lic := *m.cached // copy
	return &lic
}

// Tier returns the active tier.
func (m *LicenseManager) Tier() Tier {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cached.Tier
}

// Features returns the active feature flags.
func (m *LicenseManager) Features() FeatureFlags {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cached.Features
}

// ClampDays limits the requested days to the tier's maximum.
func (m *LicenseManager) ClampDays(requested int) int {
	return m.Current().ClampDays(requested)
}

// HasTier returns true if the current tier is at least minTier.
func (m *LicenseManager) HasTier(min Tier) bool {
	return tierRank(m.Tier()) >= tierRank(min)
}

func tierRank(t Tier) int {
	switch t {
	case TierTeam:
		return 1
	case TierEnterprise:
		return 2
	default:
		return 0
	}
}
