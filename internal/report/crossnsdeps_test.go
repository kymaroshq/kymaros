package report

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newTestCrossNSDepsChecker(objects ...runtime.Object) *CrossNSDepsChecker {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	return NewCrossNSDepsChecker(c, slog.Default())
}

func TestCrossNSDepsSingleNamespace(t *testing.T) {
	checker := newTestCrossNSDepsChecker()
	ratio, result, err := checker.Check(context.Background(), map[string]string{
		"my-app": "sandbox-my-app-abc123",
	})
	require.NoError(t, err)
	assert.Equal(t, 1.0, ratio)
	assert.Equal(t, "pass", result.Status)
}

func TestCrossNSDepsAllSatisfied(t *testing.T) {
	// backend references db via env var "postgres.database.svc.cluster.local"
	objects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "sandbox-backend-abc"},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "api"}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "api"}},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "api",
							Image: "api:latest",
							Env: []corev1.EnvVar{
								{Name: "DATABASE_URL", Value: "postgres://postgres.database.svc.cluster.local:5432/app"},
							},
						}},
					},
				},
			},
		},
	}

	checker := newTestCrossNSDepsChecker(objects...)
	ratio, result, err := checker.Check(context.Background(), map[string]string{
		"backend":  "sandbox-backend-abc",
		"database": "sandbox-database-xyz",
	})
	require.NoError(t, err)
	assert.Equal(t, 1.0, ratio)
	assert.Equal(t, "pass", result.Status)
	assert.Len(t, result.Tested, 1)
	assert.Empty(t, result.NotTested)
}

func TestCrossNSDepsMissing(t *testing.T) {
	// backend references "database" namespace, but database is NOT in the restore group
	objects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "sandbox-backend-abc"},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "api"}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "api"}},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "api",
							Image: "api:latest",
							Env: []corev1.EnvVar{
								{Name: "DATABASE_URL", Value: "postgres://pg.database.svc:5432/app"},
							},
						}},
					},
				},
			},
		},
	}

	checker := newTestCrossNSDepsChecker(objects...)
	// only "backend" is in the group, "database" is missing
	ratio, result, err := checker.Check(context.Background(), map[string]string{
		"backend":  "sandbox-backend-abc",
		"frontend": "sandbox-frontend-def",
	})
	require.NoError(t, err)
	assert.Equal(t, 0.0, ratio)
	assert.Equal(t, "fail", result.Status)
	assert.Len(t, result.NotTested, 1)
}

func TestCrossNSDepsPartial(t *testing.T) {
	// backend references both "database" (in group) and "cache" (not in group)
	objects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "sandbox-backend-abc"},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "api"}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "api"}},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "api",
							Image: "api:latest",
							Env: []corev1.EnvVar{
								{Name: "DATABASE_URL", Value: "pg.database.svc.cluster.local:5432"},
								{Name: "REDIS_URL", Value: "redis.cache.svc:6379"},
							},
						}},
					},
				},
			},
		},
	}

	checker := newTestCrossNSDepsChecker(objects...)
	ratio, result, err := checker.Check(context.Background(), map[string]string{
		"backend":  "sandbox-backend-abc",
		"database": "sandbox-database-xyz",
	})
	require.NoError(t, err)
	assert.Equal(t, 0.5, ratio)
	assert.Equal(t, "partial", result.Status)
	assert.Len(t, result.Tested, 1)
	assert.Len(t, result.NotTested, 1)
}

func TestCrossNSDepsNoCrossRefs(t *testing.T) {
	// multi-namespace but no cross-NS DNS references found
	objects := []runtime.Object{
		&appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "sandbox-backend-abc"},
			Spec: appsv1.DeploymentSpec{
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "api"}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": "api"}},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "api",
							Image: "api:latest",
							Env: []corev1.EnvVar{
								{Name: "PORT", Value: "8080"},
							},
						}},
					},
				},
			},
		},
	}

	checker := newTestCrossNSDepsChecker(objects...)
	ratio, result, err := checker.Check(context.Background(), map[string]string{
		"backend":  "sandbox-backend-abc",
		"database": "sandbox-database-xyz",
	})
	require.NoError(t, err)
	assert.Equal(t, 1.0, ratio)
	assert.Equal(t, "pass", result.Status)
}

func TestCrossNSDepsExternalNameService(t *testing.T) {
	objects := []runtime.Object{
		&corev1.Service{
			ObjectMeta: metav1.ObjectMeta{Name: "db-proxy", Namespace: "sandbox-backend-abc"},
			Spec: corev1.ServiceSpec{
				Type:         corev1.ServiceTypeExternalName,
				ExternalName: "postgres.database.svc.cluster.local",
			},
		},
	}

	checker := newTestCrossNSDepsChecker(objects...)
	ratio, result, err := checker.Check(context.Background(), map[string]string{
		"backend":  "sandbox-backend-abc",
		"database": "sandbox-database-xyz",
	})
	require.NoError(t, err)
	assert.Equal(t, 1.0, ratio)
	assert.Equal(t, "pass", result.Status)
	assert.Len(t, result.Tested, 1)
}

func TestExtractNamespacesFromDNS(t *testing.T) {
	tests := []struct {
		value     string
		excludeNS string
		expected  []string
	}{
		{"postgres.database.svc.cluster.local:5432", "backend", []string{"database"}},
		{"redis.cache.svc:6379", "backend", []string{"cache"}},
		{"http://api.backend.svc/health", "frontend", []string{"backend"}},
		{"localhost:5432", "backend", nil},
		{"plain-value", "backend", nil},
		// self-reference should be excluded
		{"api.backend.svc:8080", "backend", nil},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			result := extractNamespacesFromDNS(tt.value, tt.excludeNS)
			assert.Equal(t, tt.expected, result)
		})
	}
}
