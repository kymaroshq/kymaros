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

// HealthCheckPolicySpec defines the desired state of HealthCheckPolicy
type HealthCheckPolicySpec struct {
	Checks []HealthCheck `json:"checks"`
}

// HealthCheck defines a single health check
type HealthCheck struct {
	// Name of this check (unique within policy)
	Name string `json:"name"`

	// Type: podStatus | httpGet | exec | tcpSocket | resourceExists
	// +kubebuilder:validation:Enum=podStatus;httpGet;exec;tcpSocket;resourceExists
	Type string `json:"type"`

	// Type-specific config (only one should be set)

	// +optional
	PodStatus *PodStatusCheck `json:"podStatus,omitempty"`
	// +optional
	HTTPGet *HTTPGetCheck `json:"httpGet,omitempty"`
	// +optional
	Exec *ExecCheck `json:"exec,omitempty"`
	// +optional
	TCPSocket *TCPSocketCheck `json:"tcpSocket,omitempty"`
	// +optional
	ResourceExists *ResourceExistsCheck `json:"resourceExists,omitempty"`
}

// PodStatusCheck verifies pod readiness
type PodStatusCheck struct {
	LabelSelector map[string]string `json:"labelSelector"`
	MinReady      int               `json:"minReady"`
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

// HTTPGetCheck performs an HTTP GET health check
type HTTPGetCheck struct {
	Service        string `json:"service"`
	Port           int    `json:"port"`
	Path           string `json:"path"`
	ExpectedStatus int    `json:"expectedStatus"`
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`
	// +optional
	Retries int `json:"retries,omitempty"`
}

// ExecCheck runs a command in a pod
type ExecCheck struct {
	PodSelector map[string]string `json:"podSelector"`
	// +optional
	Container string   `json:"container,omitempty"`
	Command   []string `json:"command"`
	// +kubebuilder:default=0
	SuccessExitCode int `json:"successExitCode"`
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

// TCPSocketCheck verifies TCP connectivity
type TCPSocketCheck struct {
	Service string `json:"service"`
	Port    int    `json:"port"`
	// +optional
	Timeout metav1.Duration `json:"timeout,omitempty"`
}

// ResourceExistsCheck verifies that K8s resources exist
type ResourceExistsCheck struct {
	Resources []ResourceRef `json:"resources"`
}

// ResourceRef identifies a K8s resource
type ResourceRef struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// HealthCheckPolicyStatus defines the observed state of HealthCheckPolicy.
type HealthCheckPolicyStatus struct{}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=hcp

// HealthCheckPolicy is the Schema for the healthcheckpolicies API
type HealthCheckPolicy struct {
	metav1.TypeMeta `json:",inline"`

	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// +required
	Spec HealthCheckPolicySpec `json:"spec"`

	// +optional
	Status HealthCheckPolicyStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// HealthCheckPolicyList contains a list of HealthCheckPolicy
type HealthCheckPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []HealthCheckPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HealthCheckPolicy{}, &HealthCheckPolicyList{})
}
