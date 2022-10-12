package controllers

import (
	"context"
	"fmt"
	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const finalizer = "gateway.netapp.com" //special word

func (reconciler *StorageVirtualMachineReconciler) finalizeSVM(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client) error {

	err := oc.DeleteStorageVM(svmCR.Spec.SvmUuid)
	if err != nil {
		return fmt.Errorf("SVM not deleted yet")
	}
	return nil
}

func (reconciler *StorageVirtualMachineReconciler) addFinalizer(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(svmCR, finalizer) {
		controllerutil.AddFinalizer(svmCR, finalizer)
		err := reconciler.Update(ctx, svmCR)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (reconciler *StorageVirtualMachineReconciler) tryDeletions(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client) (ctrl.Result, error) {
	isSMVMarkedToBeDeleted := svmCR.GetDeletionTimestamp() != nil
	if isSMVMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(svmCR, finalizer) {
			if err := reconciler.finalizeSVM(ctx, svmCR, oc); err != nil {
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(svmCR, finalizer)
			err := reconciler.Update(ctx, svmCR)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}
