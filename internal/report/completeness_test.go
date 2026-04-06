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

func newTestCompletenessChecker(objects ...runtime.Object) *CompletenessChecker {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	_ = appsv1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	return NewCompletenessChecker(c, slog.Default())
}

func TestCompletenessFullMatch(t *testing.T) {
	objects := []runtime.Object{
		// Source
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "source"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api-svc", Namespace: "source"}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "source"}},
		// Sandbox (same resources)
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "sandbox"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api-svc", Namespace: "sandbox"}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "sandbox"}},
	}

	checker := newTestCompletenessChecker(objects...)
	status, ratio, err := checker.Check(context.Background(), "source", "sandbox")
	require.NoError(t, err)
	assert.Equal(t, 1.0, ratio)
	assert.Equal(t, "1/1", status.Deployments)
	assert.Equal(t, "1/1", status.Services)
	assert.Equal(t, "1/1", status.Secrets)
}

func TestCompletenessPartial(t *testing.T) {
	objects := []runtime.Object{
		// Source has 3 resources
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "source"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api-svc", Namespace: "source"}},
		&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "source"}},
		// Sandbox has only 1
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "sandbox"}},
	}

	checker := newTestCompletenessChecker(objects...)
	status, ratio, err := checker.Check(context.Background(), "source", "sandbox")
	require.NoError(t, err)
	assert.InDelta(t, 1.0/3.0, ratio, 0.01)
	assert.Equal(t, "1/1", status.Deployments)
	assert.Equal(t, "0/1", status.Services)
	assert.Equal(t, "0/1", status.Secrets)
}

func TestCompletenessEmpty(t *testing.T) {
	checker := newTestCompletenessChecker()
	_, ratio, err := checker.Check(context.Background(), "source", "sandbox")
	require.NoError(t, err)
	assert.Equal(t, 0.0, ratio)
}
