package controller

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"github.com/kymaroshq/kymaros/internal/sandbox"
)

const (
	testName       = "my-test"
	testBackupName = "test-backup"
	testSandboxNS  = "rp-test-sandbox"
	testRestoreID  = "rp-restore-1"
)

// --- Test helpers ---

func newTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	_ = corev1.AddToScheme(s)
	_ = networkingv1.AddToScheme(s)
	_ = restorev1alpha1.AddToScheme(s)
	for _, gvk := range []schema.GroupVersionKind{
		{Group: "velero.io", Version: "v1", Kind: "Backup"},
		{Group: "velero.io", Version: "v1", Kind: "BackupList"},
		{Group: "velero.io", Version: "v1", Kind: "Restore"},
		{Group: "velero.io", Version: "v1", Kind: "RestoreList"},
	} {
		if gvk.Kind == "BackupList" || gvk.Kind == "RestoreList" {
			s.AddKnownTypeWithName(gvk, &unstructured.UnstructuredList{})
		} else {
			s.AddKnownTypeWithName(gvk, &unstructured.Unstructured{})
		}
	}
	return s
}

func newTestReconciler(objects ...runtime.Object) *RestoreTestReconciler {
	scheme := newTestScheme()
	c := fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(objects...).
		WithStatusSubresource(&restorev1alpha1.RestoreTest{}, &restorev1alpha1.RestoreReport{}).
		Build()
	return &RestoreTestReconciler{
		Client:         c,
		Scheme:         scheme,
		Sandbox:        sandbox.NewManager(c, slog.Default()),
		RestoreTimeout: 10 * time.Second,
		PodWaitTimeout: 100 * time.Millisecond,
	}
}

func makeTest() *restorev1alpha1.RestoreTest {
	return &restorev1alpha1.RestoreTest{
		ObjectMeta: metav1.ObjectMeta{
			Name:       testName,
			Namespace:  "default",
			Finalizers: []string{finalizerName},
		},
		Spec: restorev1alpha1.RestoreTestSpec{
			BackupSource: restorev1alpha1.BackupSource{
				Provider:   "velero",
				BackupName: "test-backup",
				Namespaces: []restorev1alpha1.NamespaceMapping{{Name: "production"}},
			},
			Schedule: restorev1alpha1.ScheduleConfig{Cron: "0 3 * * *"},
			Sandbox: restorev1alpha1.SandboxConfig{
				NamespacePrefix:  "rp-test",
				NetworkIsolation: "strict",
			},
		},
	}
}

func makeBackup(phase string, namespaces []string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "velero.io", Version: "v1", Kind: "Backup"})
	obj.SetName(testBackupName)
	obj.SetNamespace("velero")
	_ = unstructured.SetNestedField(obj.Object, phase, "status", "phase")
	if len(namespaces) > 0 {
		ns := make([]interface{}, len(namespaces))
		for i, n := range namespaces {
			ns[i] = n
		}
		_ = unstructured.SetNestedSlice(obj.Object, ns, "spec", "includedNamespaces")
	}
	return obj
}

func makeRestore(name, phase string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(schema.GroupVersionKind{Group: "velero.io", Version: "v1", Kind: "Restore"})
	obj.SetName(name)
	obj.SetNamespace("velero")
	if phase != "" {
		_ = unstructured.SetNestedField(obj.Object, phase, "status", "phase")
	}
	return obj
}

func reconcile(r *RestoreTestReconciler, name string) (ctrl.Result, error) {
	return r.Reconcile(context.Background(), ctrl.Request{
		NamespacedName: types.NamespacedName{Name: name, Namespace: "default"},
	})
}

func getTest(r *RestoreTestReconciler) restorev1alpha1.RestoreTest {
	var t restorev1alpha1.RestoreTest
	_ = r.Get(context.Background(), types.NamespacedName{Name: testName, Namespace: "default"}, &t)
	return t
}

// --- Basic lifecycle tests ---

