package controller

import (
	"context"
	"encoding/json"
	"fmt"

	gateway "gateway/api/v1beta1"
	"gateway/internal/controller/ontap"

	"github.com/go-logr/logr"
)

func (r *StorageVirtualMachineReconciler) reconcileAggregates(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, svmRetrieved ontap.SvmByUUID, oc *ontap.Client, log logr.Logger) error {

	log.Info("STEP 12: Update SVM aggregates")
	var patchSVM ontap.SvmPatch
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

			log.Info("SVM aggregates payload: " + fmt.Sprintf("%#v\n", patchSVM))

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
