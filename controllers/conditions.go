package controllers

import (
	"context"

	gatewayv1alpha1 "github.com/NetApp-Learning-Services/gateway/api/v1alpha1"
)

const CONDITION_STATUS_TRUE = "True"
const CONDITION_STATUS_FALSE = "False"
const CONDITION_STATUS_UNKNOWN = "Unknown"

// Note: Status of RESOURCE_FOUND can only be True; otherwise there is no condition
const CONDITION_TYPE_RESOURCE_FOUND = "ResourceFound"
const CONDITION_REASON_RESOURCE_FOUND = "ResourceFound"
const CONDITION_MESSAGE_RESOURCE_FOUND = "Resource found in k18n"

func (reconciler *StorageVirtualMachineReconciler) setConditionResourceFound(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine) error {

	if !reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_RESOURCE_FOUND) {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_RESOURCE_FOUND, CONDITION_STATUS_TRUE,
			CONDITION_REASON_RESOURCE_FOUND, CONDITION_MESSAGE_RESOURCE_FOUND)
	}
	return nil
}
