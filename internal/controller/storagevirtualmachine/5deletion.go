// From: https://github.com/nheidloff/operator-sample-go/blob/bc4571d4d7431b60676919379ad3c3a2abcfd175/operator-application/controllers/application/deletions.go

package controller

import (
	"context"
	"fmt"
	gateway "gateway/api/v1beta2"
	"gateway/internal/controller/ontap"
	defaultLog "log"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const finalizerName = "gateway.netapp.com/finalizer" //magic word

func (r *StorageVirtualMachineReconciler) reconcileDeletions(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, oc *ontap.Client, log logr.Logger) (ctrl.Result, error) {

	log.Info("STEP 5: Delete SVM in ONTAP and remove custom resource")
	var currentDeletionPolicy = svmCR.Spec.SvmDeletionPolicy
	if oc.Debug {
		defaultLog.Printf("%s", "[DEBUG] current deletion policy: "+fmt.Sprintf("%v", currentDeletionPolicy))
	}

	isSMVMarkedToBeDeleted := svmCR.GetDeletionTimestamp() != nil
	if isSMVMarkedToBeDeleted {
		if controllerutil.ContainsFinalizer(svmCR, finalizerName) {
			if err := r.finalizeSVM(svmCR, oc, log); err != nil {
				log.Error(err, "Error during deletionpolicy implementation - requeuing")
				_ = r.setConditionSVMDeleted(ctx, svmCR, CONDITION_STATUS_FALSE)
				return ctrl.Result{}, err
			}

			controllerutil.RemoveFinalizer(svmCR, finalizerName)
			err := r.Update(ctx, svmCR)
			if err != nil {
				log.Error(err, "Error during removal of finalizer - requeuing")
				_ = r.setConditionSVMDeleted(ctx, svmCR, CONDITION_STATUS_UNKNOWN)
				return ctrl.Result{}, err
			}
		}
		// Can't do this because custom resource is deleted
		//_ = r.setConditionSVMDeleted(ctx, svmCR, CONDITION_STATUS_TRUE)
		if currentDeletionPolicy == gateway.DeletionPolicyDelete {
			log.Info("SVM deleted, removed finalizer, cleaning up custom resource")
		} else {
			// default policy or currentDeletionPolicy == gateway.DeletionPolicyRetain
			log.Info("SVM retained, removed finalizer, cleaning up custom resource")
		}

		return ctrl.Result{}, nil
	}
	// CR is not deleted
	return ctrl.Result{}, nil
}

func (r *StorageVirtualMachineReconciler) finalizeSVM(
	svmCR *gateway.StorageVirtualMachine, oc *ontap.Client, log logr.Logger) error {

	if svmCR.Spec.SvmDeletionPolicy == gateway.DeletionPolicyDelete {

		//check to see if S3 buckets are presenty
		bucketsRetrieved, err := oc.GetS3BucketsBySvmUuid(svmCR.Spec.SvmUuid)

		if err != nil {
			log.Error(err, "Error retrieving S3 buckets list from SVM: "+svmCR.Spec.SvmName)
		}

		if bucketsRetrieved.NumRecords != 0 {
			for i := 0; i < bucketsRetrieved.NumRecords; i++ {
				err = oc.DeleteS3Bucket(svmCR.Spec.SvmUuid, bucketsRetrieved.Records[i].Uuid)

				if err != nil {
					log.Error(err, "Error deleting S3 bucket: "+bucketsRetrieved.Records[i].Name)
				}
			}

		}

		err = oc.DeleteStorageVM(svmCR.Spec.SvmUuid)
		if err != nil {
			return fmt.Errorf("SVM not deleted yet")
		}
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
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(svmCR, CONDITION_REASON_SVM_DELETION) {
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
