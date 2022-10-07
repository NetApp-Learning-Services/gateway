/*
Copyright 2022.

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

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// StorageVirtualMachineSpec defines the desired state of StorageVirtualMachine
type StorageVirtualMachineSpec struct {

	// Provides required SVM name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=3
	// +kubebuilder:validation:MaxLength:=253
	/// +kubebuilder:validation:Format:=hostname
	SvmName string `json:"svmName"`

	// Provides required Cluster management LIF URL
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern="(http|ftp|https)://([a-zA-Z0-9~!@#$%^&*()_=+/?.:;',-]*)?"
	ClusterManagementUrl string `json:"clusterUrl"`

	// Stores SVM's uuid after it is created
	SvmUuid string `json:"uuid,omitempty"`

	// Provides optional SVM managment LIF
	// +kubebuilder:validation:Optional
	ManagementLIF ManagementLIF `json:"management,omitempty"`

	// Provides required ONTAP cluster administrator credentials
	// +kubebuilder:validation:Required
	ClusterCredentialSecret NamespacedName `json:"clusterCredentials"`

	// Provides optional SVM administrator credentials
	// +kubebuilder:validation:Optional
	VsadminCredentialSecret NamespacedName `json:"vsadminCredentials,omitempty"`
}

// StorageVirtualMachineStatus defines the observed state of StorageVirtualMachine
type StorageVirtualMachineStatus struct {

	// // LastUpdate records the last time an update was executed
	// LastUpdate metav1.Time `json:"lastUpdate,omitempty"`

	// // State shows the reconcile run
	// // +kubebuilder:validation:Enum=Provisioning;Provisioned;Failed;Unknown;Deleting
	// OperationState OperationState `json:"state,omitempty"`

	// // Message provides related status update
	// Message string `json:"message"`

	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=storagevirtualmachines,shortName=svm
// StorageVirtualMachine is the Schema for the storagevirtualmachines API
type StorageVirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   StorageVirtualMachineSpec   `json:"spec,omitempty"`
	Status StorageVirtualMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// StorageVirtualMachineList contains a list of StorageVirtualMachine
type StorageVirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []StorageVirtualMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&StorageVirtualMachine{}, &StorageVirtualMachineList{})
}

func (svm *StorageVirtualMachine) GetConditions() []metav1.Condition {
	return svm.Status.Conditions
}

func (svm *StorageVirtualMachine) SetConditions(conditions []metav1.Condition) {
	svm.Status.Conditions = conditions
}
