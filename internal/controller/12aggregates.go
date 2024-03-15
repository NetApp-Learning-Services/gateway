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

func (r *StorageVirtualMachineReconciler) reconcileAggregates(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, svmRetrieved ontap.SvmByUUID, oc *ontap.Client, log logr.Logger) error {

	log.Info("STEP 12: Update SVM aggregates")
	var patchSVM ontap.SvmAggregatePatch
	needToUpdate := true //default to always update #16

	// interate over custom resoource svmCR and look for differences in retrieved SVM
	if svmCR.Spec.Aggregates != nil {
		//aggregates defined

		for _, val := range svmCR.Spec.Aggregates {
			r := doesElementExist(svmRetrieved.Aggregates, val.Name)
			if !r {
				needToUpdate = true
			}
		}

		if needToUpdate {
			//Add all aggregates to the patch request to prevent any aggregates missing
			for _, val := range svmCR.Spec.Aggregates {
				var res ontap.Resource
				res.Name = val.Name
				patchSVM.Aggregates = append(patchSVM.Aggregates, res)
			}

			if oc.Debug {
				defaultLog.Printf("[DEBUG] SVM aggregates payload: " + fmt.Sprintf("%#v\n", patchSVM))
			}

			jsonPayload, err := json.Marshal(patchSVM)
			if err != nil {
				//error creating the json body
				log.Error(err, "Error creating the json payload for SVM aggregates update - requeuing")
				_ = r.setConditionAggregateAssigned(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

			// After building update string execute it and check for errors
			log.Info("SVM aggregates update attempt for SVM: " + svmRetrieved.Uuid)
			err = oc.PatchStorageVM(svmRetrieved.Uuid, jsonPayload)
			if err != nil {
				log.Error(err, "Error occurred when updating SVM aggregates - requeuing")
				_ = r.setConditionAggregateAssigned(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "SvmUpdateAggregateFailed", "Update SVM aggregate(s) failed")
				return err
			}
			log.Info("SVM aggregates updated successful")
			_ = r.setConditionAggregateAssigned(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "SvmUpdateAggregateSucceeded", "Updated SVM aggregate(s) successfully")

		} else {
			log.Info("No changes detected for SVM aggregates - skipping STEP 12")
			return nil
		}

	} else {
		log.Info("No SVM aggregates defined - skipping STEP 12")
	}

	return nil
}

func doesElementExist(s []ontap.Aggregate, str string) bool {
	for _, v := range s {
		if v.Name == str {
			return true
		}
	}
	return false
}

// STEP 12
// Aggregate assigned
// Note: Status of AGGREGATE_ASSIGNED can only be true or false
const CONDITION_TYPE_AGGREGATE_ASSIGNED = "12AggregateAssigned"
const CONDITION_REASON_AGGREGATE_ASSIGNED = "AggregateAssigned"
const CONDITION_MESSAGE_AGGREGATE_ASSIGNED_TRUE = "Aggregate assigned to SVM succeeded"
const CONDITION_MESSAGE_AGGREGATE_ASSIGNED_FALSE = "Aggregate assigned to SVM failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionAggregateAssigned(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

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
