package license

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(s)
	return s
}

func TestManagerInitialState(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testScheme()).Build()
	m := NewManager(c)

	assert.Equal(t, TierCommunity, m.Tier())
	assert.Equal(t, CommunityFeatures(), m.Features())
	assert.Equal(t, 7, m.ClampDays(30))
	assert.True(t, m.HasTier(TierCommunity))
	assert.False(t, m.HasTier(TierTeam))
	assert.False(t, m.HasTier(TierEnterprise))
}

func TestManagerRefreshNoSecret(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testScheme()).Build()
	m := NewManager(c)
	m.Refresh(context.Background())

	// Still community when no secret exists.
	assert.Equal(t, TierCommunity, m.Tier())
}

func TestManagerRefreshWithTrialSecret(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			"tier":    []byte("team"),
			"key":     []byte("KYM-TRIAL-test123"),
			"expires": []byte("2099-12-31"),
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(testScheme()).
		WithObjects(secret).
		Build()

	m := NewManager(c)
	m.Refresh(context.Background())

	assert.Equal(t, TierTeam, m.Tier())
	assert.True(t, m.HasTier(TierTeam))
	assert.Equal(t, 90, m.Features().MaxHistoryDays)
	assert.True(t, m.Features().CompliancePage)

	lic := m.Current()
	require.NotNil(t, lic.TrialEndsAt, "trial key should set TrialEndsAt")
}

func TestManagerHasTier(t *testing.T) {
	tests := []struct {
		current Tier
		check   Tier
		expect  bool
	}{
		{TierCommunity, TierCommunity, true},
		{TierCommunity, TierTeam, false},
		{TierCommunity, TierEnterprise, false},
		{TierTeam, TierCommunity, true},
		{TierTeam, TierTeam, true},
		{TierTeam, TierEnterprise, false},
		{TierEnterprise, TierCommunity, true},
		{TierEnterprise, TierTeam, true},
		{TierEnterprise, TierEnterprise, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.current)+">="+string(tt.check), func(t *testing.T) {
			c := fake.NewClientBuilder().WithScheme(testScheme()).Build()
			m := NewManager(c)
			// Manually set the cached license for the test.
			m.cached = &License{
				Tier:     tt.current,
				Features: FeaturesForTier(tt.current),
			}
			assert.Equal(t, tt.expect, m.HasTier(tt.check))
		})
	}
}

func TestManagerCurrentReturnsCopy(t *testing.T) {
	c := fake.NewClientBuilder().WithScheme(testScheme()).Build()
	m := NewManager(c)

	lic1 := m.Current()
	lic2 := m.Current()

	// Must be different pointers.
	assert.NotSame(t, lic1, lic2)
	// But same content.
	assert.Equal(t, lic1.Tier, lic2.Tier)
}

func TestManagerConcurrentAccess(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: secretNamespace,
		},
		Data: map[string][]byte{
			"tier":    []byte("team"),
			"key":     []byte("KYM-TRIAL-concurrent"),
			"expires": []byte("2099-12-31"),
		},
	}

	c := fake.NewClientBuilder().
		WithScheme(testScheme()).
		WithObjects(secret).
		Build()

	m := NewManager(c)

	var wg sync.WaitGroup
	for range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			m.Refresh(context.Background())
		}()
		go func() {
			defer wg.Done()
			_ = m.Tier()
			_ = m.Features()
			_ = m.Current()
			_ = m.ClampDays(30)
			_ = m.HasTier(TierTeam)
		}()
	}
	wg.Wait()

	// If we get here without panic or data race, the test passes.
	assert.Equal(t, TierTeam, m.Tier())
}
