/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/client-go/kubernetes"

	restorev1alpha1 "github.com/kymaroshq/kymaros/api/v1alpha1"
	"github.com/kymaroshq/kymaros/internal/adapter"
	"github.com/kymaroshq/kymaros/internal/healthcheck"
	rpmetrics "github.com/kymaroshq/kymaros/internal/metrics"
	"github.com/kymaroshq/kymaros/internal/notify"
	"github.com/kymaroshq/kymaros/internal/report"
	"github.com/kymaroshq/kymaros/internal/sandbox"
)

const (
	finalizerName   = "kymaros.io/sandbox-cleanup"
	requeuePoll     = 10 * time.Second
	backupLatest    = "latest"
)

// RestoreTestReconciler reconciles a RestoreTest object
type RestoreTestReconciler struct {
	client.Client
	Scheme         *runtime.Scheme
	Sandbox        *sandbox.Manager
	Clientset      kubernetes.Interface // needed for streaming pod logs
	RestoreTimeout time.Duration        // max wait for Velero restore (default 5min)
	PodWaitTimeout time.Duration        // max wait for pods ready (default 2min)
}

func (r *RestoreTestReconciler) restoreTimeout() time.Duration {
	if r.RestoreTimeout > 0 {
		return r.RestoreTimeout
	}
	return 5 * time.Minute
}

func (r *RestoreTestReconciler) podWaitTimeout() time.Duration {
	if r.PodWaitTimeout > 0 {
		return r.PodWaitTimeout
	}
	return 2 * time.Minute
}

// +kubebuilder:rbac:groups=restore.kymaros.io,resources=restoretests,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=restore.kymaros.io,resources=restoretests/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=restore.kymaros.io,resources=restoretests/finalizers,verbs=update
// +kubebuilder:rbac:groups=restore.kymaros.io,resources=healthcheckpolicies,verbs=get;list;watch
// +kubebuilder:rbac:groups=restore.kymaros.io,resources=restorereports,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=restore.kymaros.io,resources=restorereports/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get
// +kubebuilder:rbac:groups="",resources=pods/exec,verbs=create
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=secrets;configmaps,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=resourcequotas;limitranges,verbs=create;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=networkpolicies,verbs=create;delete
// +kubebuilder:rbac:groups=apps,resources=deployments;statefulsets,verbs=get;list;watch;update
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch
// +kubebuilder:rbac:groups=velero.io,resources=backups,verbs=get;list;watch
// +kubebuilder:rbac:groups=velero.io,resources=restores,verbs=get;list;watch;create;delete
// +kubebuilder:rbac:groups=velero.io,resources=schedules,verbs=get;list;watch

// Reconcile runs the restore test lifecycle as a phase-based state machine.
func (r *RestoreTestReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := slog.With("controller", "RestoreTest", "name", req.Name, "namespace", req.Namespace)

	// 1. Fetch RestoreTest CR
	var test restorev1alpha1.RestoreTest
	if err := r.Get(ctx, req.NamespacedName, &test); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("fetch RestoreTest: %w", err)
	}

	// 2. Handle deletion — cleanup sandbox if exists
	if !test.DeletionTimestamp.IsZero() {
		return r.handleDeletion(ctx, &test, logger)
	}

	// 3. Ensure finalizer
	if !controllerutil.ContainsFinalizer(&test, finalizerName) {
		controllerutil.AddFinalizer(&test, finalizerName)
		if err := r.Update(ctx, &test); err != nil {
			return ctrl.Result{}, fmt.Errorf("add finalizer: %w", err)
		}
		return ctrl.Result{Requeue: true}, nil
	}

	// 4. Phase-based state machine
	logger.InfoContext(ctx, "reconcile",
		"phase", test.Status.Phase,
		"sandbox", test.Status.SandboxNamespace,
		"restoreID", test.Status.RestoreID,
		"lastScore", test.Status.LastScore,
	)

	// Apply global timeout for the Running phase if spec.timeout is set
	runCtx := ctx
	if test.Status.Phase == restorev1alpha1.PhaseRunning && test.Spec.Timeout != nil && test.Spec.Timeout.Duration > 0 {
		var cancel context.CancelFunc
		runCtx, cancel = context.WithTimeout(ctx, test.Spec.Timeout.Duration)
		defer cancel()
	}

	switch test.Status.Phase {
	case "", restorev1alpha1.PhaseIdle:
		return r.reconcileIdle(ctx, &test, logger)
	case restorev1alpha1.PhaseRunning:
		result, err := r.reconcileRunning(runCtx, &test, logger)
		// If the global timeout expired, fail the test with a clear message
		if runCtx.Err() == context.DeadlineExceeded {
			logger.ErrorContext(ctx, "global timeout exceeded", "timeout", test.Spec.Timeout.Duration)
			return r.failTest(ctx, &test, logger, fmt.Sprintf("global timeout exceeded (%s)", test.Spec.Timeout.Duration))
		}
		return result, err
	case restorev1alpha1.PhaseCompleted, restorev1alpha1.PhaseFailed:
		return r.reconcileFinished(ctx, &test, logger)
	default:
		logger.WarnContext(ctx, "unknown phase, resetting to Idle", "phase", test.Status.Phase)
		return r.setPhase(ctx, &test, restorev1alpha1.PhaseIdle)
	}
}

