package adapter

import (
	"context"
	"time"
)

// BackupAdapter is the interface for interacting with backup providers
type BackupAdapter interface {
	ListBackups(ctx context.Context, namespace string) ([]BackupInfo, error)
	GetLatestBackup(ctx context.Context, namespace string) (*BackupInfo, error)
	TriggerRestore(ctx context.Context, opts RestoreOptions) (string, error)
	WaitForRestore(ctx context.Context, restoreID string, timeout time.Duration) (*RestoreResult, error)
	CleanupRestore(ctx context.Context, restoreID string) error
}

// BackupInfo contains metadata about a backup
type BackupInfo struct {
	Name      string
	Namespace string
	CreatedAt time.Time
	Size      int64
	Status    string // completed | failed | partial
}

// RestoreOptions defines parameters for a restore operation
type RestoreOptions struct {
	BackupName      string
	SourceNamespace string
	TargetNamespace string
	LabelSelector   map[string]string
	// VeleroNamespace is the namespace where Velero is installed (default: "velero")
	VeleroNamespace string
}

// ScheduleInfo contains metadata about a Velero backup schedule
type ScheduleInfo struct {
	Name       string
	Cron       string
	Namespaces []string
	LastBackup string
}

// RestoreResult captures the outcome of a restore operation
type RestoreResult struct {
	Success  bool
	Duration time.Duration
	Warnings []string
	Errors   []string
}