func TestReconcileNotFound(t *testing.T) {
	r := newTestReconciler()
	result, err := reconcile(r, "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, ctrl.Result{}, result)
}

func TestReconcileAddsFinalizer(t *testing.T) {
	test := makeTest()
	test.Finalizers = nil // no finalizer yet
	r := newTestReconciler(test)

	result, err := reconcile(r, testName)
	require.NoError(t, err)
	assert.True(t, result.RequeueAfter > 0 || result.Requeue) //nolint:staticcheck // testing legacy Requeue behavior

	updated := getTest(r)
	assert.Contains(t, updated.Finalizers, finalizerName)
}

func TestReconcileIdleFirstRunStartsImmediately(t *testing.T) {
	test := makeTest()
	r := newTestReconciler(test, makeBackup("Completed", []string{"production"}))

	result, err := reconcile(r, testName)
	require.NoError(t, err)
	assert.True(t, result.RequeueAfter > 0 || result.Requeue) //nolint:staticcheck // testing legacy Requeue behavior

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseRunning, updated.Status.Phase)
}

func TestReconcileIdleNotYetDue(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseIdle
	now := metav1.Now()
	test.Status.LastRunAt = &now

	r := newTestReconciler(test)
	result, err := reconcile(r, testName)
	require.NoError(t, err)
	assert.True(t, result.RequeueAfter > 0)
}

// --- startRestore: sandbox + restore trigger ---

func TestStartRestoreCreatesSandboxAndRestoreID(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	sourceNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "production"}}
	r := newTestReconciler(test, sourceNS, makeBackup("Completed", []string{"production"}))

	_, err := reconcile(r, testName)
	require.NoError(t, err)

	updated := getTest(r)
	// CRITICAL: both sandbox AND restoreID must be set in a single update
	assert.NotEmpty(t, updated.Status.SandboxNamespace, "sandbox must be persisted")
	assert.NotEmpty(t, updated.Status.RestoreID, "restoreID must be persisted")
	assert.Contains(t, updated.Status.SandboxNamespace, "rp-test-my-test-")
}

func TestStartRestoreNoBackupCleansSandbox(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Spec.BackupSource.BackupName = "latest"
	sourceNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "production"}}
	// No backups exist → GetLatestBackup will fail
	r := newTestReconciler(test, sourceNS)

	_, err := reconcile(r, testName)
	require.NoError(t, err)

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseFailed, updated.Status.Phase)

	// CRITICAL: no orphaned sandbox namespaces
	var nsList corev1.NamespaceList
	_ = r.List(context.Background(), &nsList)
	for _, ns := range nsList.Items {
		assert.NotContains(t, ns.Name, "rp-test-", "sandbox should be cleaned up on failure")
	}
}

// --- checkRestore: restore phase handling ---

func TestCheckRestoreCompleted(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Status.SandboxNamespace = testSandboxNS
	test.Status.RestoreID = testRestoreID

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testSandboxNS}}
	restore := makeRestore(testRestoreID, "Completed")
	backup := makeBackup("Completed", []string{"production"})

	r := newTestReconciler(test, sandboxNS, restore, backup)
	_, err := reconcile(r, testName)
	require.NoError(t, err)

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseCompleted, updated.Status.Phase)
	assert.NotEmpty(t, updated.Status.LastReportRef)
	// sandbox and restoreID must be cleared
	assert.Empty(t, updated.Status.SandboxNamespace)
	assert.Empty(t, updated.Status.RestoreID)
	assert.GreaterOrEqual(t, updated.Status.LastScore, 25) // at minimum, restore succeeded
}

func TestCheckRestorePartiallyFailedContinues(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Status.SandboxNamespace = testSandboxNS
	test.Status.RestoreID = "rp-restore-partial"

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testSandboxNS}}
	restore := makeRestore("rp-restore-partial", "PartiallyFailed")
	backup := makeBackup("Completed", []string{"production"})

	r := newTestReconciler(test, sandboxNS, restore, backup)
	_, err := reconcile(r, testName)
	require.NoError(t, err)

	updated := getTest(r)
	// CRITICAL: PartiallyFailed should NOT fail the test — it continues to scoring
	assert.Equal(t, restorev1alpha1.PhaseCompleted, updated.Status.Phase, "PartiallyFailed restore should complete, not fail")
}

func TestCheckRestoreFailedFailsTest(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Status.SandboxNamespace = testSandboxNS
	test.Status.RestoreID = "rp-restore-fail"

	restore := makeRestore("rp-restore-fail", "Failed")
	backup := makeBackup("Completed", []string{"production"})
	r := newTestReconciler(test, restore, backup)

	_, err := reconcile(r, testName)
	require.NoError(t, err)

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseFailed, updated.Status.Phase)
	assert.Equal(t, 0, updated.Status.LastScore)
}

