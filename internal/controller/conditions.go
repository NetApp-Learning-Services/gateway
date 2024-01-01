// From: https://github.com/nheidloff/operator-sample-go/blob/bc4571d4d7431b60676919379ad3c3a2abcfd175/operator-application/controllers/application/conditions.go

package controller

import (
	"context"

	gatewayv1alpha2 "gateway/api/v1alpha2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CONDITION_STATUS_TRUE = "True"
const CONDITION_STATUS_FALSE = "False"
const CONDITION_STATUS_UNKNOWN = "Unknown"

// STEP 1
// Report whether custom resource was located by operator
// Note: Status of RESOURCE_FOUND can only be true; otherwise there is no condition
const CONDITION_TYPE_RESOURCE_FOUND = "1ResourceDiscovered"
const CONDITION_REASON_RESOURCE_FOUND = "ResourceFound"
const CONDITION_MESSAGE_RESOURCE_FOUND = "Resource discovered"

func (reconciler *StorageVirtualMachineReconciler) setConditionResourceFound(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine) error {

	if !reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_RESOURCE_FOUND) {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_RESOURCE_FOUND, CONDITION_STATUS_TRUE,
			CONDITION_REASON_RESOURCE_FOUND, CONDITION_MESSAGE_RESOURCE_FOUND)
	}
	return nil
}

// STEP 2
// Host name discovery
// Note: Status of HOST_FOUND can only be true, false, unknown
const CONDITION_TYPE_HOST_FOUND = "2HostDiscovered"
const CONDITION_REASON_HOST_FOUND = "HostFound"
const CONDITION_MESSAGE_HOST_FOUND_TRUE = "A valid host found"
const CONDITION_MESSAGE_HOST_FOUND_FALSE = "A valid host was not found"

func (reconciler *StorageVirtualMachineReconciler) setConditionHostFound(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_HOST_FOUND) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_HOST_FOUND, CONDITION_REASON_HOST_FOUND)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_HOST_FOUND, status,
			CONDITION_REASON_HOST_FOUND, CONDITION_MESSAGE_HOST_FOUND_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_HOST_FOUND, status,
			CONDITION_REASON_HOST_FOUND, CONDITION_MESSAGE_HOST_FOUND_FALSE)
	}

	if status == CONDITION_STATUS_UNKNOWN {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_HOST_FOUND, status,
			CONDITION_REASON_HOST_FOUND, CONDITION_MESSAGE_HOST_FOUND_FALSE)
	}

	return nil
}

// STEP 3
// Resolve Secret
// Note: Status of CLUSTER_SECRET_LOOKUP can only be true or false
const CONDITION_TYPE_CLUSTER_SECRET_LOOKUP = "3ClusterAdminSecretLookup"
const CONDITION_REASON_CLUSTER_SECRET_LOOKUP = "ClusterAdminSecretLookup"
const CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_TRUE = "Cluster Admin credentials available"
const CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_FALSE = "Cluster Admin credentials NOT available"

func (reconciler *StorageVirtualMachineReconciler) setConditionClusterSecretLookup(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_CLUSTER_SECRET_LOOKUP) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_CLUSTER_SECRET_LOOKUP, CONDITION_REASON_CLUSTER_SECRET_LOOKUP)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_CLUSTER_SECRET_LOOKUP, status,
			CONDITION_REASON_CLUSTER_SECRET_LOOKUP, CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_CLUSTER_SECRET_LOOKUP, status,
			CONDITION_REASON_CLUSTER_SECRET_LOOKUP, CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_FALSE)
	}
	return nil
}

// STEP 4
// ONTAP client Creation
// Note: Status of ONTAP_CREATED can only be true or false
const CONDITION_TYPE_ONTAP_CREATED = "4CreatedONTAPClient"
const CONDITION_REASON_ONTAP_CREATED = "ONTAPClientCreation"
const CONDITION_MESSAGE_ONTAP_CREATED_TRUE = "ONTAP client created"
const CONDITION_MESSAGE_ONTAP_CREATED_FALSE = "ONTAP client failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionONTAPCreation(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_ONTAP_CREATED) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_ONTAP_CREATED, CONDITION_REASON_ONTAP_CREATED)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ONTAP_CREATED, status,
			CONDITION_REASON_ONTAP_CREATED, CONDITION_MESSAGE_ONTAP_CREATED_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ONTAP_CREATED, status,
			CONDITION_REASON_ONTAP_CREATED, CONDITION_MESSAGE_ONTAP_CREATED_FALSE)
	}
	return nil
}