// handleDeletion cleans up sandbox resources before allowing CR deletion.
func (r *RestoreTestReconciler) handleDeletion(ctx context.Context, test *restorev1alpha1.RestoreTest, logger *slog.Logger) (ctrl.Result, error) { //nolint:unparam // result is always zero but signature matches Reconcile pattern
	if controllerutil.ContainsFinalizer(test, finalizerName) {
		if test.Status.SandboxNamespace != "" {
			logger.InfoContext(ctx, "cleaning up sandbox on deletion", "sandbox", test.Status.SandboxNamespace)
			if err := r.Sandbox.Cleanup(ctx, test.Status.SandboxNamespace); err != nil {
				if !errors.IsNotFound(err) {
					return ctrl.Result{}, fmt.Errorf("cleanup sandbox on deletion: %w", err)
				}
			}
		}
		controllerutil.RemoveFinalizer(test, finalizerName)
		if err := r.Update(ctx, test); err != nil {
			return ctrl.Result{}, fmt.Errorf("remove finalizer: %w", err)
		}
	}
	return ctrl.Result{}, nil
}

// reconcileIdle checks the cron schedule and transitions to Running if it's time.
func (r *RestoreTestReconciler) reconcileIdle(ctx context.Context, test *restorev1alpha1.RestoreTest, logger *slog.Logger) (ctrl.Result, error) {
	schedule, err := cron.ParseStandard(test.Spec.Schedule.Cron)
	if err != nil {
		logger.ErrorContext(ctx, "invalid cron expression", "cron", test.Spec.Schedule.Cron, "error", err)
		return ctrl.Result{}, fmt.Errorf("parse cron %q: %w", test.Spec.Schedule.Cron, err)
	}

	now := time.Now()

	// If never run before, run now
	if test.Status.LastRunAt == nil {
		logger.InfoContext(ctx, "first run — starting immediately")
		return r.setPhase(ctx, test, restorev1alpha1.PhaseRunning)
	}

	// Calculate next run from last run
	nextRun := schedule.Next(test.Status.LastRunAt.Time)
	if now.Before(nextRun) {
		delay := nextRun.Sub(now)
		logger.DebugContext(ctx, "not time yet", "nextRun", nextRun, "delay", delay)
		test.Status.NextRunAt = &metav1.Time{Time: nextRun}
		if err := r.Status().Update(ctx, test); err != nil {
			return ctrl.Result{}, fmt.Errorf("update nextRunAt: %w", err)
		}
		return ctrl.Result{RequeueAfter: delay}, nil
	}

	logger.InfoContext(ctx, "schedule due — starting test")
	return r.setPhase(ctx, test, restorev1alpha1.PhaseRunning)
}

// reconcileRunning handles the multi-step restore and validation flow.
func (r *RestoreTestReconciler) reconcileRunning(ctx context.Context, test *restorev1alpha1.RestoreTest, logger *slog.Logger) (ctrl.Result, error) {
	// Sub-step A: Create sandboxes and trigger restores
	if len(test.Status.SandboxNamespaces) == 0 && test.Status.SandboxNamespace == "" {
		return r.startRestore(ctx, test, logger)
	}

	// Sub-step B: Check if all restores are done
	if len(test.Status.RestoreIDs) > 0 || test.Status.RestoreID != "" {
		return r.checkRestore(ctx, test, logger)
	}

	logger.WarnContext(ctx, "sandbox exists but no restoreID, cleaning up")
	return r.failTest(ctx, test, logger, "inconsistent state: sandbox without restore")
}

