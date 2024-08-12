package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DeploymentSyncSpec defines the desired state of DeploymentSync
type DeploymentSyncSpec struct {
    SourceNamespace      string `json:"sourceNamespace"`
    DestinationNamespace string `json:"destinationNamespace"`
    DeploymentName        string `json:"deploymentName"`
}
// DeploymentSyncStatus defines the observed state of DeploymentSync
type DeploymentSyncStatus struct {
    LastSyncTime metav1.Time `json:"lastSyncTime"`
}
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// DeploymentSync is the Schema for the deploymentsyncs API
type DeploymentSync struct {
    metav1.TypeMeta   `json:",inline"`
    metav1.ObjectMeta `json:"metadata,omitempty"`
    Spec   DeploymentSyncSpec   `json:"spec,omitempty"`
    Status DeploymentSyncStatus `json:"status,omitempty"`
}
// +kubebuilder:object:root=true
// DeploymentSyncList contains a list of DeploymentSync
type DeploymentSyncList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []DeploymentSync `json:"items"`
}
func init() {
    SchemeBuilder.Register(&DeploymentSync{}, &DeploymentSyncList{})
}