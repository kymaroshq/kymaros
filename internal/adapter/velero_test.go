package adapter

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func newFakeVeleroAdapter(objects ...runtime.Object) *VeleroAdapter {
	scheme := runtime.NewScheme()

	// Register the Velero GVKs so the fake client can handle them
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "velero.io", Version: "v1", Kind: "Backup"},
		&unstructured.Unstructured{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "velero.io", Version: "v1", Kind: "BackupList"},
		&unstructured.UnstructuredList{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "velero.io", Version: "v1", Kind: "Restore"},
		&unstructured.Unstructured{},
	)
	scheme.AddKnownTypeWithName(
		schema.GroupVersionKind{Group: "velero.io", Version: "v1", Kind: "RestoreList"},
		&unstructured.UnstructuredList{},
	)

	c := fake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objects...).Build()
	return NewVeleroAdapterWithNamespace(c, "velero")
}

func makeBackup(name string, createdAt time.Time, phase string, includedNamespaces []string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(backupGVK)
	obj.SetName(name)
	obj.SetNamespace("velero")
	obj.SetCreationTimestamp(metav1.NewTime(createdAt))

	spec := map[string]any{}
	if len(includedNamespaces) > 0 {
		ns := make([]any, len(includedNamespaces))
		for i, n := range includedNamespaces {
			ns[i] = n
		}
		spec["includedNamespaces"] = ns
	}
	_ = unstructured.SetNestedMap(obj.Object, spec, "spec")
	_ = unstructured.SetNestedField(obj.Object, phase, "status", "phase")

	return obj
}

func makeRestore(name string, phase string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(restoreGVK)
	obj.SetName(name)
	obj.SetNamespace("velero")

	if phase != "" {
		_ = unstructured.SetNestedField(obj.Object, phase, "status", "phase")
	}

	return obj
}

func TestListBackups(t *testing.T) {
	now := time.Now()
	older := now.Add(-2 * time.Hour)
	newest := now.Add(-30 * time.Minute)

	adapter := newFakeVeleroAdapter(
		makeBackup("backup-old", older, "Completed", []string{"production"}),
		makeBackup("backup-new", newest, "Completed", []string{"production"}),
		makeBackup("backup-failed", now, "Failed", []string{"production"}),
	)

	backups, err := adapter.ListBackups(context.Background(), "production")
	require.NoError(t, err)

	// Should only have 2 (Completed), newest first
	require.Len(t, backups, 2)
	assert.Equal(t, "backup-new", backups[0].Name)
	assert.Equal(t, "backup-old", backups[1].Name)
}

func TestListBackupsFiltersByNamespace(t *testing.T) {
	now := time.Now()

	adapter := newFakeVeleroAdapter(
		makeBackup("backup-prod", now, "Completed", []string{"production"}),
		makeBackup("backup-staging", now.Add(-1*time.Hour), "Completed", []string{"staging"}),
		makeBackup("backup-all", now.Add(-2*time.Hour), "Completed", []string{"*"}),
		makeBackup("backup-empty", now.Add(-3*time.Hour), "Completed", nil), // nil = all namespaces
	)

	backups, err := adapter.ListBackups(context.Background(), "production")
	require.NoError(t, err)

	// Should match: backup-prod (exact), backup-all (wildcard), backup-empty (nil = all)
	require.Len(t, backups, 3)

	names := make([]string, len(backups))
	for i, b := range backups {
		names[i] = b.Name
	}
	assert.Contains(t, names, "backup-prod")
	assert.Contains(t, names, "backup-all")
	assert.Contains(t, names, "backup-empty")
	assert.NotContains(t, names, "backup-staging")
}

func TestGetLatestBackup(t *testing.T) {
	now := time.Now()

	adapter := newFakeVeleroAdapter(
		makeBackup("backup-old", now.Add(-2*time.Hour), "Completed", []string{"production"}),
		makeBackup("backup-new", now.Add(-30*time.Minute), "Completed", []string{"production"}),
	)

	latest, err := adapter.GetLatestBackup(context.Background(), "production")
	require.NoError(t, err)
	assert.Equal(t, "backup-new", latest.Name)
}

func TestGetLatestBackupNone(t *testing.T) {
	adapter := newFakeVeleroAdapter()

	_, err := adapter.GetLatestBackup(context.Background(), "production")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no completed backups found")
}