// startRestore creates sandboxes and triggers restores for ALL namespace mappings.
func (r *RestoreTestReconciler) startRestore(ctx context.Context, test *restorev1alpha1.RestoreTest, logger *slog.Logger) (ctrl.Result, error) {
	// Pre-check: verify all source namespaces still exist
	for _, nsMapping := range test.Spec.BackupSource.Namespaces {
		var ns corev1.Namespace
		if err := r.Get(ctx, types.NamespacedName{Name: nsMapping.Name}, &ns); err != nil {
			if errors.IsNotFound(err) {
				logger.ErrorContext(ctx, "source namespace no longer exists", "namespace", nsMapping.Name)
				return r.failTest(ctx, test, logger, fmt.Sprintf("Source namespace %q no longer exists", nsMapping.Name))
			}
			return ctrl.Result{}, fmt.Errorf("check source namespace %q: %w", nsMapping.Name, err)
		}
	}

	ba, err := adapter.NewBackupAdapter(test.Spec.BackupSource.Provider, r.Client)
	if err != nil {
		return r.failTest(ctx, test, logger, fmt.Sprintf("create adapter: %v", err))
	}

	// Resolve backup name once (shared across all namespaces)
	backupName := test.Spec.BackupSource.BackupName
	if backupName == "" || backupName == backupLatest {
		sourceNS := test.Spec.BackupSource.Namespaces[0].Name
		latestBackup, err := ba.GetLatestBackup(ctx, sourceNS)
		if err != nil {
			return r.failTest(ctx, test, logger, fmt.Sprintf("get latest backup: %v", err))
		}
		backupName = latestBackup.Name
		logger.InfoContext(ctx, "resolved latest backup", "backup", backupName)
	}

	var sandboxNames []string
	var restoreIDs []string

	// Create a sandbox + restore for each namespace mapping
	for _, nsMapping := range test.Spec.BackupSource.Namespaces {
		sourceNS := nsMapping.Name
		nsName, err := r.Sandbox.Create(ctx, test.Spec.Sandbox, test.Name, test.Namespace)
		if err != nil {
			// Cleanup already-created sandboxes
			for _, ns := range sandboxNames {
				_ = r.Sandbox.Cleanup(ctx, ns)
			}
			return r.failTest(ctx, test, logger, fmt.Sprintf("create sandbox for %s: %v", sourceNS, err))
		}
		logger.InfoContext(ctx, "sandbox created", "source", sourceNS, "sandbox", nsName)

		restoreID, err := ba.TriggerRestore(ctx, adapter.RestoreOptions{
			BackupName:      backupName,
			SourceNamespace: sourceNS,
			TargetNamespace: nsName,
			LabelSelector:   test.Spec.BackupSource.LabelSelector,
		})
		if err != nil {
			_ = r.Sandbox.Cleanup(ctx, nsName)
			for _, ns := range sandboxNames {
				_ = r.Sandbox.Cleanup(ctx, ns)
			}
			return r.failTest(ctx, test, logger, fmt.Sprintf("trigger restore for %s: %v", sourceNS, err))
		}
		logger.InfoContext(ctx, "restore triggered", "source", sourceNS, "restoreID", restoreID)

		sandboxNames = append(sandboxNames, nsName)
		restoreIDs = append(restoreIDs, restoreID)
	}

	// Persist all sandbox + restore IDs
	if err := r.Get(ctx, types.NamespacedName{Name: test.Name, Namespace: test.Namespace}, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("re-fetch test: %w", err)
	}
	test.Status.SandboxNamespaces = sandboxNames
	test.Status.RestoreIDs = restoreIDs
	// Backward compat: set the singular fields to the first
	if len(sandboxNames) > 0 {
		test.Status.SandboxNamespace = sandboxNames[0]
	}
	if len(restoreIDs) > 0 {
		test.Status.RestoreID = restoreIDs[0]
	}
	if err := r.Status().Update(ctx, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("persist sandboxes+restores: %w", err)
	}

	return ctrl.Result{RequeueAfter: requeuePoll}, nil
}

// checkRestore polls the restore status and proceeds to scoring when done.
func (r *RestoreTestReconciler) checkRestore(ctx context.Context, test *restorev1alpha1.RestoreTest, logger *slog.Logger) (ctrl.Result, error) {
	// Pre-check: verify the source backup still exists
	if exists, err := r.backupExists(ctx, test.Spec.BackupSource.Provider, test.Spec.BackupSource.BackupName, test.Spec.BackupSource.Namespaces); err != nil {
		logger.WarnContext(ctx, "failed to check backup existence", "error", err)
	} else if !exists {
		logger.ErrorContext(ctx, "source backup no longer exists")
		return r.failTest(ctx, test, logger, "Source backup no longer exists")
	}

	ba, err := adapter.NewBackupAdapter(test.Spec.BackupSource.Provider, r.Client)
	if err != nil {
		return r.failTest(ctx, test, logger, fmt.Sprintf("create adapter: %v", err))
	}

	// Check ALL restores (multi-namespace support)
	restoreIDs := test.Status.RestoreIDs
	if len(restoreIDs) == 0 && test.Status.RestoreID != "" {
		restoreIDs = []string{test.Status.RestoreID}
	}

	var totalDuration time.Duration
	for _, rid := range restoreIDs {
		result, err := ba.WaitForRestore(ctx, rid, r.restoreTimeout())
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "not found") {
				return r.failTest(ctx, test, logger, fmt.Sprintf("velero restore %s disappeared: %v", rid, err))
			}
			if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "context deadline") {
				logger.InfoContext(ctx, "restore still in progress", "restoreID", rid)
				return ctrl.Result{RequeueAfter: 30 * time.Second}, nil
			}
			return r.failTest(ctx, test, logger, fmt.Sprintf("check restore %s: %v", rid, err))
		}
		if !result.Success {
			errMsg := fmt.Sprintf("restore %s failed", rid)
			if len(result.Errors) > 0 {
				errMsg = fmt.Sprintf("restore %s failed: %v", rid, result.Errors)
			}
			return r.failTest(ctx, test, logger, errMsg)
		}
		totalDuration += result.Duration
		logger.InfoContext(ctx, "restore completed", "restoreID", rid, "duration", result.Duration)
	}

	// All restores succeeded — proceed to scoring
	combinedResult := &adapter.RestoreResult{
		Success:  true,
		Duration: totalDuration,
	}
	return r.scoreAndReport(ctx, test, combinedResult, logger)
}