// STEP 5
// SVM Deletion
// Note: Status of SVM_DELETION can only be false or unknown
// Never have a true state because the custom resource is deleted if true occurs
// and therefore can't update the condition status on the custom resource
const CONDITION_TYPE_SVM_DELETION = "5SVMDeletion"
const CONDITION_REASON_SVM_DELETION = "SVMDeleted"

// const CONDITION_MESSAGE_SVM_DELETION_TRUE = "SVM deleted"
const CONDITION_MESSAGE_SVM_DELETION_FALSE = "SVM NOT deleted - finalizer remains"
const CONDITION_MESSAGE_SVM_DELETION_UNKNOWN = "SVM deletion in unknown state"

func (reconciler *StorageVirtualMachineReconciler) setConditionSVMDeleted(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_SVM_DELETION) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_SVM_DELETION, CONDITION_REASON_SVM_DELETION)
	}

	// if status == CONDITION_STATUS_TRUE {
	// 	return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_DELETION, status,
	// 		CONDITION_REASON_SVM_DELETION, CONDITION_MESSAGE_SVM_DELETION_TRUE)
	// }

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_DELETION, status,
			CONDITION_REASON_SVM_DELETION, CONDITION_MESSAGE_SVM_DELETION_FALSE)
	}

	if status == CONDITION_STATUS_UNKNOWN {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_DELETION, status,
			CONDITION_REASON_SVM_DELETION, CONDITION_MESSAGE_SVM_DELETION_UNKNOWN)
	}
	return nil
}

// STEP 6
// SVM Lookup
// Note: Status of SVM_FOUND can only be true, false, or unknown
const CONDITION_TYPE_SVM_FOUND = "6SVMDiscovered"
const CONDITION_REASON_SVM_FOUND = "SVMFound"
const CONDITION_MESSAGE_SVM_FOUND_TRUE = "UUID maps to SVM"
const CONDITION_MESSAGE_SVM_FOUND_FALSE = "NO UUID"
const CONDITION_MESSAGE_SVM_FOUND_UNKNOWN = "UUID does NOT map to SVM"

func (reconciler *StorageVirtualMachineReconciler) setConditionSVMFound(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_SVM_FOUND) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_SVM_FOUND, CONDITION_REASON_SVM_FOUND)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_FOUND, status,
			CONDITION_REASON_SVM_FOUND, CONDITION_MESSAGE_SVM_FOUND_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_FOUND, status,
			CONDITION_REASON_SVM_FOUND, CONDITION_MESSAGE_SVM_FOUND_FALSE)
	}

	if status == CONDITION_STATUS_UNKNOWN {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_FOUND, status,
			CONDITION_REASON_SVM_FOUND, CONDITION_MESSAGE_SVM_FOUND_UNKNOWN)
	}
	return nil
}

// STEP 7
// SVM Creation
// Note: Status of SVM_CREATED can only be true or false
const CONDITION_TYPE_SVM_CREATED = "7CreatedSVM"
const CONDITION_REASON_SVM_CREATED = "SVMCreation"
const CONDITION_MESSAGE_SVM_CREATED_TRUE = "SVM creation succeeded"
const CONDITION_MESSAGE_SVM_CREATED_FALSE = "SVM creation failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionSVMCreation(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_SVM_CREATED) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_SVM_CREATED, CONDITION_REASON_SVM_CREATED)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_CREATED, status,
			CONDITION_REASON_SVM_CREATED, CONDITION_MESSAGE_SVM_CREATED_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_CREATED, status,
			CONDITION_REASON_SVM_CREATED, CONDITION_MESSAGE_SVM_CREATED_FALSE)
	}
	return nil
}

// STEP 8
// VSADMIN LOOKUP
// Note: Status of VSADMIN_SECRET_LOOKUP can only be true or false
const CONDITION_TYPE_VSADMIN_SECRET_LOOKUP = "8VsAdminSecretLookup"
const CONDITION_REASON_VSADMIN_SECRET_LOOKUP = "VsAdminSecretLookup"
const CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_TRUE = "SVM Admin credentials available"
const CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_FALSE = "SVM Admin credentials NOT available"

