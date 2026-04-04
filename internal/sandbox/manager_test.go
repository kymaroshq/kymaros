package sandbox

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

func newTestManager() *Manager {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = networkingv1.AddToScheme(scheme)

	c := fake.NewClientBuilder().WithScheme(scheme).Build()
	logger := slog.Default()
	return NewManager(c, logger)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name       string
		config     restorev1alpha1.SandboxConfig
		testName   string
		testNS     string
		wantPrefix string
		wantLabels map[string]string
	}{
		{
			name: "default config creates namespace with strict isolation",
			config: restorev1alpha1.SandboxConfig{
				NetworkIsolation: "strict",
			},
			testName:   "mytest",
			testNS:     "default",
			wantPrefix: "rp-test-mytest-",
			wantLabels: map[string]string{
				LabelManagedBy:     ManagedByValue,
				LabelTestName:      "mytest",
				LabelTestNamespace: "default",
				LabelGroup:         "mytest",
			},
		},
		{
			name: "custom prefix",
			config: restorev1alpha1.SandboxConfig{
				NamespacePrefix:  "sandbox",
				NetworkIsolation: "strict",
			},
			testName:   "backup-test",
			testNS:     "prod",
			wantPrefix: "sandbox-backup-test-",
			wantLabels: map[string]string{
				LabelManagedBy:     ManagedByValue,
				LabelTestName:      "backup-test",
				LabelTestNamespace: "prod",
				LabelGroup:         "backup-test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestManager()
			ctx := context.Background()

			nsName, err := m.Create(ctx, tt.config, tt.testName, tt.testNS)
			require.NoError(t, err)
			assert.Contains(t, nsName, tt.wantPrefix[:len(tt.wantPrefix)-1]) // strip trailing dash for contains check

			// Verify namespace exists with correct labels
			var ns corev1.Namespace
			err = m.client.Get(ctx, types.NamespacedName{Name: nsName}, &ns)
			require.NoError(t, err)
			for k, v := range tt.wantLabels {
				assert.Equal(t, v, ns.Labels[k], "label %s", k)
			}
		})
	}
}

func TestCreateAppliesDenyAll(t *testing.T) {
	m := newTestManager()
	ctx := context.Background()

	config := restorev1alpha1.SandboxConfig{
		NetworkIsolation: "strict",
	}
	nsName, err := m.Create(ctx, config, "test-deny", "default")
	require.NoError(t, err)

	// Verify deny-all NetworkPolicy exists
	var np networkingv1.NetworkPolicy
	err = m.client.Get(ctx, types.NamespacedName{Name: "deny-all", Namespace: nsName}, &np)
	require.NoError(t, err)
	assert.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeIngress)
	assert.Contains(t, np.Spec.PolicyTypes, networkingv1.PolicyTypeEgress)
	assert.Empty(t, np.Spec.Ingress, "deny-all should have no ingress rules")
	assert.Empty(t, np.Spec.Egress, "deny-all should have no egress rules")
}

func TestCreateAppliesGroupAllow(t *testing.T) {
	m := newTestManager()
	ctx := context.Background()

	config := restorev1alpha1.SandboxConfig{
		NetworkIsolation: "group",
	}
	nsName, err := m.Create(ctx, config, "test-group", "default")
	require.NoError(t, err)

	// Verify allow-group NetworkPolicy exists
	var np networkingv1.NetworkPolicy
	err = m.client.Get(ctx, types.NamespacedName{Name: "allow-group", Namespace: nsName}, &np)
	require.NoError(t, err)

	require.Len(t, np.Spec.Ingress, 1)
	require.Len(t, np.Spec.Ingress[0].From, 1)
	assert.Equal(t, "test-group", np.Spec.Ingress[0].From[0].NamespaceSelector.MatchLabels[LabelGroup])

	require.Len(t, np.Spec.Egress, 1)
	require.Len(t, np.Spec.Egress[0].To, 1)
	assert.Equal(t, "test-group", np.Spec.Egress[0].To[0].NamespaceSelector.MatchLabels[LabelGroup])
}