// scoreAndReport calculates the score, creates a RestoreReport, and cleans up.
func (r *RestoreTestReconciler) scoreAndReport(ctx context.Context, test *restorev1alpha1.RestoreTest, restoreResult *adapter.RestoreResult, logger *slog.Logger) (ctrl.Result, error) {
	now := metav1.Now()

	// Get all sandboxes
	sandboxes := test.Status.SandboxNamespaces
	if len(sandboxes) == 0 && test.Status.SandboxNamespace != "" {
		sandboxes = []string{test.Status.SandboxNamespace}
	}

	// Run all validations and compute the score
	vr := r.runValidations(ctx, test, sandboxes, restoreResult, logger)

	// Build RestoreReport CR
	reportName := fmt.Sprintf("%s-%s", test.Name, now.Format("20060102-150405"))
	rr := r.buildReport(test, reportName, now, restoreResult, vr.checkResults, vr.completeness,
		vr.completenessRatio, vr.podsReadyRatio, vr.healthChecksPassRatio, vr.totalReady, vr.totalPods,
		vr.depsLevelResult, vr.rtoWithinSLA)
	rr.Status.Score = vr.score
	rr.Status.Result = vr.result

	// Detect regressions and send notifications
	r.detectAndNotify(ctx, test, vr.score, vr.result, reportName, logger)

	// Collect pod logs and events BEFORE cleanup (sandbox is still alive)
	logCfg := test.Spec.LogCollection
	if logCfg == nil || logCfg.Enabled {
		if r.Clientset != nil {
			for _, ns := range sandboxes {
				podLogs, err := report.CollectPodLogs(ctx, r.Client, r.Clientset, ns, logCfg)
				if err != nil {
					logger.WarnContext(ctx, "failed to collect pod logs", "sandbox", ns, "error", err)
				} else {
					rr.Status.PodLogs = append(rr.Status.PodLogs, podLogs...)
				}

				if logCfg == nil || logCfg.IncludeEvents {
					events, err := report.CollectEvents(ctx, r.Client, ns)
					if err != nil {
						logger.WarnContext(ctx, "failed to collect events", "sandbox", ns, "error", err)
					} else {
						rr.Status.Events = append(rr.Status.Events, events...)
					}
				}
			}
			// Enforce size limit to avoid exceeding etcd object size
			rr.Status.PodLogs = report.TruncateLogs(rr.Status.PodLogs, report.MaxTotalLogBytes)
			logger.InfoContext(ctx, "pod logs collected", "pods", len(rr.Status.PodLogs), "events", len(rr.Status.Events))
		}
	}

	// CRITICAL: Cleanup ALL sandboxes and Velero restores FIRST.
	for _, ns := range sandboxes {
		if err := r.Sandbox.Cleanup(ctx, ns); err != nil {
			logger.WarnContext(ctx, "sandbox cleanup failed", "sandbox", ns, "error", err)
		}
	}
	restoreIDs := test.Status.RestoreIDs
	if len(restoreIDs) == 0 && test.Status.RestoreID != "" {
		restoreIDs = []string{test.Status.RestoreID}
	}
	ba, _ := adapter.NewBackupAdapter(test.Spec.BackupSource.Provider, r.Client)
	if ba != nil {
		for _, rid := range restoreIDs {
			if err := ba.CleanupRestore(ctx, rid); err != nil {
				logger.WarnContext(ctx, "velero restore cleanup failed", "restoreID", rid, "error", err)
			}
		}
	}

	// Re-fetch the test to get the latest ResourceVersion before status update
	if err := r.Get(ctx, types.NamespacedName{Name: test.Name, Namespace: test.Namespace}, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("re-fetch test: %w", err)
	}

	// Update test status to Completed
	test.Status.Phase = restorev1alpha1.PhaseCompleted
	test.Status.LastRunAt = &now
	test.Status.LastScore = vr.score
	test.Status.LastResult = vr.result
	test.Status.LastReportRef = reportName
	test.Status.SandboxNamespace = ""
	test.Status.RestoreID = ""
	test.Status.SandboxNamespaces = nil
	test.Status.RestoreIDs = nil
	if err := r.Status().Update(ctx, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("update status to Completed: %w", err)
	}

	// Create RestoreReport (non-critical — score is already persisted in test status)
	// Save status separately since Create strips the status subresource
	reportStatus := rr.Status
	rr.Status = restorev1alpha1.RestoreReportStatus{} // clear before create
	if err := r.Create(ctx, rr); err != nil {
		logger.WarnContext(ctx, "failed to create RestoreReport", "error", err)
	} else {
		// Patch status using MergeFrom (more reliable than Update for freshly created objects)
		patch := client.MergeFrom(rr.DeepCopy())
		rr.Status = reportStatus
		if err := r.Status().Patch(ctx, rr, patch); err != nil {
			logger.WarnContext(ctx, "failed to patch RestoreReport status", "error", err)
		} else {
			logger.InfoContext(ctx, "RestoreReport created", "report", reportName, "score", vr.score)
		}
	}

	// Record metrics
	rpmetrics.TestsTotal.WithLabelValues(test.Name, vr.result).Inc()
	rpmetrics.Score.WithLabelValues(test.Name).Set(float64(vr.score))
	rpmetrics.RTOSeconds.WithLabelValues(test.Name).Set(restoreResult.Duration.Seconds())
	rpmetrics.TestDuration.WithLabelValues(test.Name).Observe(restoreResult.Duration.Seconds())

	// Enforce history limit: use spec.historyLimit (default 10)
	historyLimit := int32(10)
	if test.Spec.HistoryLimit != nil {
		historyLimit = *test.Spec.HistoryLimit
	}
	r.enforceHistoryLimit(ctx, test, int(historyLimit), logger)

	return ctrl.Result{Requeue: true}, nil
}

