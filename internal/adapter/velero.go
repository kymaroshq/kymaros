package adapter

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"sort"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	backupGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "Backup",
	}
	backupListGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "BackupList",
	}
	restoreGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "Restore",
	}
	scheduleListGVK = schema.GroupVersionKind{
		Group:   "velero.io",
		Version: "v1",
		Kind:    "ScheduleList",
	}
)

const (
	pollInterval    = 5 * time.Second
	labelManagedBy  = "kymaros.io/managed-by"
	managedByValue  = "kymaros"
	defaultVeleroNS = "velero"
)

// VeleroAdapter implements BackupAdapter for Velero using unstructured objects
type VeleroAdapter struct {
	client    client.Client
	logger    *slog.Logger
	namespace string // Velero install namespace
}

// NewVeleroAdapter creates a new VeleroAdapter
func NewVeleroAdapter(c client.Client) *VeleroAdapter {
	return &VeleroAdapter{
		client:    c,
		logger:    slog.With("adapter", "velero"),
		namespace: defaultVeleroNS,
	}
}

// NewVeleroAdapterWithNamespace creates a VeleroAdapter targeting a specific namespace
func NewVeleroAdapterWithNamespace(c client.Client, namespace string) *VeleroAdapter {
	if namespace == "" {
		namespace = defaultVeleroNS
	}
	return &VeleroAdapter{
		client:    c,
		logger:    slog.With("adapter", "velero"),
		namespace: namespace,
	}
}

// ListBackups returns Velero backups that include the given namespace and have status Completed.
// Results are sorted by creation time (newest first).
func (v *VeleroAdapter) ListBackups(ctx context.Context, namespace string) ([]BackupInfo, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(backupListGVK)

	if err := v.client.List(ctx, list, client.InNamespace(v.namespace)); err != nil {
		return nil, fmt.Errorf("list velero backups: %w", err)
	}

	var result []BackupInfo
	for _, item := range list.Items {
		// Check status phase
		phase, _, _ := unstructured.NestedString(item.Object, "status", "phase")
		if phase != "Completed" && phase != "PartiallyFailed" {
			continue
		}

		// Check if backup includes the requested namespace
		if namespace != "" {
			includedNS, _, _ := unstructured.NestedStringSlice(item.Object, "spec", "includedNamespaces")
			if !containsOrWildcard(includedNS, namespace) {
				continue
			}
		}

		createdAt := item.GetCreationTimestamp().Time

		result = append(result, BackupInfo{
			Name:      item.GetName(),
			Namespace: v.namespace,
			CreatedAt: createdAt,
			Status:    phase,
		})
	}

	// Sort newest first
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	v.logger.InfoContext(ctx, "listed backups", "count", len(result), "filterNamespace", namespace)
	return result, nil
}

// GetLatestBackup returns the most recent completed backup that includes the given namespace
func (v *VeleroAdapter) GetLatestBackup(ctx context.Context, namespace string) (*BackupInfo, error) {
	backups, err := v.ListBackups(ctx, namespace)
	if err != nil {
		return nil, err
	}
	if len(backups) == 0 {
		return nil, fmt.Errorf("no completed backups found for namespace %q", namespace)
	}
	return &backups[0], nil
}

// TriggerRestore creates a Velero Restore CR that restores into the target sandbox namespace.
// Returns the name of the created Restore CR.
func (v *VeleroAdapter) TriggerRestore(ctx context.Context, opts RestoreOptions) (string, error) {
	veleroNS := opts.VeleroNamespace
	if veleroNS == "" {
		veleroNS = v.namespace
	}

	suffix, err := randomSuffix(6)
	if err != nil {
		return "", fmt.Errorf("generate restore name: %w", err)
	}

	restoreName := fmt.Sprintf("rp-%s-%s", opts.BackupName, suffix)
	if len(restoreName) > 63 {
		restoreName = restoreName[:63]
	}

	restore := &unstructured.Unstructured{}
	restore.SetGroupVersionKind(restoreGVK)
	restore.SetName(restoreName)
	restore.SetNamespace(veleroNS)
	restore.SetLabels(map[string]string{
		labelManagedBy: managedByValue,
	})

	spec := map[string]any{
		"backupName":         opts.BackupName,
		"includedNamespaces": []any{opts.SourceNamespace},
		"namespaceMapping": map[string]any{
			opts.SourceNamespace: opts.TargetNamespace,
		},
	}

	if len(opts.LabelSelector) > 0 {
		matchLabels := make(map[string]any, len(opts.LabelSelector))
		for k, val := range opts.LabelSelector {
			matchLabels[k] = val
		}
		spec["labelSelector"] = map[string]any{
			"matchLabels": matchLabels,
		}
	}

	if err := unstructured.SetNestedMap(restore.Object, spec, "spec"); err != nil {
		return "", fmt.Errorf("set restore spec: %w", err)
	}

	if err := v.client.Create(ctx, restore); err != nil {
		return "", fmt.Errorf("create velero restore %q: %w", restoreName, err)
	}

	v.logger.InfoContext(ctx, "velero restore triggered",
		"restore", restoreName,
		"backup", opts.BackupName,
		"source", opts.SourceNamespace,
		"target", opts.TargetNamespace,
	)
	return restoreName, nil
}

