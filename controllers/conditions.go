// From: https://github.com/nheidloff/operator-sample-go/blob/bc4571d4d7431b60676919379ad3c3a2abcfd175/operator-application/controllers/application/conditions.go

package controllers

import (
	"context"

	gatewayv1alpha1 "gateway/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CONDITION_STATUS_TRUE = "True"
const CONDITION_STATUS_FALSE = "False"
const CONDITION_STATUS_UNKNOWN = "Unknown"

// Note: Status of RESOURCE_FOUND can only be True; otherwise there is no condition
const CONDITION_TYPE_RESOURCE_FOUND = "ResourceDiscovered"
const CONDITION_REASON_RESOURCE_FOUND = "ResourceFound"
const CONDITION_MESSAGE_RESOURCE_FOUND = "Resource found by gateway operator"

func (reconciler *StorageVirtualMachineReconciler) setConditionResourceFound(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine) error {

	if !reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_RESOURCE_FOUND) {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_RESOURCE_FOUND, CONDITION_STATUS_TRUE,
			CONDITION_REASON_RESOURCE_FOUND, CONDITION_MESSAGE_RESOURCE_FOUND)
	}
	return nil
}

// Note: Status of SVM_CREATED can only be True; otherwise there is no condition
const CONDITION_TYPE_SVM_CREATED = "ResourceCreatedSVM"
const CONDITION_REASON_SVM_CREATED = "SVMCreation"
const CONDITION_MESSAGE_SVM_CREATED_TRUE = "SVM creation succeeded"
const CONDITION_MESSAGE_SVM_CREATED_FALSE = "SVM creation failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionSVMCreation(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_SVM_CREATED) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_SVM_CREATED, CONDITION_REASON_SVM_CREATED)
	}

	if status == "True" {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_CREATED, status,
			CONDITION_REASON_SVM_CREATED, CONDITION_MESSAGE_SVM_CREATED_TRUE)
	}

	if status == "False" {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_CREATED, status,
			CONDITION_REASON_SVM_CREATED, CONDITION_MESSAGE_SVM_CREATED_FALSE)
	}
	return nil
}
