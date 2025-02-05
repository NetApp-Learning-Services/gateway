/*
Copyright 2023.

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

package v1beta2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DeletionPolicy string

const (
	DeletionPolicyRetain DeletionPolicy = "Retain"
	DeletionPolicyDelete DeletionPolicy = "Delete"
)

// StorageVirtualMachineSpec defines the desired state of StorageVirtualMachine
type StorageVirtualMachineSpec struct {
	// Provides required SVM name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=3
	// +kubebuilder:validation:MaxLength:=253
	/// +kubebuilder:validation:Format:=hostname
	SvmName string `json:"svmName"`

	// Provides required Cluster management LIF host IP address or host name
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`((^\s*((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))\s*$)|(^\s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}))|:)))(%.+)?\s*$))`
	ClusterManagementHost string `json:"clusterHost"`

	// Stores SVM's uuid after it is created
	SvmUuid string `json:"svmUuid,omitempty"`

	// Stores optional SVM's comment
	// +kubebuilder:validation:Optional
	SvmComment string `json:"svmComment,omitempty"`

	// Stores optional SVM's deletion policy
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Enum="Delete";"Retain"
	// +kubebuilder:default:=Retain
	SvmDeletionPolicy DeletionPolicy `json:"svmDeletionPolicy,omitempty"`

	// Stores optional debug
	// +kubebuilder:default:=false
	SvmDebug bool `json:"debug,omitempty"`

	// Stores optional aggregate list
	// +kubebuilder:validation:Optional
	Aggregates []Aggregate `json:"aggregates,omitempty"`

	// Provides optional SVM managment LIF
	// +kubebuilder:validation:Optional
	ManagementLIF *LIF `json:"management,omitempty"`

	// Provides required ONTAP cluster administrator credentials
	// +kubebuilder:validation:Required
	ClusterCredentialSecret NamespacedName `json:"clusterCredentials"`

	// Provides optional SVM administrator credentials
	// +kubebuilder:validation:Optional
	VsadminCredentialSecret NamespacedName `json:"vsadminCredentials,omitempty"`

	// Provide optional NFS configuration
	// +kubebuilder:validation:Optional
	NfsConfig *NfsSubSpec `json:"nfs,omitempty"`

	// Provide optional iSCSI configuration
	// +kubebuilder:validation:Optional
	IscsiConfig *IscsiSubSpec `json:"iscsi,omitempty"`

	// Provide optional NVMe configuration
	// +kubebuilder:validation:Optional
	NvmeConfig *NvmeSubSpec `json:"nvme,omitempty"`

	// Provide optional S3 configuration
	// +kubebuilder:validation:Optional
	S3Config *S3SubSpec `json:"s3,omitempty"`

	// Provide optional SVM peering configuration
	// +kubebuilder:validation:Optional
	PeerConfig *PeerSubSpec `json:"peer,omitempty"`
}

// StorageVirtualMachineStatus defines the observed state of StorageVirtualMachine
type StorageVirtualMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Conditions []metav1.Condition `json:"conditions"`
}

// CHECK OUT THIS:  https://www.brendanp.com/pretty-printing-with-kubebuilder/
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="SVM UUID",type="string",JSONPath=`.spec.svmUuid`
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

//+kubebuilder:object:root=true

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