func TestCheckRestoreInProgressRequeues(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Status.SandboxNamespace = testSandboxNS
	test.Status.RestoreID = "rp-restore-inprogress"

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testSandboxNS}}
	restore := makeRestore("rp-restore-inprogress", "InProgress")
	backup := makeBackup("Completed", []string{"production"})

	r := newTestReconciler(test, sandboxNS, restore, backup)
	result, err := reconcile(r, testName)
	require.NoError(t, err)

	// CRITICAL: must requeue, NOT fail
	assert.True(t, result.RequeueAfter > 0, "should requeue for in-progress restore")

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseRunning, updated.Status.Phase, "should stay Running while restore is in progress")
}

// --- scoreAndReport: report creation ---

func TestScoreAndReportCreatesReport(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Status.SandboxNamespace = testSandboxNS
	test.Status.RestoreID = testRestoreID

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testSandboxNS}}
	restore := makeRestore(testRestoreID, "Completed")
	backup := makeBackup("Completed", []string{"production"})

	r := newTestReconciler(test, sandboxNS, restore, backup)
	_, err := reconcile(r, testName)
	require.NoError(t, err)

	// Verify a RestoreReport was created
	var rrList restorev1alpha1.RestoreReportList
	err = r.List(context.Background(), &rrList)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(rrList.Items), 1, "at least one RestoreReport should be created")

	rr := rrList.Items[0]
	assert.Equal(t, "my-test", rr.Spec.TestRef)
	assert.GreaterOrEqual(t, rr.Status.Score, 25)
}

func TestScoreAndReportDoesNotLoopOnMultipleReconciles(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Status.SandboxNamespace = testSandboxNS
	test.Status.RestoreID = testRestoreID

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testSandboxNS}}
	restore := makeRestore(testRestoreID, "Completed")
	backup := makeBackup("Completed", []string{"production"})

	r := newTestReconciler(test, sandboxNS, restore, backup)

	// First reconcile: should score and complete
	_, err := reconcile(r, testName)
	require.NoError(t, err)

	// Second reconcile: should transition to Idle, NOT create another report
	_, err = reconcile(r, testName)
	require.NoError(t, err)

	// Third reconcile: should requeue for next cron
	_, err = reconcile(r, testName)
	require.NoError(t, err)

	// CRITICAL: should have exactly 1 report, not 3
	var rrList restorev1alpha1.RestoreReportList
	_ = r.List(context.Background(), &rrList)
	assert.Equal(t, 1, len(rrList.Items), "must create exactly 1 report, not loop")
}

// --- Cleanup ---

func TestSandboxCleanedUpAfterCompletion(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseRunning
	test.Status.SandboxNamespace = testSandboxNS
	test.Status.RestoreID = testRestoreID

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testSandboxNS}}
	restore := makeRestore(testRestoreID, "Completed")
	backup := makeBackup("Completed", []string{"production"})

	r := newTestReconciler(test, sandboxNS, restore, backup)
	_, err := reconcile(r, testName)
	require.NoError(t, err)

	// CRITICAL: sandbox namespace must be deleted
	var ns corev1.Namespace
	err = r.Get(context.Background(), types.NamespacedName{Name: testSandboxNS}, &ns)
	assert.True(t, errors.IsNotFound(err), "sandbox must be deleted after completion")
}

func TestSandboxCleanedUpOnDeletion(t *testing.T) {
	now := metav1.Now()
	test := makeTest()
	test.DeletionTimestamp = &now
	test.Status.SandboxNamespace = "rp-test-cleanup"

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "rp-test-cleanup"}}
	r := newTestReconciler(test, sandboxNS)

	_, err := reconcile(r, testName)
	require.NoError(t, err)

	var ns corev1.Namespace
	err = r.Get(context.Background(), types.NamespacedName{Name: "rp-test-cleanup"}, &ns)
	assert.True(t, errors.IsNotFound(err), "sandbox must be deleted on CR deletion")
}