// validationResult holds the output of all validation checks and scoring.
type validationResult struct {
	score                 int
	result                string
	totalReady, totalPods int
	podsReadyRatio        float64
	healthChecksPassRatio float64
	checkResults          []healthcheck.Result
	completenessRatio     float64
	completeness          *restorev1alpha1.CompletenessStatus
	depsCoverageRatio     float64
	depsLevelResult       *restorev1alpha1.LevelResult
	rtoWithinSLA          bool
}

// runValidations performs pod readiness, health checks, completeness, cross-NS deps,
// RTO compliance checks, and calculates the final score.
func (r *RestoreTestReconciler) runValidations(ctx context.Context, test *restorev1alpha1.RestoreTest, sandboxes []string, restoreResult *adapter.RestoreResult, logger *slog.Logger) validationResult {
	var vr validationResult

	// Wait for pods to become Ready across ALL sandboxes
	waitDeadline := time.Now().Add(r.podWaitTimeout())
	for time.Now().Before(waitDeadline) {
		vr.totalReady, vr.totalPods = 0, 0
		for _, ns := range sandboxes {
			rdy, tot, err := r.countReadyPods(ctx, ns)
			if err != nil {
				logger.WarnContext(ctx, "failed to count pods", "sandbox", ns, "error", err)
				continue
			}
			vr.totalReady += rdy
			vr.totalPods += tot
		}
		if vr.totalPods > 0 && vr.totalReady == vr.totalPods {
			logger.InfoContext(ctx, "all pods ready", "ready", vr.totalReady, "total", vr.totalPods)
			break
		}
		logger.InfoContext(ctx, "pods not ready yet", "ready", vr.totalReady, "total", vr.totalPods, "remaining", time.Until(waitDeadline).Round(time.Second))
		time.Sleep(5 * time.Second)
	}

	if vr.totalPods > 0 {
		vr.podsReadyRatio = float64(vr.totalReady) / float64(vr.totalPods)
	}
	logger.InfoContext(ctx, "pod readiness", "ready", vr.totalReady, "total", vr.totalPods)

	// Run health checks across ALL sandboxes
	if test.Spec.HealthChecks.PolicyRef != "" {
		var policy restorev1alpha1.HealthCheckPolicy
		if err := r.Get(ctx, types.NamespacedName{Name: test.Spec.HealthChecks.PolicyRef, Namespace: test.Namespace}, &policy); err != nil {
			logger.WarnContext(ctx, "failed to fetch HealthCheckPolicy", "policy", test.Spec.HealthChecks.PolicyRef, "error", err)
		} else {
			runner := healthcheck.NewRunner(r.Client, slog.Default())
			for _, ns := range sandboxes {
				results := runner.RunAll(ctx, policy.Spec.Checks, ns)
				vr.checkResults = append(vr.checkResults, results...)
			}
			vr.healthChecksPassRatio = healthcheck.PassRatio(vr.checkResults)
			logger.InfoContext(ctx, "health checks completed", "passRatio", vr.healthChecksPassRatio, "total", len(vr.checkResults))
		}
	}

	// Completeness check
	completenessChecker := report.NewCompletenessChecker(r.Client, slog.Default())
	if len(test.Spec.BackupSource.Namespaces) > 0 && len(sandboxes) > 0 {
		sourceNS := test.Spec.BackupSource.Namespaces[0].Name
		var err error
		vr.completeness, vr.completenessRatio, err = completenessChecker.Check(ctx, sourceNS, sandboxes[0])
		if err != nil {
			logger.WarnContext(ctx, "completeness check failed", "error", err)
			vr.completenessRatio = 1.0
		}
	}

	// Cross-namespace dependency coverage
	crossNSChecker := report.NewCrossNSDepsChecker(r.Client, slog.Default())
	sourceToSandbox := make(map[string]string, len(test.Spec.BackupSource.Namespaces))
	for i, ns := range test.Spec.BackupSource.Namespaces {
		if i < len(sandboxes) {
			sourceToSandbox[ns.Name] = sandboxes[i]
		}
	}
	vr.depsCoverageRatio, vr.depsLevelResult, _ = crossNSChecker.Check(ctx, sourceToSandbox)
	if vr.depsLevelResult == nil {
		vr.depsCoverageRatio = 1.0
		vr.depsLevelResult = &restorev1alpha1.LevelResult{Status: "not_tested"}
	}

	// Check RTO compliance
	vr.rtoWithinSLA = true
	if test.Spec.SLA != nil {
		vr.rtoWithinSLA = restoreResult.Duration <= test.Spec.SLA.MaxRTO.Duration
	}

	// Calculate score
	scoreInput := &report.ScoreInput{
		RestoreSucceeded:      true,
		CompletenessRatio:     vr.completenessRatio,
		PodsReadyRatio:        vr.podsReadyRatio,
		HealthChecksPassRatio: vr.healthChecksPassRatio,
		DepsCoverageRatio:     vr.depsCoverageRatio,
		RTOWithinSLA:          vr.rtoWithinSLA,
	}
	vr.score = report.CalculateScore(scoreInput)

	vr.result = restorev1alpha1.ResultPass
	if vr.score < 70 {
		vr.result = restorev1alpha1.ResultFail
	} else if vr.score < 90 {
		vr.result = restorev1alpha1.ResultPartial
	}
	logger.InfoContext(ctx, "score calculated", "score", vr.score, "result", vr.result)

	return vr
}

