// From: https://github.com/nheidloff/operator-sample-go/blob/bc4571d4d7431b60676919379ad3c3a2abcfd175/operator-application/controllers/application/deletions.go

package controllers

import (
	"context"
	"fmt"
	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const finalizerName = "gateway.netapp.com" //magic word

func (r *StorageVirtualMachineReconciler) reconcileDeletions(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client, log logr.Logger) (ctrl.Result, error) {

	log.Info("STEP 5: Delete SVM in ONTAP and remove custom resource")

	isSMVMarkedToBeDeleted := svmCR.GetDeletionTimestamp() != nil
	if isSMVMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(svmCR, finalizerName) {
			if err := r.finalizeSVM(ctx, svmCR, oc); err != nil {
				log.Error(err, "Error during custom resource deletion - requeuing")
				_ = r.setConditionSVMDeleted(ctx, svmCR, CONDITION_STATUS_FALSE)
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(svmCR, finalizerName)
			err := r.Update(ctx, svmCR)
			if err != nil {
				log.Error(err, "Error during custom resource deletion - requeuing")
				_ = r.setConditionSVMDeleted(ctx, svmCR, CONDITION_STATUS_UNKNOWN)
				return ctrl.Result{}, err
			}
		}
		// Can't do this because custom resource is deleted
		//_ = r.setConditionSVMDeleted(ctx, svmCR, CONDITION_STATUS_TRUE)
		log.Info("SVM deleted, removed finalizer, cleaning up custom resource")
		return ctrl.Result{}, nil
	}
	// Not deleted
	return ctrl.Result{}, nil
}

func (r *StorageVirtualMachineReconciler) finalizeSVM(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client) error {

	err := oc.DeleteStorageVM(svmCR.Spec.SvmUuid)
	if err != nil {
		return fmt.Errorf("SVM not deleted yet")
	}
	return nil
}

// Moved to 7svmcreation.go because that is where it is executed
// func (r *StorageVirtualMachineReconciler) addFinalizer(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine) (ctrl.Result, error) {
// 	if !controllerutil.ContainsFinalizer(svmCR, finalizer) {
// 		controllerutil.AddFinalizer(svmCR, finalizer)
// 		err := r.Update(ctx, svmCR)
// 		if err != nil {
// 			return ctrl.Result{}, err
// 		}
// 	}
// 	return ctrl.Result{}, nil
// }