func (reconciler *StorageVirtualMachineReconciler) setConditionVsadminSecretLookup(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_VSADMIN_SECRET_LOOKUP) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_VSADMIN_SECRET_LOOKUP, CONDITION_REASON_VSADMIN_SECRET_LOOKUP)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_VSADMIN_SECRET_LOOKUP, status,
			CONDITION_REASON_VSADMIN_SECRET_LOOKUP, CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_VSADMIN_SECRET_LOOKUP, status,
			CONDITION_REASON_VSADMIN_SECRET_LOOKUP, CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_FALSE)
	}
	return nil
}

// STEP 9
// VSADMIN UPDATE
// Note: Status of VSADMIN_UPDATE can only be true or false
const CONDITION_TYPE_VSADMIN_SECRET_UPDATE = "9VsAdminSecretUpdate"
const CONDITION_REASON_VSADMIN_SECRET_UPDATE = "VsAdminSecretUpdate"
const CONDITION_MESSAGE_VSADMIN_SECRET_UPDATE_TRUE = "SVM Admin credentials updated in ONTAP"
const CONDITION_MESSAGE_VSADMIN_SECRET_UPDATE_FALSE = "SVM Admin credentials NOT updated in ONTAP"

func (reconciler *StorageVirtualMachineReconciler) setConditionVsadminSecretUpdate(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_VSADMIN_SECRET_UPDATE) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_VSADMIN_SECRET_UPDATE, CONDITION_REASON_VSADMIN_SECRET_UPDATE)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_VSADMIN_SECRET_UPDATE, status,
			CONDITION_REASON_VSADMIN_SECRET_UPDATE, CONDITION_MESSAGE_VSADMIN_SECRET_UPDATE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_VSADMIN_SECRET_UPDATE, status,
			CONDITION_REASON_VSADMIN_SECRET_UPDATE, CONDITION_MESSAGE_VSADMIN_SECRET_UPDATE_FALSE)
	}
	return nil
}

// STEP 10
// SVM Update
// Note: Status of SVM_UPDATED can only be true or false
const CONDITION_TYPE_SVM_UPDATED = "10UpdatedSVM"
const CONDITION_REASON_SVM_UPDATED = "SVMUpdate"
const CONDITION_MESSAGE_SVM_UPDATED_TRUE = "SVM update succeeded"
const CONDITION_MESSAGE_SVM_UPDATED_FALSE = "SVM update failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionSVMUpdate(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_SVM_UPDATED) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_SVM_UPDATED, CONDITION_REASON_SVM_UPDATED)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_UPDATED, status,
			CONDITION_REASON_SVM_UPDATED, CONDITION_MESSAGE_SVM_UPDATED_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_UPDATED, status,
			CONDITION_REASON_SVM_UPDATED, CONDITION_MESSAGE_SVM_UPDATED_FALSE)
	}
	return nil
}

// STEP 11
// Management LIF Upsert
// Note: Status of MANGEMENTLIF_UPSERT can only be true or false
const CONDITION_TYPE_MANGEMENTLIF_UPSERT = "11UpsertdManagementLIF"
const CONDITION_REASON_MANGEMENTLIF_UPDATED = "ManagementLIFUpdate"
const CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_TRUE = "Management LIF update succeeded"
const CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_FALSE = "Management LIF update failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionManagementLIFUpdate(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_MANGEMENTLIF_UPSERT) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPDATED, CONDITION_REASON_MANGEMENTLIF_UPDATED)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_UPDATED, CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_UPDATED, CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_FALSE)
	}
	return nil
}

const CONDITION_REASON_MANGEMENTLIF_CREATION = "ManagementLIFCreation"
const CONDITION_MESSAGE_MANGEMENTLIF_CREATION_TRUE = "Management LIF creation succeeded"
const CONDITION_MESSAGE_MANGEMENTLIF_CREATION_FALSE = "Management LIF creation failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionManagementLIFCreation(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_MANGEMENTLIF_CREATION) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, CONDITION_REASON_MANGEMENTLIF_CREATION)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_CREATION, CONDITION_MESSAGE_MANGEMENTLIF_CREATION_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_CREATION, CONDITION_MESSAGE_MANGEMENTLIF_CREATION_FALSE)
	}
	return nil
}

