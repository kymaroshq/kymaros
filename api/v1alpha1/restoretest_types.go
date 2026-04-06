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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Phase constants for RestoreTest
const (
	PhaseIdle      = "Idle"
	PhaseRunning   = "Running"
	PhaseCompleted = "Completed"
	PhaseFailed    = "Failed"
)

// Result constants
const (
	ResultPass    = "pass"
	ResultFail    = "fail"
	ResultPartial = "partial"
)

// Label and annotation constants
const (
	LabelTestName  = "kymaros.io/test"
	LabelTestGroup = "kymaros.io/group"
)

// BackupSource defines which backup to test
type BackupSource struct {
	// Provider: velero
	// +kubebuilder:validation:Enum=velero
	Provider string `json:"provider"`

	// BackupName: specific name or "latest"
	BackupName string `json:"backupName"`

	// Namespaces to restore (supports multi-namespace restore groups)
	Namespaces []NamespaceMapping `json:"namespaces"`

	// LabelSelector to filter resources (optional)
	// +optional
	LabelSelector map[string]string `json:"labelSelector,omitempty"`
}

// NamespaceMapping maps a source namespace to an optional sandbox name
type NamespaceMapping struct {
	// Name of the source namespace
	Name string `json:"name"`
	// SandboxName override for the sandbox namespace (optional)
	// +optional
	SandboxName string `json:"sandboxName,omitempty"`
}

// ScheduleConfig defines when to run tests
type ScheduleConfig struct {
	// Cron expression
	Cron string `json:"cron"`
	// Timezone (default: UTC)
	// +optional
	Timezone string `json:"timezone,omitempty"`
}

// SandboxConfig defines the isolated namespace settings
type SandboxConfig struct {
	// NamespacePrefix for sandbox namespaces (default: "rp-test")
	// +kubebuilder:default="rp-test"
	// +optional
	NamespacePrefix string `json:"namespacePrefix,omitempty"`
	// TTL before forced cleanup (default: 30m)
	// +optional
	TTL metav1.Duration `json:"ttl,omitempty"`
	// ResourceQuota limits
	// +optional
	ResourceQuota *ResourceQuotaConfig `json:"resourceQuota,omitempty"`
	// NetworkIsolation mode: "strict" (deny-all) | "group" (allow intra-group)
	// +kubebuilder:validation:Enum=strict;group
	// +kubebuilder:default="strict"
	// +optional
	NetworkIsolation string `json:"networkIsolation,omitempty"`
}

// ResourceQuotaConfig defines resource limits for the sandbox
type ResourceQuotaConfig struct {
	// +optional
	CPU string `json:"cpu,omitempty"`
	// +optional
	Memory string `json:"memory,omitempty"`
	// +optional
	Storage string `json:"storage,omitempty"`
}

