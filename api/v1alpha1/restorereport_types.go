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

// RestoreReportSpec defines the desired state of RestoreReport
type RestoreReportSpec struct {
	// TestRef name of the RestoreTest that generated this report
	TestRef string `json:"testRef"`
}

// RestoreReportStatus defines the observed state of RestoreReport.
type RestoreReportStatus struct {
	// Score 0-100 confidence score
	Score int `json:"score"`

	// Result: pass | fail | partial
	Result string `json:"result"`

	// Timing
	StartedAt   metav1.Time `json:"startedAt"`
	CompletedAt metav1.Time `json:"completedAt"`

	// RTO measurement
	RTO RTOStatus `json:"rto"`

	// Backup info
	Backup BackupStatus `json:"backup"`

	// Individual check results
	// +optional
	Checks []CheckResult `json:"checks,omitempty"`

	// Resource completeness
	Completeness CompletenessStatus `json:"completeness"`

	// Validation levels (layered validation)
	ValidationLevels ValidationLevels `json:"validationLevels"`

	// PodLogs contains logs from pods in the sandbox namespace
	// +optional
	PodLogs []PodLog `json:"podLogs,omitempty"`

	// Events contains Kubernetes events from the sandbox namespace
	// +optional
	Events []EventLog `json:"events,omitempty"`
}

// PodLog captures logs for a single pod in the sandbox
type PodLog struct {
	// PodName is the name of the pod
	PodName string `json:"podName"`
	// Namespace is the sandbox namespace
	Namespace string `json:"namespace"`
	// Phase is the pod phase (Running, Failed, CrashLoopBackOff, etc.)
	Phase string `json:"phase"`
	// Containers holds logs per container
	Containers []ContainerLog `json:"containers"`
}

// ContainerLog captures log output for a single container
type ContainerLog struct {
	// Name of the container
	Name string `json:"name"`
	// Type: "container", "init", "previous"
	Type string `json:"type"`
	// Log contains the last N lines of log output (truncated to maxLogLines)
	Log string `json:"log"`
	// Truncated indicates whether the log was truncated
	Truncated bool `json:"truncated"`
	// TotalLines is the number of lines before truncation
	// +optional
	TotalLines int `json:"totalLines,omitempty"`
}

// EventLog captures a single Kubernetes event
type EventLog struct {
	// Type: Normal, Warning
	Type string `json:"type"`
	// Reason: Failed, FailedScheduling, BackOff, etc.
	Reason string `json:"reason"`
	// Message is the event message
	Message string `json:"message"`
	// InvolvedObject is the object reference (e.g. Pod/my-pod)
	InvolvedObject string `json:"involvedObject"`
	// LastTimestamp is the last time the event occurred
	LastTimestamp string `json:"lastTimestamp"`
	// Count is the number of occurrences
	// +optional
	Count int `json:"count,omitempty"`
}

// RTOStatus captures RTO measurement
type RTOStatus struct {
	Measured metav1.Duration `json:"measured"`
	// +optional
	Target    metav1.Duration `json:"target,omitempty"`
	WithinSLA bool            `json:"withinSLA"`
}

// BackupStatus captures backup metadata
type BackupStatus struct {
	Name string `json:"name"`
	Age  string `json:"age"`
	// +optional
	Size string `json:"size,omitempty"`
}

// CheckResult captures a single health check result
type CheckResult struct {
	Name string `json:"name"`
	// Status: pass | fail | skip
	Status string `json:"status"`
	// +optional
	Duration string `json:"duration,omitempty"`
	// +optional
	Message string `json:"message,omitempty"`
}

// CompletenessStatus captures resource completeness
type CompletenessStatus struct {
	// +optional
	Deployments string `json:"deployments,omitempty"`
	// +optional
	Services string `json:"services,omitempty"`
	// +optional
	Secrets string `json:"secrets,omitempty"`
	// +optional
	ConfigMaps string `json:"configMaps,omitempty"`
	// +optional
	PVCs string `json:"pvcs,omitempty"`
	// +optional
	CustomResources string `json:"customResources,omitempty"`
}

// ValidationLevels captures layered validation results
type ValidationLevels struct {
	RestoreIntegrity   LevelResult `json:"restoreIntegrity"`
	Completeness       LevelResult `json:"completeness"`
	PodStartup         LevelResult `json:"podStartup"`
	InternalHealth     LevelResult `json:"internalHealth"`
	CrossNamespaceDeps LevelResult `json:"crossNamespaceDeps"`
	RTOCompliance      LevelResult `json:"rtoCompliance"`
}

// LevelResult captures a single validation level result
type LevelResult struct {
	// Status: pass | fail | partial | not_tested
	Status string `json:"status"`
	// +optional
	Detail string `json:"detail,omitempty"`
	// +optional
	Tested []string `json:"tested,omitempty"`
	// +optional
	NotTested []string `json:"notTested,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=rr
// +kubebuilder:printcolumn:name="Score",type=integer,JSONPath=`.status.score`
// +kubebuilder:printcolumn:name="Result",type=string,JSONPath=`.status.result`
// +kubebuilder:printcolumn:name="Test",type=string,JSONPath=`.spec.testRef`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// RestoreReport is the Schema for the restorereports API
type RestoreReport struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// +required
	Spec RestoreReportSpec `json:"spec"`

	// +optional
	Status RestoreReportStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// RestoreReportList contains a list of RestoreReport
type RestoreReportList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []RestoreReport `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RestoreReport{}, &RestoreReportList{})
}
