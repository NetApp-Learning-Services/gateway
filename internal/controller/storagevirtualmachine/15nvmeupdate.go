package controller

import (
	"context"
	"encoding/json"
	"fmt"
	gateway "gateway/api/v1beta3"
	"gateway/internal/controller/ontap"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NvmeLifServicePolicy = "default-data-nvme-tcp" //magic word

const NvmeLifServicePolicyScope = "svm" //magic word

func (r *StorageVirtualMachineReconciler) reconcileNvmeUpdate(ctx context.Context, svmCR *gateway.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {
	log.Info("STEP 15: Update NVMe service")

	// NVMe SERVICE

	createNvmeService := false
	updateNvmeService := false

	// Check to see if NVMe configuration is provided in custom resource
	if svmCR.Spec.NvmeConfig == nil {
		// If not, exit with no error
		log.Info("No NVMe service defined - skipping STEP 15")
		return nil
	}

	NvmeService, err := oc.GetNvmeServiceBySvmUuid(uuid)
	if err != nil && errors.IsNotFound(err) {
		createNvmeService = true
	} else if err != nil {
		//some other error
		log.Error(err, "Error retrieving NVMe service for SVM by UUID - requeuing")
		return err
	}

	var upsertNvmeService ontap.NvmeService

	if createNvmeService {
		log.Info("No NVMe service defined for SVM: " + uuid + " - creating NVMe service")

		upsertNvmeService.Svm.Uuid = svmCR.Spec.SvmUuid
		upsertNvmeService.Enabled = &svmCR.Spec.NvmeConfig.Enabled

		jsonPayload, err := json.Marshal(upsertNvmeService)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for NVMe service creation - requeuing")
			_ = r.setConditionNvmeService(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err

		}

		if oc.Debug {
			log.Info("[DEBUG] NVMe service creation payload: " + fmt.Sprintf("%#v\n", upsertNvmeService))
		}

		err = oc.CreateNvmeService(jsonPayload)
		if err != nil {
			log.Error(err, "Error creating the NVMe service - requeuing")
			_ = r.setConditionNvmeService(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "NvmeCreationFailed", "Error: "+err.Error())
			return err
		}
		_ = r.setConditionNvmeService(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "NvmeCreationSucceeded", "Created NVMe service successfully")
		log.Info("NVMe service created successful")
	} else {
		// Compare enabled to custom resource enabled
		if *NvmeService.Enabled != svmCR.Spec.NvmeConfig.Enabled {
			updateNvmeService = true
			upsertNvmeService.Enabled = &svmCR.Spec.NvmeConfig.Enabled
		}

		if oc.Debug && updateNvmeService {
			log.Info("[DEBUG] NVMe service update payload: " + fmt.Sprintf("%#v\n", upsertNvmeService))
		}

		if updateNvmeService {
			jsonPayload, err := json.Marshal(upsertNvmeService)
			if err != nil {
				//error creating the json body
				log.Error(err, "Error creating the json payload for NVMe service update - requeuing")
				_ = r.setConditionNvmeService(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

			//Patch NVMe service
			log.Info("NVMe service update attempt for SVM: " + uuid)
			err = oc.PatchNvmeService(uuid, jsonPayload)
			if err != nil {
				log.Error(err, "Error updating the NVMe service - requeuing")
				_ = r.setConditionNvmeService(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "NvmeUpdateFailed", "Error: "+err.Error())
				return err
			}
			log.Info("NVMe service updated successful")
			_ = r.setConditionNvmeService(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "NvmeUpdateSucceeded", "Updated NVMe service successfully")
		} else {
			log.Info("No NVMe service changes detected - skip updating")
		}
	}

	// END NVMe SERVICE

	// NVMe LIFS

	// Check to see if NVMe interfaces are defined in custom resource
	if svmCR.Spec.NvmeConfig.Lifs == nil {
		// If not, exit with no error
		log.Info("No NVMe LIFs defined - skipping updates")
		return nil
	}

	createNvmeLifs := false

	// Check to see if NVMe interfaces defined and compare to custom resource's definitions
	lifs, err := oc.GetNvmeInterfacesBySvmUuid(uuid, NvmeLifServicePolicy)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error getting NVMe service LIFs for SVM: "+uuid)
		_ = r.setConditionNvmeLif(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}

	if lifs.NumRecords == 0 {
		// no data LIFs for the SVM provided in UUID
		// create new LIF(s)
		log.Info("No LIFs defined for SVM: " + uuid + " - creating NVMe Lif(s)")
		createNvmeLifs = true
	}

	for index, val := range svmCR.Spec.NvmeConfig.Lifs {

		// Check to see need to create all LIFS or
		// if lifs.Records[index] is out of index - if so, need to create LIF
		if createNvmeLifs || index > lifs.NumRecords-1 {
			// Need to create LIF for val
			err = CreateLif(val, NvmeLifServicePolicy, NvmeLifServicePolicyScope, uuid, oc, log)
			if err != nil {
				_ = r.setConditionNvmeLif(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "NvmeCreationLifFailed", "Error: "+err.Error())
				return err
			}

		} else {
			//check to see if we need to update the LIF

			//do I need this? checking to see if I have valid LIF returned
			if reflect.ValueOf(lifs.Records[index]).IsZero() {
				break
			}

			err = UpdateLif(val, lifs.Records[index], NvmeLifServicePolicy, oc, log)
			if err != nil {
				_ = r.setConditionNvmeLif(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "NvmeUpdateLifFailed", "Error: "+err.Error())
				// e := err.(*apiError)
				// if e.errorCode == 1 {
				// // Json parsing error
				// 	return err
				// } else if e.errorCode == 2 {
				// // Patch error
				// 	return err
				// } else {
				// 	return err
				// }
				return err
			}
		}

		// Delete all SVM data LIFs that are not defined in the custom resource
		for i := len(svmCR.Spec.NvmeConfig.Lifs); i < lifs.NumRecords; i++ {
			log.Info("NVMe LIF delete attempt: " + lifs.Records[i].Name)
			oc.DeleteIpInterface(lifs.Records[i].Uuid)
			if err != nil {
				log.Error(err, "Error occurred when deleting NVMe LIF: "+lifs.Records[i].Name)
				// don't requeue on failed delete request
				// no condition error
				// return err
			} else {
				log.Info("NVMe LIF delete successful: " + lifs.Records[i].Name)
			}
		}

		_ = r.setConditionNvmeLif(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "NvmeUpsertLifSucceeded", "Upserted NVMe LIF(s) successfully")
	} // End looping through NVMe LIF definitions in custom resource

	// END NVMe LIFS

	return nil
}

// STEP 15
// NVMe update
// Note: Status of NVME_SERVICE can only be true or false
const CONDITION_TYPE_NVME_SERVICE = "15NVMeservice"
const CONDITION_REASON_NVME_SERVICE = "NVMeservice"
const CONDITION_MESSAGE_NVME_SERVICE_TRUE = "NVMe service configuration succeeded"
const CONDITION_MESSAGE_NVME_SERVICE_FALSE = "NVMe service configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNvmeService(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NVME_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NVME_SERVICE, CONDITION_REASON_NVME_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NVME_SERVICE, status,
			CONDITION_REASON_NVME_SERVICE, CONDITION_MESSAGE_NVME_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NVME_SERVICE, status,
			CONDITION_REASON_NVME_SERVICE, CONDITION_MESSAGE_NVME_SERVICE_FALSE)
	}
	return nil
}

const CONDITION_REASON_NVME_LIF = "NVMelif"
const CONDITION_MESSAGE_NVME_LIF_TRUE = "NVMe LIF configuration succeeded"
const CONDITION_MESSAGE_NVME_LIF_FALSE = "NVMe LIF configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNvmeLif(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NVME_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NVME_SERVICE, CONDITION_REASON_NVME_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NVME_SERVICE, status,
			CONDITION_REASON_NVME_LIF, CONDITION_MESSAGE_NVME_LIF_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NVME_SERVICE, status,
			CONDITION_REASON_NVME_LIF, CONDITION_MESSAGE_NVME_LIF_FALSE)
	}
	return nil
}