func TestCreateAppliesResourceQuota(t *testing.T) {
	m := newTestManager()
	ctx := context.Background()

	config := restorev1alpha1.SandboxConfig{
		NetworkIsolation: "strict",
		ResourceQuota: &restorev1alpha1.ResourceQuotaConfig{
			CPU:    "4",
			Memory: "8Gi",
		},
	}
	nsName, err := m.Create(ctx, config, "test-quota", "default")
	require.NoError(t, err)

	// Verify ResourceQuota
	var quota corev1.ResourceQuota
	err = m.client.Get(ctx, types.NamespacedName{Name: "sandbox-quota", Namespace: nsName}, &quota)
	require.NoError(t, err)
	assert.Equal(t, resource.MustParse("4"), quota.Spec.Hard[corev1.ResourceLimitsCPU])
	assert.Equal(t, resource.MustParse("8Gi"), quota.Spec.Hard[corev1.ResourceLimitsMemory])
}

func TestCreateAppliesLimitRange(t *testing.T) {
	m := newTestManager()
	ctx := context.Background()

	config := restorev1alpha1.SandboxConfig{
		NetworkIsolation: "strict",
	}
	nsName, err := m.Create(ctx, config, "test-limits", "default")
	require.NoError(t, err)

	// Verify LimitRange
	var lr corev1.LimitRange
	err = m.client.Get(ctx, types.NamespacedName{Name: "sandbox-limits", Namespace: nsName}, &lr)
	require.NoError(t, err)
	require.Len(t, lr.Spec.Limits, 1)
	assert.Equal(t, corev1.LimitTypeContainer, lr.Spec.Limits[0].Type)
	assert.Equal(t, resource.MustParse("500m"), lr.Spec.Limits[0].Default[corev1.ResourceCPU])
	assert.Equal(t, resource.MustParse("512Mi"), lr.Spec.Limits[0].Default[corev1.ResourceMemory])
	assert.Equal(t, resource.MustParse("100m"), lr.Spec.Limits[0].DefaultRequest[corev1.ResourceCPU])
	assert.Equal(t, resource.MustParse("128Mi"), lr.Spec.Limits[0].DefaultRequest[corev1.ResourceMemory])
}

func TestCleanup(t *testing.T) {
	m := newTestManager()
	ctx := context.Background()

	// Create a namespace first
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "test-cleanup-ns"},
	}
	require.NoError(t, m.client.Create(ctx, ns))

	// Cleanup should delete it
	err := m.Cleanup(ctx, "test-cleanup-ns")
	require.NoError(t, err)

	// Verify it's gone
	var deleted corev1.Namespace
	err = m.client.Get(ctx, types.NamespacedName{Name: "test-cleanup-ns"}, &deleted)
	assert.Error(t, err, "namespace should be deleted")
}

func TestIsExpired(t *testing.T) {
	m := newTestManager()

	tests := []struct {
		name    string
		created time.Time
		ttl     time.Duration
		want    bool
	}{
		{
			name:    "not expired",
			created: time.Now().Add(-5 * time.Minute),
			ttl:     30 * time.Minute,
			want:    false,
		},
		{
			name:    "expired",
			created: time.Now().Add(-45 * time.Minute),
			ttl:     30 * time.Minute,
			want:    true,
		},
		{
			name:    "zero TTL never expires",
			created: time.Now().Add(-24 * time.Hour),
			ttl:     0,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns := &corev1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(tt.created),
				},
			}
			assert.Equal(t, tt.want, m.IsExpired(ns, tt.ttl))
		})
	}
}

func TestNamespaceTruncation(t *testing.T) {
	m := newTestManager()
	ctx := context.Background()

	config := restorev1alpha1.SandboxConfig{
		NamespacePrefix:  "very-long-prefix-name",
		NetworkIsolation: "strict",
	}
	nsName, err := m.Create(ctx, config, "very-long-test-name-that-exceeds-limits", "default")
	require.NoError(t, err)
	assert.LessOrEqual(t, len(nsName), 63, "namespace name must be <= 63 chars")
}

func TestNoQuotaWhenNilConfig(t *testing.T) {
	m := newTestManager()
	ctx := context.Background()

	config := restorev1alpha1.SandboxConfig{
		NetworkIsolation: "strict",
		ResourceQuota:    nil, // no quota
	}
	nsName, err := m.Create(ctx, config, "test-noquota", "default")
	require.NoError(t, err)

	// ResourceQuota should NOT exist
	var quota corev1.ResourceQuota
	err = m.client.Get(ctx, types.NamespacedName{Name: "sandbox-quota", Namespace: nsName}, &quota)
	assert.Error(t, err, "no ResourceQuota should be created when config is nil")
}