// buildReport constructs a RestoreReport CR from scoring results.
func (r *RestoreTestReconciler) buildReport(
	test *restorev1alpha1.RestoreTest, reportName string, now metav1.Time,
	restoreResult *adapter.RestoreResult, checkResults []healthcheck.Result,
	completeness *restorev1alpha1.CompletenessStatus, completenessRatio, podsReadyRatio, healthChecksPassRatio float64,
	totalReady, totalPods int, depsLevelResult *restorev1alpha1.LevelResult, rtoWithinSLA bool,
) *restorev1alpha1.RestoreReport {
	ensureGVK(test)

	rr := &restorev1alpha1.RestoreReport{
		ObjectMeta: metav1.ObjectMeta{
			Name:      reportName,
			Namespace: test.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: test.APIVersion,
					Kind:       test.Kind,
					Name:       test.Name,
					UID:        test.UID,
				},
			},
		},
		Spec: restorev1alpha1.RestoreReportSpec{
			TestRef: test.Name,
		},
	}

	reportChecks := make([]restorev1alpha1.CheckResult, 0, len(checkResults))
	for _, cr := range checkResults {
		reportChecks = append(reportChecks, restorev1alpha1.CheckResult{
			Name:     cr.Name,
			Status:   cr.Status,
			Duration: cr.Duration,
			Message:  cr.Message,
		})
	}

	completenessStatus := restorev1alpha1.CompletenessStatus{}
	completenessLevelStatus := "not_tested"
	if completeness != nil {
		completenessStatus = *completeness
		completenessLevelStatus = podStatusResult(completenessRatio)
	}

	healthLevelStatus := "not_tested"
	if len(checkResults) > 0 {
		healthLevelStatus = podStatusResult(healthChecksPassRatio)
	}

	rr.Status = restorev1alpha1.RestoreReportStatus{
		StartedAt:   metav1.Time{Time: now.Add(-restoreResult.Duration)},
		CompletedAt: now,
		RTO: restorev1alpha1.RTOStatus{
			Measured:  metav1.Duration{Duration: restoreResult.Duration},
			WithinSLA: rtoWithinSLA,
		},
		Backup: restorev1alpha1.BackupStatus{
			Name: test.Status.RestoreID,
		},
		Checks:       reportChecks,
		Completeness: completenessStatus,
		ValidationLevels: restorev1alpha1.ValidationLevels{
			RestoreIntegrity: restorev1alpha1.LevelResult{Status: "pass"},
			Completeness: restorev1alpha1.LevelResult{
				Status: completenessLevelStatus,
			},
			PodStartup: restorev1alpha1.LevelResult{
				Status: podStatusResult(podsReadyRatio),
				Detail: fmt.Sprintf("%d/%d pods ready", totalReady, totalPods),
			},
			InternalHealth: restorev1alpha1.LevelResult{
				Status: healthLevelStatus,
			},
			CrossNamespaceDeps: *depsLevelResult,
			RTOCompliance: restorev1alpha1.LevelResult{
				Status: boolToStatus(rtoWithinSLA),
				Detail: fmt.Sprintf("measured %s", restoreResult.Duration),
			},
		},
	}
	if test.Spec.SLA != nil {
		rr.Status.RTO.Target = test.Spec.SLA.MaxRTO
	}

	return rr
}