func TestTriggerRestore(t *testing.T) {
	adapter := newFakeVeleroAdapter()
	ctx := context.Background()

	opts := RestoreOptions{
		BackupName:      "daily-backup",
		SourceNamespace: "production",
		TargetNamespace: "rp-test-prod-abc123",
	}

	restoreID, err := adapter.TriggerRestore(ctx, opts)
	require.NoError(t, err)
	assert.Contains(t, restoreID, "rp-daily-backup-")

	// Verify the Restore CR was created with correct spec
	restore := &unstructured.Unstructured{}
	restore.SetGroupVersionKind(restoreGVK)
	err = adapter.client.Get(ctx, types.NamespacedName{Name: restoreID, Namespace: "velero"}, restore)
	require.NoError(t, err)

	backupName, _, _ := unstructured.NestedString(restore.Object, "spec", "backupName")
	assert.Equal(t, "daily-backup", backupName)

	nsMapping, _, _ := unstructured.NestedStringMap(restore.Object, "spec", "namespaceMapping")
	assert.Equal(t, "rp-test-prod-abc123", nsMapping["production"])

	includedNS, _, _ := unstructured.NestedStringSlice(restore.Object, "spec", "includedNamespaces")
	assert.Equal(t, []string{"production"}, includedNS)

	labels := restore.GetLabels()
	assert.Equal(t, "kymaros", labels["kymaros.io/managed-by"])
}

func TestTriggerRestoreWithLabelSelector(t *testing.T) {
	adapter := newFakeVeleroAdapter()
	ctx := context.Background()

	opts := RestoreOptions{
		BackupName:      "backup-1",
		SourceNamespace: "production",
		TargetNamespace: "sandbox-1",
		LabelSelector: map[string]string{
			"app":  "myapp",
			"tier": "frontend",
		},
	}

	restoreID, err := adapter.TriggerRestore(ctx, opts)
	require.NoError(t, err)

	restore := &unstructured.Unstructured{}
	restore.SetGroupVersionKind(restoreGVK)
	err = adapter.client.Get(ctx, types.NamespacedName{Name: restoreID, Namespace: "velero"}, restore)
	require.NoError(t, err)

	matchLabels, _, _ := unstructured.NestedStringMap(restore.Object, "spec", "labelSelector", "matchLabels")
	assert.Equal(t, "myapp", matchLabels["app"])
	assert.Equal(t, "frontend", matchLabels["tier"])
}

func TestWaitForRestoreCompleted(t *testing.T) {
	// Pre-create a restore that's already completed
	adapter := newFakeVeleroAdapter(
		makeRestore("test-restore-1", "Completed"),
	)

	result, err := adapter.WaitForRestore(context.Background(), "test-restore-1", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, result.Success)
}

func TestWaitForRestoreFailed(t *testing.T) {
	adapter := newFakeVeleroAdapter(
		makeRestore("test-restore-fail", "Failed"),
	)

	result, err := adapter.WaitForRestore(context.Background(), "test-restore-fail", 10*time.Second)
	require.NoError(t, err)
	assert.False(t, result.Success)
}

func TestWaitForRestorePartiallyFailed(t *testing.T) {
	adapter := newFakeVeleroAdapter(
		makeRestore("test-restore-partial", "PartiallyFailed"),
	)

	result, err := adapter.WaitForRestore(context.Background(), "test-restore-partial", 10*time.Second)
	require.NoError(t, err)
	assert.True(t, result.Success) // PartiallyFailed is treated as partial success
}

func TestWaitForRestoreTimeout(t *testing.T) {
	// Create a restore stuck in InProgress
	adapter := newFakeVeleroAdapter(
		makeRestore("test-restore-stuck", "InProgress"),
	)

	_, err := adapter.WaitForRestore(context.Background(), "test-restore-stuck", 1*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "timeout")
}

func TestWaitForRestoreNotFound(t *testing.T) {
	adapter := newFakeVeleroAdapter()

	_, err := adapter.WaitForRestore(context.Background(), "nonexistent", 5*time.Second)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCleanupRestore(t *testing.T) {
	adapter := newFakeVeleroAdapter(
		makeRestore("cleanup-me", "Completed"),
	)
	ctx := context.Background()

	err := adapter.CleanupRestore(ctx, "cleanup-me")
	require.NoError(t, err)

	// Verify it's gone
	restore := &unstructured.Unstructured{}
	restore.SetGroupVersionKind(restoreGVK)
	err = adapter.client.Get(ctx, types.NamespacedName{Name: "cleanup-me", Namespace: "velero"}, restore)
	assert.True(t, errors.IsNotFound(err))
}

func TestCleanupRestoreNotFound(t *testing.T) {
	adapter := newFakeVeleroAdapter()

	// Should not error if the restore doesn't exist
	err := adapter.CleanupRestore(context.Background(), "nonexistent")
	assert.NoError(t, err)
}

func TestContainsOrWildcard(t *testing.T) {
	tests := []struct {
		name  string
		slice []string
		value string
		want  bool
	}{
		{"exact match", []string{"prod", "staging"}, "prod", true},
		{"no match", []string{"prod", "staging"}, "dev", false},
		{"wildcard", []string{"*"}, "anything", true},
		{"empty slice (all)", nil, "anything", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, containsOrWildcard(tt.slice, tt.value))
		})
	}
}
