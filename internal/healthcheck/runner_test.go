package healthcheck

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

func newTestRunner(objects ...runtime.Object) *Runner {
	scheme := runtime.NewScheme()
	_ = corev1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	return NewRunner(c, slog.Default())
}

func TestCheckPodStatusPass(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-1",
			Namespace: "sandbox",
			Labels:    map[string]string{"app": "api"},
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionTrue},
			},
		},
	}

	r := newTestRunner(pod)
	result := r.CheckPodStatus(context.Background(), restorev1alpha1.PodStatusCheck{
		LabelSelector: map[string]string{"app": "api"},
		MinReady:      1,
	}, "sandbox")

	assert.Equal(t, "pass", result.Status)
	assert.Contains(t, result.Message, "1/1 pods ready")
}

func TestCheckPodStatusFail(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "api-1",
			Namespace: "sandbox",
			Labels:    map[string]string{"app": "api"},
		},
		Status: corev1.PodStatus{
			Conditions: []corev1.PodCondition{
				{Type: corev1.PodReady, Status: corev1.ConditionFalse},
			},
		},
	}

	r := newTestRunner(pod)
	result := r.CheckPodStatus(context.Background(), restorev1alpha1.PodStatusCheck{
		LabelSelector: map[string]string{"app": "api"},
		MinReady:      1,
	}, "sandbox")

	assert.Equal(t, "fail", result.Status)
}

func TestCheckPodStatusMinReady(t *testing.T) {
	pods := []runtime.Object{
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "api-1", Namespace: "sandbox", Labels: map[string]string{"app": "api"}},
			Status:     corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "api-2", Namespace: "sandbox", Labels: map[string]string{"app": "api"}},
			Status:     corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}},
		},
	}

	r := newTestRunner(pods...)
	result := r.CheckPodStatus(context.Background(), restorev1alpha1.PodStatusCheck{
		LabelSelector: map[string]string{"app": "api"},
		MinReady:      3,
	}, "sandbox")

	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Message, "2/2 pods ready (min: 3 required)")
}

func TestCheckResourceExistsPass(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "api-creds", Namespace: "sandbox"},
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{Name: "api-config", Namespace: "sandbox"},
	}

	r := newTestRunner(secret, cm)
	result := r.CheckResourceExists(context.Background(), restorev1alpha1.ResourceExistsCheck{
		Resources: []restorev1alpha1.ResourceRef{
			{Kind: "Secret", Name: "api-creds"},
			{Kind: "ConfigMap", Name: "api-config"},
		},
	}, "sandbox")

	assert.Equal(t, "pass", result.Status)
	assert.Contains(t, result.Message, "2 resources exist")
}

func TestCheckResourceExistsMissing(t *testing.T) {
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "api-creds", Namespace: "sandbox"},
	}

	r := newTestRunner(secret)
	result := r.CheckResourceExists(context.Background(), restorev1alpha1.ResourceExistsCheck{
		Resources: []restorev1alpha1.ResourceRef{
			{Kind: "Secret", Name: "api-creds"},
			{Kind: "ConfigMap", Name: "missing-config"},
		},
	}, "sandbox")

	assert.Equal(t, "fail", result.Status)
	assert.Contains(t, result.Message, "ConfigMap/missing-config")
}

func TestRunAllDispatch(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "api-1", Namespace: "sandbox", Labels: map[string]string{"app": "api"}},
		Status:     corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: "sandbox"},
	}

	r := newTestRunner(pod, secret)
	checks := []restorev1alpha1.HealthCheck{
		{
			Name: "pods-ready",
			Type: "podStatus",
			PodStatus: &restorev1alpha1.PodStatusCheck{
				LabelSelector: map[string]string{"app": "api"},
				MinReady:      1,
			},
		},
		{
			Name: "creds-exist",
			Type: "resourceExists",
			ResourceExists: &restorev1alpha1.ResourceExistsCheck{
				Resources: []restorev1alpha1.ResourceRef{{Kind: "Secret", Name: "creds"}},
			},
		},
		{
			Name: "exec-check",
			Type: "exec",
		},
	}

	results := r.RunAll(context.Background(), checks, "sandbox")
	require.Len(t, results, 3)

	assert.Equal(t, "pods-ready", results[0].Name)
	assert.Equal(t, "pass", results[0].Status)

	assert.Equal(t, "creds-exist", results[1].Name)
	assert.Equal(t, "pass", results[1].Status)

	assert.Equal(t, "exec-check", results[2].Name)
	assert.Equal(t, "skip", results[2].Status)
}

func TestPassRatio(t *testing.T) {
	tests := []struct {
		name    string
		results []Result
		want    float64
	}{
		{"all pass", []Result{{Status: "pass"}, {Status: "pass"}}, 1.0},
		{"half pass", []Result{{Status: "pass"}, {Status: "fail"}}, 0.5},
		{"none pass", []Result{{Status: "fail"}, {Status: "fail"}}, 0.0},
		{"empty", nil, 0.0},
		{"with skip", []Result{{Status: "pass"}, {Status: "skip"}, {Status: "fail"}}, 1.0 / 3.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.InDelta(t, tt.want, PassRatio(tt.results), 0.01)
		})
	}
}