// detectAndNotify checks for score regressions and sends notifications.
func (r *RestoreTestReconciler) detectAndNotify(ctx context.Context, test *restorev1alpha1.RestoreTest, score int, result, reportName string, logger *slog.Logger) {
	regDetector := report.NewRegressionDetector(r.Client, slog.Default())
	regResult, err := regDetector.Detect(ctx, test, score)
	if err != nil {
		logger.WarnContext(ctx, "regression detection failed", "error", err)
	} else if regResult.Detected {
		logger.WarnContext(ctx, "SCORE REGRESSION",
			"previous", regResult.PreviousScore,
			"current", regResult.CurrentScore,
			"delta", regResult.Delta,
		)
		r.sendNotifications(ctx, test, notify.Notification{
			TestName:  test.Name,
			Score:     score,
			Result:    result,
			ReportRef: reportName,
			Message:   fmt.Sprintf("Score regression: %d → %d (%d points)", regResult.PreviousScore, regResult.CurrentScore, regResult.Delta),
		}, logger)
	}

	if result == restorev1alpha1.ResultFail && test.Spec.Notifications != nil {
		r.sendNotifications(ctx, test, notify.Notification{
			TestName:  test.Name,
			Score:     score,
			Result:    result,
			ReportRef: reportName,
			Message:   fmt.Sprintf("Restore test failed with score %d/100", score),
		}, logger)
	}
}

// enforceHistoryLimit deletes old RestoreReports beyond the limit.
func (r *RestoreTestReconciler) enforceHistoryLimit(ctx context.Context, test *restorev1alpha1.RestoreTest, limit int, logger *slog.Logger) {
	var reportList restorev1alpha1.RestoreReportList
	if err := r.List(ctx, &reportList, client.InNamespace(test.Namespace)); err != nil {
		return
	}

	var matching []restorev1alpha1.RestoreReport
	for _, rr := range reportList.Items {
		if rr.Spec.TestRef == test.Name {
			matching = append(matching, rr)
		}
	}

	if len(matching) <= limit {
		return
	}

	// Sort oldest first
	sort.Slice(matching, func(i, j int) bool {
		return matching[i].CreationTimestamp.Before(&matching[j].CreationTimestamp)
	})

	toDelete := matching[:len(matching)-limit]
	for i := range toDelete {
		if err := r.Delete(ctx, &toDelete[i]); err != nil {
			logger.WarnContext(ctx, "failed to delete old report", "report", toDelete[i].Name, "error", err)
		} else {
			logger.InfoContext(ctx, "deleted old report (history limit)", "report", toDelete[i].Name)
		}
	}
}

// reconcileFinished transitions back to Idle and calculates the next run time.
func (r *RestoreTestReconciler) reconcileFinished(ctx context.Context, test *restorev1alpha1.RestoreTest, logger *slog.Logger) (ctrl.Result, error) {
	// Ensure sandbox is cleaned up (safety net for Failed state)
	if test.Status.SandboxNamespace != "" {
		if err := r.Sandbox.Cleanup(ctx, test.Status.SandboxNamespace); err != nil {
			if !errors.IsNotFound(err) {
				logger.WarnContext(ctx, "sandbox cleanup failed", "error", err)
			}
		}
		test.Status.SandboxNamespace = ""
		test.Status.RestoreID = ""
	}

	// GC: clean up any orphaned sandbox namespaces for this test
	var nsList corev1.NamespaceList
	if err := r.List(ctx, &nsList, client.MatchingLabels{
		"kymaros.io/managed-by": "kymaros",
		"kymaros.io/test":       test.Name,
	}); err == nil {
		for _, ns := range nsList.Items {
			logger.InfoContext(ctx, "cleaning orphaned sandbox", "namespace", ns.Name)
			_ = r.Sandbox.Cleanup(ctx, ns.Name)
		}
	}

	// Calculate next run
	schedule, err := cron.ParseStandard(test.Spec.Schedule.Cron)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("parse cron for next run: %w", err)
	}

	var nextRun time.Time
	if test.Status.LastRunAt != nil {
		nextRun = schedule.Next(test.Status.LastRunAt.Time)
	} else {
		nextRun = schedule.Next(time.Now())
	}

	delay := max(time.Until(nextRun), 0)

	test.Status.Phase = restorev1alpha1.PhaseIdle
	test.Status.NextRunAt = &metav1.Time{Time: nextRun}
	if err := r.Status().Update(ctx, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("update status to Idle: %w", err)
	}

	logger.InfoContext(ctx, "transitioning to Idle", "nextRun", nextRun, "delay", delay)
	return ctrl.Result{RequeueAfter: delay}, nil
}

// failTest sets the phase to Failed, records the error, and cleans up.
func (r *RestoreTestReconciler) failTest(ctx context.Context, test *restorev1alpha1.RestoreTest, logger *slog.Logger, reason string) (ctrl.Result, error) {
	logger.ErrorContext(ctx, "test failed", "reason", reason)

	// Re-fetch to get latest ResourceVersion
	if err := r.Get(ctx, types.NamespacedName{Name: test.Name, Namespace: test.Namespace}, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("re-fetch test for fail: %w", err)
	}

	now := metav1.Now()
	test.Status.Phase = restorev1alpha1.PhaseFailed
	test.Status.LastRunAt = &now
	test.Status.LastResult = restorev1alpha1.ResultFail
	test.Status.LastScore = 0

	rpmetrics.TestsTotal.WithLabelValues(test.Name, restorev1alpha1.ResultFail).Inc()
	rpmetrics.Score.WithLabelValues(test.Name).Set(0)

	if err := r.Status().Update(ctx, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("update status to Failed: %w", err)
	}

	return ctrl.Result{Requeue: true}, nil
}

