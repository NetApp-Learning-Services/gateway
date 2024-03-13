// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	defaultLog "log"

	gateway "gateway/api/v1beta1"
	"gateway/internal/controller/ontap"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmUpdate(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, svmRetrieved ontap.SvmByUUID, oc *ontap.Client, log logr.Logger) error {

	log.Info("STEP 10: Update SVM")

	execute := false
	var patchSVM ontap.SvmPatch

	// interate over custom resoource svmCR and look for differences in retrieved SVM
	if svmCR.Spec.SvmName != "" && svmCR.Spec.SvmName != svmRetrieved.Name {
		//update name
		patchSVM.Name = svmCR.Spec.SvmName
		execute = true
	}

	if svmCR.Spec.SvmComment != "" && svmCR.Spec.SvmComment != svmRetrieved.Comment {
		//update comment
		patchSVM.Comment = svmCR.Spec.SvmComment
		execute = true
	}

	if svmCR.Spec.NvmeConfig != nil && svmCR.Spec.NvmeConfig.Enabled != svmRetrieved.Nvme.Enabled {
		patchSVM.Nvme.Enabled = svmCR.Spec.NvmeConfig.Enabled
		execute = true
	}

	if !execute {
		log.Info("No changes for SVM - skipping STEP 10")
		return nil
	}
	if oc.Debug {
		defaultLog.Printf("[DEBUG] SVM update payload: " + fmt.Sprintf("%#v\n", patchSVM))
	}

	jsonPayload, err := json.Marshal(patchSVM)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for SVM update - requeuing")
		_ = r.setConditionSVMUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}

	// After building update string execute it and check for errors
	log.Info("SVM update attempt for SVM: " + svmRetrieved.Uuid)
	err = oc.PatchStorageVM(svmRetrieved.Uuid, jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when updating SVM ")
		_ = r.setConditionSVMUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
		r.Recorder.Event(svmCR, "Warning", "SvmUpdateFailed", "Error: "+err.Error())
		return err
	}
	log.Info("SVM updated successful")
	err = r.setConditionSVMUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
	r.Recorder.Event(svmCR, "Normal", "SvmUpdateSuccessed", "Updated SVM successfully")
	if err != nil {
		return nil //even though condition not create, don't reconcile again
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
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

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
