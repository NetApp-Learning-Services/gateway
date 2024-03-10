package controller

import (
	"context"
	gateway "gateway/api/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *StorageVirtualMachineReconciler) reconcileDiscoverObject(ctx context.Context,
	req ctrl.Request, log logr.Logger) (*gateway.StorageVirtualMachine, error) {

	svmCR := &gateway.StorageVirtualMachine{}
	log.Info("STEP 1: Discover Custom Resource")

	err := r.Get(ctx, req.NamespacedName, svmCR)
	if err != nil && errors.IsNotFound(err) {
		log.Info("StorageVirtualMachine custom resource not found, ignoring since object must be deleted - stopping reconcile")
		return nil, err
	} else if err != nil {
		log.Error(err, "Failed to get StorageVirtualMachine custom resource, re-running reconcile")
		return nil, err
	}

	//Set condition for CR found
	err = r.setConditionResourceFound(ctx, svmCR)
	if err != nil {
		return svmCR, err
	}

	return svmCR, nil

}

// STEP 1
// Report whether custom resource was located by operator
// Note: Status of RESOURCE_FOUND can only be true; otherwise there is no condition
const CONDITION_TYPE_RESOURCE_FOUND = "1ResourceDiscovered"
const CONDITION_REASON_RESOURCE_FOUND = "ResourceFound"
const CONDITION_MESSAGE_RESOURCE_FOUND = "Resource discovered"

func (reconciler *StorageVirtualMachineReconciler) setConditionResourceFound(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine) error {

	if !reconciler.containsCondition(svmCR, CONDITION_REASON_RESOURCE_FOUND) {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_RESOURCE_FOUND, CONDITION_STATUS_TRUE,
			CONDITION_REASON_RESOURCE_FOUND, CONDITION_MESSAGE_RESOURCE_FOUND)
	}
	return nil
}