// setPhase updates the status phase and requeues.
func (r *RestoreTestReconciler) setPhase(ctx context.Context, test *restorev1alpha1.RestoreTest, phase string) (ctrl.Result, error) {
	test.Status.Phase = phase
	if err := r.Status().Update(ctx, test); err != nil {
		return ctrl.Result{}, fmt.Errorf("set phase to %s: %w", phase, err)
	}
	return ctrl.Result{Requeue: true}, nil
}

// countReadyPods lists pods in the given namespace and counts those with Ready condition.
func (r *RestoreTestReconciler) countReadyPods(ctx context.Context, namespace string) (ready, total int, err error) {
	var podList corev1.PodList
	if err := r.List(ctx, &podList, client.InNamespace(namespace)); err != nil {
		return 0, 0, fmt.Errorf("list pods in %q: %w", namespace, err)
	}

	total = len(podList.Items)
	for _, pod := range podList.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				ready++
				break
			}
		}
	}
	return ready, total, nil
}

// podStatusResult converts a ratio to a status string.
func podStatusResult(ratio float64) string {
	if ratio >= 1.0 {
		return "pass"
	}
	if ratio > 0 {
		return "partial"
	}
	return "fail"
}

// boolToStatus converts a bool to pass/fail.
func boolToStatus(b bool) string {
	if b {
		return "pass"
	}
	return "fail"
}

// SetupWithManager sets up the controller with the Manager.
func (r *RestoreTestReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&restorev1alpha1.RestoreTest{}).
		Owns(&restorev1alpha1.RestoreReport{}).
		Named("restoretest").
		Complete(r)
}

// sendNotifications sends notifications to all configured channels.
func (r *RestoreTestReconciler) sendNotifications(ctx context.Context, test *restorev1alpha1.RestoreTest, n notify.Notification, logger *slog.Logger) {
	if test.Spec.Notifications == nil {
		return
	}
	channels := test.Spec.Notifications.OnFailure
	if n.Result == restorev1alpha1.ResultPass {
		channels = test.Spec.Notifications.OnSuccess
	}
	for _, ch := range channels {
		var notifier notify.Notifier
		switch ch.Type {
		case "slack":
			// Read webhook URL from Secret
			webhookURL := r.getWebhookURL(ctx, test.Namespace, ch.WebhookSecretRef)
			if webhookURL == "" {
				logger.WarnContext(ctx, "slack webhook secret not found", "secret", ch.WebhookSecretRef)
				continue
			}
			notifier = notify.NewSlackNotifier(webhookURL, ch.Channel)
		case "webhook":
			webhookURL := r.getWebhookURL(ctx, test.Namespace, ch.WebhookSecretRef)
			if webhookURL == "" {
				logger.WarnContext(ctx, "webhook secret not found", "secret", ch.WebhookSecretRef)
				continue
			}
			notifier = notify.NewWebhookNotifier(webhookURL)
		default:
			continue
		}
		if err := notifier.Send(ctx, n); err != nil {
			logger.WarnContext(ctx, "notification send failed", "type", ch.Type, "error", err)
		} else {
			logger.InfoContext(ctx, "notification sent", "type", ch.Type, "channel", ch.Channel)
		}
	}
}

// getWebhookURL reads a webhook URL from a K8s Secret.
func (r *RestoreTestReconciler) getWebhookURL(ctx context.Context, namespace, secretName string) string {
	if secretName == "" {
		return ""
	}
	var secret corev1.Secret
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: namespace}, &secret); err != nil {
		return ""
	}
	if url, ok := secret.Data["url"]; ok {
		return string(url)
	}
	if url, ok := secret.Data["webhook-url"]; ok {
		return string(url)
	}
	return ""
}

// backupExists checks whether the source Velero Backup CR still exists.
// For "latest" backups, it verifies at least one completed backup exists for the source namespace.
func (r *RestoreTestReconciler) backupExists(ctx context.Context, provider, backupName string, namespaces []restorev1alpha1.NamespaceMapping) (bool, error) {
	ba, err := adapter.NewBackupAdapter(provider, r.Client)
	if err != nil {
		return false, err
	}
	if backupName == "" || backupName == backupLatest {
		if len(namespaces) == 0 {
			return false, nil
		}
		_, err := ba.GetLatestBackup(ctx, namespaces[0].Name)
		if err != nil {
			if strings.Contains(err.Error(), "no completed backups found") {
				return false, nil
			}
			return false, err
		}
		return true, nil
	}
	// For named backups, list and check if the name is present
	backups, err := ba.ListBackups(ctx, "")
	if err != nil {
		return false, err
	}
	for _, b := range backups {
		if b.Name == backupName {
			return true, nil
		}
	}
	return false, nil
}

// ensureGVK sets the TypeMeta on the test object (needed for ownerReferences).
func ensureGVK(test *restorev1alpha1.RestoreTest) {
	if test.APIVersion == "" {
		test.APIVersion = "restore.kymaros.io/v1alpha1"
	}
	if test.Kind == "" {
		test.Kind = "RestoreTest"
	}
}