// STEP 12
// Aggregate assigned
// Note: Status of AGGREGATE_ASSIGNED can only be true or false
const CONDITION_TYPE_AGGREGATE_ASSIGNED = "12AggregateAssigned"
const CONDITION_REASON_AGGREGATE_ASSIGNED = "AggregateAssigned"
const CONDITION_MESSAGE_AGGREGATE_ASSIGNED_TRUE = "Aggregate assigned to SVM succeeded"
const CONDITION_MESSAGE_AGGREGATE_ASSIGNED_FALSE = "Aggregate assigned to SVM failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionAggregateAssigned(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_AGGREGATE_ASSIGNED) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_AGGREGATE_ASSIGNED, CONDITION_REASON_AGGREGATE_ASSIGNED)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_AGGREGATE_ASSIGNED, status,
			CONDITION_REASON_AGGREGATE_ASSIGNED, CONDITION_MESSAGE_AGGREGATE_ASSIGNED_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_AGGREGATE_ASSIGNED, status,
			CONDITION_REASON_AGGREGATE_ASSIGNED, CONDITION_MESSAGE_AGGREGATE_ASSIGNED_FALSE)
	}
	return nil
}

// STEP 13
// NFS update
// Note: Status of NFS_SERVICE can only be true or false
const CONDITION_TYPE_NFS_SERVICE = "13NFSservice"
const CONDITION_REASON_NFS_SERVICE = "NFSservice"
const CONDITION_MESSAGE_NFS_SERVICE_TRUE = "NFS service configuration succeeded"
const CONDITION_MESSAGE_NFS_SERVICE_FALSE = "NFS service configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNfsService(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NFS_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NFS_SERVICE, CONDITION_REASON_NFS_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_SERVICE, CONDITION_MESSAGE_NFS_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_SERVICE, CONDITION_MESSAGE_NFS_SERVICE_FALSE)
	}
	return nil
}

const CONDITION_REASON_NFS_LIF = "NFSlif"
const CONDITION_MESSAGE_NFS_LIF_TRUE = "NFS LIF configuration succeeded"
const CONDITION_MESSAGE_NFS_LIF_FALSE = "NFS LIF configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNfsLif(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NFS_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NFS_SERVICE, CONDITION_REASON_NFS_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_LIF, CONDITION_MESSAGE_NFS_LIF_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_LIF, CONDITION_MESSAGE_NFS_LIF_FALSE)
	}
	return nil
}

const CONDITION_REASON_NFS_EXPORT = "NFSexport"
const CONDITION_MESSAGE_NFS_EXPORT_TRUE = "NFS export configuration succeeded"
const CONDITION_MESSAGE_NFS_EXPORT_FALSE = "NFS export configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNfsExport(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NFS_EXPORT) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NFS_SERVICE, CONDITION_REASON_NFS_EXPORT)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_EXPORT, CONDITION_MESSAGE_NFS_EXPORT_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_EXPORT, CONDITION_MESSAGE_NFS_EXPORT_FALSE)
	}
	return nil
}

// STEP 14
// iSCSI update
// Note: Status of ISCSI_SERVICE can only be true or false
const CONDITION_TYPE_ISCSI_SERVICE = "14iSCSIservice"
const CONDITION_REASON_ISCSI_SERVICE = "iSCSIservice"
const CONDITION_MESSAGE_ISCSI_SERVICE_TRUE = "iSCSI service configuration succeeded"
const CONDITION_MESSAGE_ISCSI_SERVICE_FALSE = "iSCSI service configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionIscsiService(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_ISCSI_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_ISCSI_SERVICE, CONDITION_REASON_ISCSI_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ISCSI_SERVICE, status,
			CONDITION_REASON_ISCSI_SERVICE, CONDITION_MESSAGE_ISCSI_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ISCSI_SERVICE, status,
			CONDITION_REASON_ISCSI_SERVICE, CONDITION_MESSAGE_ISCSI_SERVICE_FALSE)
	}
	return nil
}

const CONDITION_REASON_ISCSI_LIF = "iSCSIlif"
const CONDITION_MESSAGE_ISCSI_LIF_TRUE = "iSCSI LIF configuration succeeded"
const CONDITION_MESSAGE_ISCSI_LIF_FALSE = "iSCSI LIF configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionIscsiLif(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_ISCSI_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_ISCSI_SERVICE, CONDITION_REASON_ISCSI_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ISCSI_SERVICE, status,
			CONDITION_REASON_ISCSI_LIF, CONDITION_MESSAGE_ISCSI_LIF_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ISCSI_SERVICE, status,
			CONDITION_REASON_ISCSI_LIF, CONDITION_MESSAGE_ISCSI_LIF_FALSE)
	}
	return nil
}