// WaitForRestore polls the Velero Restore CR until it reaches a terminal phase or the timeout expires.
func (v *VeleroAdapter) WaitForRestore(ctx context.Context, restoreID string, timeout time.Duration) (*RestoreResult, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	startTime := time.Now()

	for {
		restore := &unstructured.Unstructured{}
		restore.SetGroupVersionKind(restoreGVK)

		if err := v.client.Get(ctx, types.NamespacedName{Name: restoreID, Namespace: v.namespace}, restore); err != nil {
			if errors.IsNotFound(err) {
				return nil, fmt.Errorf("velero restore %q not found", restoreID)
			}
			return nil, fmt.Errorf("get velero restore %q: %w", restoreID, err)
		}

		phase, _, _ := unstructured.NestedString(restore.Object, "status", "phase")

		switch phase {
		case "Completed":
			warnings := countField(restore, "status", "warnings")
			v.logger.InfoContext(ctx, "velero restore completed", "restore", restoreID, "warnings", warnings)
			return &RestoreResult{
				Success:  true,
				Duration: time.Since(startTime),
				Warnings: collectValidationErrors(restore, "status", "warnings"),
			}, nil

		case "Failed":
			v.logger.ErrorContext(ctx, "velero restore failed", "restore", restoreID)
			return &RestoreResult{
				Success:  false,
				Duration: time.Since(startTime),
				Errors:   collectValidationErrors(restore, "status", "errors"),
			}, nil

		case "PartiallyFailed":
			warnings := collectValidationErrors(restore, "status", "warnings")
			errs := collectValidationErrors(restore, "status", "errors")
			v.logger.WarnContext(ctx, "velero restore partially failed", "restore", restoreID)
			return &RestoreResult{
				Success:  true, // treat as partial success — health checks will determine actual quality
				Duration: time.Since(startTime),
				Warnings: append(warnings, errs...),
				Errors:   errs,
			}, nil
		}

		// Not terminal yet — wait and retry
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for velero restore %q (last phase: %s): %w", restoreID, phase, ctx.Err())
		case <-time.After(pollInterval):
			v.logger.InfoContext(ctx, "waiting for velero restore", "restore", restoreID, "phase", phase, "elapsed", time.Since(startTime).Round(time.Second))
		}
	}
}

// CleanupRestore deletes the Velero Restore CR
func (v *VeleroAdapter) CleanupRestore(ctx context.Context, restoreID string) error {
	restore := &unstructured.Unstructured{}
	restore.SetGroupVersionKind(restoreGVK)
	restore.SetName(restoreID)
	restore.SetNamespace(v.namespace)

	if err := v.client.Delete(ctx, restore); err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("delete velero restore %q: %w", restoreID, err)
	}
	v.logger.InfoContext(ctx, "velero restore cleaned up", "restore", restoreID)
	return nil
}

// containsOrWildcard checks if the slice contains the value or a wildcard "*"
func containsOrWildcard(slice []string, value string) bool {
	if len(slice) == 0 {
		// Empty includedNamespaces means "all namespaces" in Velero
		return true
	}
	for _, s := range slice {
		if s == value || s == "*" {
			return true
		}
	}
	return false
}

// countField reads an int64 field from an unstructured object
func countField(obj *unstructured.Unstructured, fields ...string) int64 {
	val, found, err := unstructured.NestedInt64(obj.Object, fields...)
	if err != nil || !found {
		return 0
	}
	return val
}

// collectValidationErrors reads a string slice or returns a count-based message
func collectValidationErrors(obj *unstructured.Unstructured, fields ...string) []string {
	// Velero stores warnings/errors as int counts, not string arrays
	count := countField(obj, fields...)
	if count > 0 {
		return []string{fmt.Sprintf("%d %s", count, fields[len(fields)-1])}
	}
	return nil
}

// ListSchedules discovers Velero backup schedules.
func (v *VeleroAdapter) ListSchedules(ctx context.Context) ([]ScheduleInfo, error) {
	list := &unstructured.UnstructuredList{}
	list.SetGroupVersionKind(scheduleListGVK)

	if err := v.client.List(ctx, list, client.InNamespace(v.namespace)); err != nil {
		return nil, fmt.Errorf("list velero schedules: %w", err)
	}

	var result []ScheduleInfo
	for _, item := range list.Items {
		cronExpr, _, _ := unstructured.NestedString(item.Object, "spec", "schedule")
		includedNS, _, _ := unstructured.NestedStringSlice(item.Object, "spec", "template", "includedNamespaces")
		lastBackup, _, _ := unstructured.NestedString(item.Object, "status", "lastBackup")

		result = append(result, ScheduleInfo{
			Name:       item.GetName(),
			Cron:       cronExpr,
			Namespaces: includedNS,
			LastBackup: lastBackup,
		})
	}

	v.logger.InfoContext(ctx, "listed velero schedules", "count", len(result))
	return result, nil
}

func randomSuffix(n int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[idx.Int64()]
	}
	return string(b), nil
}