func TestFailedStateCleansUpSandboxOnTransitionToIdle(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseFailed
	test.Status.SandboxNamespace = "rp-test-orphan"
	now := metav1.Now()
	test.Status.LastRunAt = &now

	sandboxNS := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "rp-test-orphan"}}
	r := newTestReconciler(test, sandboxNS)

	_, err := reconcile(r, testName)
	require.NoError(t, err)

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseIdle, updated.Status.Phase)
	assert.Empty(t, updated.Status.SandboxNamespace)

	// Sandbox should be deleted
	var ns corev1.Namespace
	err = r.Get(context.Background(), types.NamespacedName{Name: "rp-test-orphan"}, &ns)
	assert.True(t, errors.IsNotFound(err), "orphaned sandbox must be cleaned up")
}

// --- Phase transitions ---

func TestCompletedTransitionsToIdle(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseCompleted
	now := metav1.Now()
	test.Status.LastRunAt = &now

	r := newTestReconciler(test)
	result, err := reconcile(r, testName)
	require.NoError(t, err)
	assert.True(t, result.RequeueAfter > 0)

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseIdle, updated.Status.Phase)
	assert.NotNil(t, updated.Status.NextRunAt)
}

func TestFailedTransitionsToIdle(t *testing.T) {
	test := makeTest()
	test.Status.Phase = restorev1alpha1.PhaseFailed
	now := metav1.Now()
	test.Status.LastRunAt = &now

	r := newTestReconciler(test)
	result, err := reconcile(r, testName)
	require.NoError(t, err)
	assert.True(t, result.RequeueAfter > 0)

	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseIdle, updated.Status.Phase)
}

// --- Full flow: end-to-end multi-reconcile ---

func TestFullFlowIdleToCompletedToIdle(t *testing.T) {
	t.Skip("Skipped: fake client doesn't propagate unstructured status on Create/Update correctly")
	test := makeTest()
	backup := makeBackup("Completed", []string{"production"})
	r := newTestReconciler(test, backup)

	// Reconcile 1: Idle → Running (first run)
	_, err := reconcile(r, testName)
	require.NoError(t, err)
	updated := getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseRunning, updated.Status.Phase)

	// Reconcile 2: Running → creates sandbox + triggers restore
	_, err = reconcile(r, testName)
	require.NoError(t, err)
	updated = getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseRunning, updated.Status.Phase)
	assert.NotEmpty(t, updated.Status.SandboxNamespace)
	assert.NotEmpty(t, updated.Status.RestoreID)

	// Create a completed Velero restore to simulate Velero finishing
	restore := makeRestore(updated.Status.RestoreID, "")
	_ = r.Create(context.Background(), restore)
	// Set status after create (fake client may strip status on create for unstructured)
	_ = unstructured.SetNestedField(restore.Object, "Completed", "status", "phase")
	_ = r.Update(context.Background(), restore)

	// Reconcile 3: Running → Completed (restore done, score + report)
	_, err = reconcile(r, testName)
	require.NoError(t, err)
	updated = getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseCompleted, updated.Status.Phase)
	assert.GreaterOrEqual(t, updated.Status.LastScore, 25)

	// Reconcile 4: Completed → Idle
	result, err := reconcile(r, testName)
	require.NoError(t, err)
	updated = getTest(r)
	assert.Equal(t, restorev1alpha1.PhaseIdle, updated.Status.Phase)
	assert.True(t, result.RequeueAfter > 0)

	// Verify exactly 1 report
	var rrList restorev1alpha1.RestoreReportList
	_ = r.List(context.Background(), &rrList)
	assert.Equal(t, 1, len(rrList.Items))
}

// --- Utility tests ---

func TestCountReadyPods(t *testing.T) {
	scheme := newTestScheme()
	ns := "test-ns"
	readyPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "ready", Namespace: ns},
		Status:     corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}},
	}
	notReadyPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "not-ready", Namespace: ns},
		Status:     corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionFalse}}},
	}
	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(readyPod, notReadyPod).Build()
	r := &RestoreTestReconciler{Client: c, Scheme: scheme}

	ready, total, err := r.countReadyPods(context.Background(), ns)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, ready)
}

func TestPodStatusResult(t *testing.T) {
	assert.Equal(t, "pass", podStatusResult(1.0))
	assert.Equal(t, "partial", podStatusResult(0.5))
	assert.Equal(t, "fail", podStatusResult(0))
}

// Suppress unused import warning
var _ = time.Second
