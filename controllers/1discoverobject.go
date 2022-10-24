package controllers

import (
	"context"
	gatewayv1alpha1 "gateway/api/v1alpha1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *StorageVirtualMachineReconciler) reconcileDiscoverObject(ctx context.Context,
	req ctrl.Request, log logr.Logger) (*gatewayv1alpha1.StorageVirtualMachine, error) {

	svmCR := &gatewayv1alpha1.StorageVirtualMachine{}
	log.Info("Start Discover Custom Resource")

	err := r.Get(ctx, req.NamespacedName, svmCR)
	if err != nil && errors.IsNotFound(err) {
		log.Info("StorageVirtualMachine custom resource not found, ignoring since object must be deleted")
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