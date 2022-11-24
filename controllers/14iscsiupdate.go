package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	gatewayv1alpha2 "gateway/api/v1alpha2"
	"gateway/ontap"
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *StorageVirtualMachineReconciler) reconcileIscsiUpdate(ctx context.Context, svmCR *gatewayv1alpha2.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {
	log.Info("STEP 14: Update iSCSI service")

	// iSCSI SERVICE
	create := false
	updateIscsiService := false

	// Check to see if iscsi configuration is provided in custom resource
	if svmCR.Spec.IscsiConfig == nil {
		// If not, exit with no error
		log.Info("No iSCSI service defined - skipping STEP 14")
		return nil
	}

	iscsiService, err := oc.GetIscsiServiceBySvmUuid(uuid)
	if err != nil && errors.IsNotFound(err) {
		create = true
	} else if err != nil {
		//some other error
		log.Error(err, "Error retrieving iSCSI service for SVM by UUID - requeuing")
		return err
	}

	var upsertIscsiService ontap.IscsiService

	if create {
		upsertIscsiService.Enabled = &svmCR.Spec.IscsiConfig.Enabled
		upsertIscsiService.Target.Alias = svmCR.Spec.IscsiConfig.Alias
		upsertIscsiService.Svm.Uuid = svmCR.Spec.SvmUuid

		jsonPayload, err := json.Marshal(upsertIscsiService)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for iSCSI service creation - requeuing")
			_ = r.setConditionIscsiService(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err

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
		if *iscsiService.Enabled != svmCR.Spec.NfsConfig.Enabled {
			updateIscsiService = true
			upsertIscsiService.Enabled = &svmCR.Spec.IscsiConfig.Enabled
		}

		if iscsiService.Target.Alias != svmCR.Spec.IscsiConfig.Alias {
			updateIscsiService = true
			upsertIscsiService.Target.Alias = svmCR.Spec.IscsiConfig.Alias
		}

		if oc.Debug {
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
	if svmCR.Spec.NfsConfig.Lifs == nil {
		// If not, exit with no error
		log.Info("No iSCSI LIFs defined - skipping updates")
	} else {
		lifsCreate := false

		// Check to see if NFS interfaces defined and compare to custom resource's definitions
		lifs, err := oc.GetIscsiInterfacesBySvmUuid(uuid)
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
			lifsCreate = true
		}

		if lifsCreate {
			//creating lifs
			for _, val := range svmCR.Spec.IscsiConfig.Lifs {
				err = CreateLIF(val, uuid, oc, log)
				if err != nil {
					_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				}
			}

		} else {
			// update LIFs
			for index, val := range svmCR.Spec.IscsiConfig.Lifs {

				// Check to see if lifs.Records[index] is out of index - if so, need to create LIF
				if index > lifs.NumRecords-1 {
					// Need to create LIF for val
					err = CreateLIF(val, uuid, oc, log)
					if err != nil {
						_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "IscsiCreationLifFailed", "Error: "+err.Error())
						return err
					}

				} else {
					netmaskAsInt, _ := strconv.Atoi(lifs.Records[index].Ip.Netmask)
					netmaskAsIP := NetmaskToString(netmaskAsInt)
					if lifs.Records[index].Ip.Address != val.IPAddress ||
						lifs.Records[index].Name != val.Name ||
						netmaskAsIP != val.Netmask ||
						lifs.Records[index].ServicePolicy.Name != NfsLifType ||
						!lifs.Records[index].Enabled {
						//reset value
						var updateLif ontap.IpInterface
						updateLif.Name = val.Name
						updateLif.Ip.Address = val.IPAddress
						updateLif.Ip.Netmask = val.Netmask
						//updateLif.Location.BroadcastDomain.Name = val.BroacastDomain
						//updateLif.Location.HomeNode.Name = val.HomeNode
						updateLif.ServicePolicy.Name = NfsLifType
						updateLif.Enabled = true

						jsonPayload, err := json.Marshal(updateLif)
						if err != nil {
							//error creating the json body
							log.Error(err, "Error creating the json payload for iSCSI LIF update: "+val.Name+" - requeuing")
							_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_FALSE)
							r.Recorder.Event(svmCR, "Warning", "IscsiUpdateLifFailed", "Error: "+err.Error())
							return err
						}
						log.Info("iSCSI LIF update attempt: " + val.Name)
						err = oc.PatchIpInterface(lifs.Records[index].Uuid, jsonPayload)
						if err != nil {
							log.Error(err, "Error occurred when updating iSCSI LIF: "+val.Name+" - requeuing")
							_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_FALSE)
							r.Recorder.Event(svmCR, "Warning", "IscsiUpdateLifFailed", "Error: "+err.Error())
							return err
						}

						log.Info("iSCSI LIF update successful: " + val.Name)
					} else {
						log.Info("No changes detected for iSCSI LIf: " + val.Name)
					}
				}

			} // Need looping through iSCSI LIF definitions in custom resource

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

		} // Checking for iSCSI LIFs updates
		_ = r.setConditionIscsiLif(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "IscsiUpsertLifSucceeded", "Upserted iSCSI LIF(s) successfully")
	} // LIFs defined in custom resource
	// END ISCSI LIFS

	return nil
}