// HealthCheckRef references a HealthCheckPolicy
type HealthCheckRef struct {
	// PolicyRef references a HealthCheckPolicy by name
	// +optional
	PolicyRef string `json:"policyRef,omitempty"`
	// Timeout for all checks combined
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

// SLAConfig defines SLA targets
type SLAConfig struct {
	// MaxRTO is the declared RTO target
	MaxRTO metav1.Duration `json:"maxRTO"`
	// AlertOnExceed triggers alert if measured RTO > MaxRTO
	// +optional
	AlertOnExceed bool `json:"alertOnExceed,omitempty"`
}

// NotificationConfig defines notification channels
type NotificationConfig struct {
	// +optional
	OnFailure []NotificationChannel `json:"onFailure,omitempty"`
	// +optional
	OnSuccess []NotificationChannel `json:"onSuccess,omitempty"`
}

// NotificationChannel defines a single notification target
type NotificationChannel struct {
	// Type: slack | webhook
	// +kubebuilder:validation:Enum=slack;webhook
	Type string `json:"type"`
	// Channel (for Slack)
	// +optional
	Channel string `json:"channel,omitempty"`
	// WebhookSecretRef references a Secret containing the webhook URL
	// +optional
	WebhookSecretRef string `json:"webhookSecretRef,omitempty"`
}

// LogCollectionSpec configures pod log collection in the sandbox
type LogCollectionSpec struct {
	// Enabled controls whether logs are collected (default: true)
	// +optional
	// +kubebuilder:default=true
	Enabled bool `json:"enabled,omitempty"`

	// MaxLines is the maximum number of log lines per container (default: 100)
	// +optional
	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=1000
	MaxLines int `json:"maxLines,omitempty"`

	// IncludeEvents controls whether K8s events are collected (default: true)
	// +optional
	// +kubebuilder:default=true
	IncludeEvents bool `json:"includeEvents,omitempty"`

	// IncludeInit controls whether init container logs are collected (default: true)
	// +optional
	// +kubebuilder:default=true
	IncludeInit bool `json:"includeInit,omitempty"`

	// IncludePrevious controls whether previous (crashed) container logs are collected (default: true)
	// +optional
	// +kubebuilder:default=true
	IncludePrevious bool `json:"includePrevious,omitempty"`
}

// RestoreTestSpec defines the desired state of RestoreTest
type RestoreTestSpec struct {
	// BackupSource defines which backup to test
	BackupSource BackupSource `json:"backupSource"`

	// Schedule defines when to run (cron format)
	Schedule ScheduleConfig `json:"schedule"`

	// Sandbox configuration
	Sandbox SandboxConfig `json:"sandbox"`

	// HealthChecks references a HealthCheckPolicy
	// +optional
	HealthChecks HealthCheckRef `json:"healthChecks,omitempty"`

	// SLA targets
	// +optional
	SLA *SLAConfig `json:"sla,omitempty"`

	// Notifications channels
	// +optional
	Notifications *NotificationConfig `json:"notifications,omitempty"`

	// Timeout is the global timeout for the entire test run (sandbox creation through health checks).
	// If the timeout expires, the test is marked as Failed.
	// If not set, no global timeout is applied.
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`

	// HistoryLimit is the maximum number of RestoreReports to keep per test.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=10
	// +optional
	HistoryLimit *int32 `json:"historyLimit,omitempty"`

	// LogCollection configures pod log collection in the sandbox
	// +optional
	LogCollection *LogCollectionSpec `json:"logCollection,omitempty"`
}

// RestoreTestStatus defines the observed state of RestoreTest.
type RestoreTestStatus struct {
	// Phase: Idle | Running | Completed | Failed
	// +kubebuilder:validation:Enum=Idle;Running;Completed;Failed
	// +optional
	Phase string `json:"phase,omitempty"`

	// LastRunAt timestamp of last execution
	// +optional
	LastRunAt *metav1.Time `json:"lastRunAt,omitempty"`

	// LastScore from the most recent RestoreReport
	// +optional
	LastScore int `json:"lastScore,omitempty"`

	// LastResult: pass | fail | partial
	// +optional
	LastResult string `json:"lastResult,omitempty"`

	// LastReportRef name of the most recent RestoreReport
	// +optional
	LastReportRef string `json:"lastReportRef,omitempty"`

	// NextRunAt calculated next execution time
	// +optional
	NextRunAt *metav1.Time `json:"nextRunAt,omitempty"`

	// SandboxNamespace is the current sandbox namespace (first, for backward compat)
	// +optional
	SandboxNamespace string `json:"sandboxNamespace,omitempty"`

	// RestoreID is the current Velero Restore CR name (first, for backward compat)
	// +optional
	RestoreID string `json:"restoreID,omitempty"`

	// SandboxNamespaces tracks all sandbox namespaces for multi-namespace restores
	// +optional
	SandboxNamespaces []string `json:"sandboxNamespaces,omitempty"`

	// RestoreIDs tracks all Velero Restore CR names for multi-namespace restores
	// +optional
	RestoreIDs []string `json:"restoreIDs,omitempty"`

	// Conditions standard K8s conditions
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rt
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Score",type=integer,JSONPath=`.status.lastScore`
// +kubebuilder:printcolumn:name="Result",type=string,JSONPath=`.status.lastResult`
// +kubebuilder:printcolumn:name="Last Run",type=date,JSONPath=`.status.lastRunAt`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RestoreTest is the Schema for the restoretests API
type RestoreTest struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// +required
	Spec RestoreTestSpec `json:"spec"`

	// +optional
	Status RestoreTestStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// RestoreTestList contains a list of RestoreTest
type RestoreTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []RestoreTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RestoreTest{}, &RestoreTestList{})
}
