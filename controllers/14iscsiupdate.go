package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	gatewayv1alpha2 "gateway/api/v1alpha2"
	"gateway/ontap"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
)

const IscsiLifType = "default-data-blocks" //magic word
/*
todo: check on this
if 9.9.1 - use default-data-blocks
if 9.11.1 - use default-data-iscsi
*/
const IscsiLifScope = "svm" //magic word

func (r *StorageVirtualMachineReconciler) reconcileIscsiUpdate(ctx context.Context, svmCR *gatewayv1alpha2.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {
	log.Info("STEP 14: Update iSCSI service")

	// iSCSI SERVICE

	createIscsiService := false
	updateIscsiService := false

	// Check to see if iscsi configuration is provided in custom resource
	if svmCR.Spec.IscsiConfig == nil {
		// If not, exit with no error
		log.Info("No iSCSI service defined - skipping STEP 14")
		return nil
	}

	iscsiService, err := oc.GetIscsiServiceBySvmUuid(uuid)
	if err != nil && errors.IsNotFound(err) {
		createIscsiService = true
	} else if err != nil {
		//some other error
		log.Error(err, "Error retrieving iSCSI service for SVM by UUID - requeuing")
		return err
	}

	var upsertIscsiService ontap.IscsiService

	if createIscsiService {
		log.Info("No iSCSI service defined for SVM: " + uuid + " - creating iSCSI service")
		upsertIscsiService.Enabled = &svmCR.Spec.IscsiConfig.Enabled
		alias := svmCR.Spec.SvmName
		if svmCR.Spec.IscsiConfig.Alias != "" {
			alias = svmCR.Spec.IscsiConfig.Alias
		}
		upsertIscsiService.Target.Alias = alias
		upsertIscsiService.Svm.Uuid = svmCR.Spec.SvmUuid

		jsonPayload, err := json.Marshal(upsertIscsiService)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for iSCSI service creation - requeuing")
			_ = r.setConditionIscsiService(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err

		}

		if oc.Debug {
			log.Info("[DEBUG] iSCSI service creation payload: " + fmt.Sprintf("%#v\n", upsertIscsiService))
		}

		err = oc.CreateIscsiService(jsonPayload)
		if err != nil {
			log.Error(err, "Error creating the iSCSI service - requeuing")
			_ = r.setConditionIscsiService(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "IscsiCreationFailed", "Error: "+err.Error())
			return err
		}
		_ = r.setConditionIscsiService(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "IscsiCreationSucceeded", "Created iSCSI service successfully")
		log.Info("iSCSI service created successful")
	} else {
		// Compare enabled to custom resource enabled
		if *iscsiService.Enabled != svmCR.Spec.IscsiConfig.Enabled {
			updateIscsiService = true
			upsertIscsiService.Enabled = &svmCR.Spec.IscsiConfig.Enabled
			upsertIscsiService.Svm.Uuid = svmCR.Spec.SvmUuid // always add the SVM UUID
		}

		if svmCR.Spec.IscsiConfig.Alias != "" && iscsiService.Target.Alias != svmCR.Spec.IscsiConfig.Alias {
			updateIscsiService = true
			upsertIscsiService.Target.Alias = svmCR.Spec.IscsiConfig.Alias
			upsertIscsiService.Svm.Uuid = svmCR.Spec.SvmUuid // always add the SVM UUID
		}

		if oc.Debug && updateIscsiService {
			log.Info("[DEBUG] iSCSI service update payload: " + fmt.Sprintf("%#v\n", upsertIscsiService))
		}

		if updateIscsiService {
			jsonPayload, err := json.Marshal(upsertIscsiService)
			if err != nil {
				//error creating the json body
				log.Error(err, "Error creating the json payload for iSCSI service update - requeuing")
				_ = r.setConditionIscsiService(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

			//Patch iSCSI service
			log.Info("iSCSI service update attempt for SVM: " + uuid)
			err = oc.PatchIscsiService(uuid, jsonPayload)
			if err != nil {
				log.Error(err, "Error updating the iSCSI service - requeuing")
				_ = r.setConditionIscsiService(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "IscsiUpdateFailed", "Error: "+err.Error())
				return err
			}
			log.Info("iSCSI service updated successful")
			_ = r.setConditionIscsiService(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "IscsiUpdateSucceeded", "Updated iSCSI service successfully")
		} else {
			log.Info("No iSCSI service changes detected - skip updating")
		}
	}

	// END ISCSI SERVICE

	// ISCSI LIFS

	// Check to see if ISCSI interfaces are defined in custom resource
	if svmCR.Spec.IscsiConfig.Lifs == nil {
		// If not, exit with no error
		log.Info("No iSCSI LIFs defined - skipping updates")
		return nil
	}

	createIscsiLifs := false

	// Check to see if iSCSI interfaces defined and compare to custom resource's definitions
	lifs, err := oc.GetIscsiInterfacesBySvmUuid(uuid, IscsiLifType)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error getting iSCSI service LIFs for SVM: "+uuid)
		_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}

	if lifs.NumRecords == 0 {
		// no data LIFs for the SVM provided in UUID
		// create new LIF(s)
		log.Info("No LIFs defined for SVM: " + uuid + " - creating iSCSI Lif(s)")
		createIscsiLifs = true
	}

	for index, val := range svmCR.Spec.IscsiConfig.Lifs {

		// Check to see need to create all LIFS or
		// if lifs.Records[index] is out of index - if so, need to create LIF
		if createIscsiLifs || index > lifs.NumRecords-1 {
			// Need to create LIF for val
			err = CreateLif(val, IscsiLifType, uuid, oc, log)
			if err != nil {
				_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "IscsiCreationLifFailed", "Error: "+err.Error())
				return err
			}

		} else {
			//check to see if we need to update the LIF

			//do I need this? checking to see if I have valid LIF returned
			if reflect.ValueOf(lifs.Records[index]).IsZero() {
				break
			}

			err = UpdateLif(val, lifs.Records[index], IscsiLifType, oc, log)
			if err != nil {
				_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "IscsiUpdateLifFailed", "Error: "+err.Error())
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
		for i := len(svmCR.Spec.IscsiConfig.Lifs); i < lifs.NumRecords; i++ {
			log.Info("iSCSI LIF delete attempt: " + lifs.Records[i].Name)
			oc.DeleteIpInterface(lifs.Records[i].Uuid)
			if err != nil {
				log.Error(err, "Error occurred when deleting iSCSI LIF: "+lifs.Records[i].Name)
				// don't requeue on failed delete request
				// no condition error
				// return err
			} else {
				log.Info("iSCSI LIF delete successful: " + lifs.Records[i].Name)
			}
		}

		_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "IscsiUpsertLifSucceeded", "Upserted iSCSI LIF(s) successfully")
	} // End looping through iSCSI LIF definitions in custom resource

	// END ISCSI LIFS

	return nil
}
