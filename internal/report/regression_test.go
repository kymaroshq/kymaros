package report

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
)

func newTestRegressionDetector(objects ...runtime.Object) *RegressionDetector {
	scheme := runtime.NewScheme()
	_ = restorev1alpha1.AddToScheme(scheme)
	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	return NewRegressionDetector(c, slog.Default())
}

func TestDetectNoRegression(t *testing.T) {
	prev := &restorev1alpha1.RestoreReport{
		ObjectMeta: metav1.ObjectMeta{Name: "test-prev", Namespace: "default", CreationTimestamp: metav1.Now()},
		Spec:       restorev1alpha1.RestoreReportSpec{TestRef: "my-test"},
		Status:     restorev1alpha1.RestoreReportStatus{Score: 90},
	}

	d := newTestRegressionDetector(prev)
	test := &restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	result, err := d.Detect(context.Background(), test, 92)
	require.NoError(t, err)
	assert.False(t, result.Detected)
	assert.Equal(t, 90, result.PreviousScore)
	assert.Equal(t, 2, result.Delta)
}

func TestDetectRegression(t *testing.T) {
	prev := &restorev1alpha1.RestoreReport{
		ObjectMeta: metav1.ObjectMeta{Name: "test-prev", Namespace: "default", CreationTimestamp: metav1.Now()},
		Spec:       restorev1alpha1.RestoreReportSpec{TestRef: "my-test"},
		Status:     restorev1alpha1.RestoreReportStatus{Score: 90},
	}

	d := newTestRegressionDetector(prev)
	test := &restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	result, err := d.Detect(context.Background(), test, 50)
	require.NoError(t, err)
	assert.True(t, result.Detected)
	assert.Equal(t, 90, result.PreviousScore)
	assert.Equal(t, -40, result.Delta)
}

func TestDetectNoPreviousReport(t *testing.T) {
	d := newTestRegressionDetector()
	test := &restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	result, err := d.Detect(context.Background(), test, 85)
	require.NoError(t, err)
	assert.False(t, result.Detected)
	assert.Equal(t, 85, result.CurrentScore)
}

func TestDetectSmallDropNotRegression(t *testing.T) {
	prev := &restorev1alpha1.RestoreReport{
		ObjectMeta: metav1.ObjectMeta{Name: "test-prev", Namespace: "default", CreationTimestamp: metav1.Now()},
		Spec:       restorev1alpha1.RestoreReportSpec{TestRef: "my-test"},
		Status:     restorev1alpha1.RestoreReportStatus{Score: 90},
	}

	d := newTestRegressionDetector(prev)
	test := &restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{Name: "my-test", Namespace: "default"},
	}

	// Drop of 5 points — not a regression (threshold is >10)
	result, err := d.Detect(context.Background(), test, 85)
	require.NoError(t, err)
	assert.False(t, result.Detected)
}
