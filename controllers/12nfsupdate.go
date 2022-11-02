package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *StorageVirtualMachineReconciler) reconcileNFSUpdate(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {

	log.Info("Step 12: Update NFS service")

	create := false
	updateNfsService := false

	// Check to see if nfs configuration is provided in custom resource
	if svmCR.Spec.NfsConfig == nil {
		// If not, exit with no error
		log.Info("No NFS service defined - skipping Step 12")
		return nil
	}

	// Get the NFS configuration of SVM
	nfsService, err := oc.GetNfsServiceBySvmUuid(uuid)
	if err != nil && errors.IsNotFound(err) {
		create = true
	} else if err != nil {
		// some other error
		log.Error(err, "Error retrieving NFS service for SVM by UUID")
		return err
	}

	var upsertNfsService ontap.NFSService

	if create {

		upsertNfsService.Enabled = &svmCR.Spec.NfsConfig.NfsEnabled
		upsertNfsService.Protocol.V3Enable = &svmCR.Spec.NfsConfig.Nfsv3
		upsertNfsService.Protocol.V4Enable = &svmCR.Spec.NfsConfig.Nfsv4
		upsertNfsService.Protocol.V41Enable = &svmCR.Spec.NfsConfig.Nfsv41
		upsertNfsService.Svm.Uuid = svmCR.Spec.SvmUuid

		jsonPayload, err := json.Marshal(upsertNfsService)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for NFS service creation")
			//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		err = oc.CreateNfsService(jsonPayload)
		if err != nil {
			log.Error(err, "Error creating the NFS service")
			//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}
		log.Info("NFS service created successful")
	} else {

		// Compare enabled to custom resource enabled
		if nfsService.Enabled != &svmCR.Spec.NfsConfig.NfsEnabled {
			updateNfsService = true
			upsertNfsService.Enabled = &svmCR.Spec.NfsConfig.NfsEnabled
		}

		if nfsService.Protocol.V3Enable != &svmCR.Spec.NfsConfig.Nfsv3 {
			updateNfsService = true
			upsertNfsService.Protocol.V3Enable = &svmCR.Spec.NfsConfig.Nfsv3
		}

		if nfsService.Protocol.V4Enable != &svmCR.Spec.NfsConfig.Nfsv4 {
			updateNfsService = true
			upsertNfsService.Protocol.V4Enable = &svmCR.Spec.NfsConfig.Nfsv4
		}

		if nfsService.Protocol.V41Enable != &svmCR.Spec.NfsConfig.Nfsv41 {
			updateNfsService = true
			upsertNfsService.Protocol.V41Enable = &svmCR.Spec.NfsConfig.Nfsv41
		}

		if oc.Debug {
			log.Info("[DEBUG] NFS service update payload: " + fmt.Sprintf("%#v\n", upsertNfsService))
		}

		if updateNfsService {

			jsonPayload, err := json.Marshal(upsertNfsService)
			if err != nil {
				//error creating the json body
				log.Error(err, "Error creating the json payload for NFS service update")
				//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

			//Patch Nfs service
			log.Info("NFS service update attempt for SVM: " + uuid)
			err = oc.PatchNfsService(uuid, jsonPayload)
			if err != nil {
				log.Error(err, "Error updating the NFS service")
				//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}
			log.Info("NFS service updated successful")
		} else {
			log.Info("No NFS service changes detected - skip updating")
		}
	}

	// Check to see if NFS interfaces are defined in custom resource
	if svmCR.Spec.NfsConfig.NfsLifs == nil {
		// If not, exit with no error
		log.Info("No NFS LIFs defined - skipping updates")
	} else {
		lifsCreate := false

		// Check to see if NFS interfaces defined and compare to custom resource's definitions
		lifs, err := oc.GetNfsInterfacesBySvmUuid(uuid)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error getting NFS service LIFs for SVM: "+uuid)
			//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if lifs.NumRecords == 0 {
			// no data LIFs for the SVM provided in UUID
			// create new LIF(s)
			log.Info("No LIFs defined for SVM: " + uuid + " - creating NFS Lif(s)")
			lifsCreate = true
		}

		if lifsCreate {
			//creating lifs
			for _, val := range svmCR.Spec.NfsConfig.NfsLifs {
				var newLif ontap.IpInterface
				newLif.Name = val.Name
				newLif.Ip.Address = val.IPAddress
				newLif.Ip.Netmask = val.Netmask
				newLif.Location.BroadcastDomain.Name = val.BroacastDomain
				newLif.Location.HomeNode.Name = val.HomeNode
				newLif.ServicePolicy.Name = "default-data-files" //special word
				newLif.Scope = "svm"                             //special word
				newLif.Svm.Uuid = uuid

				jsonPayload, err := json.Marshal(newLif)
				if err != nil {
					//error creating the json body
					log.Error(err, "Error creating the json payload for NFS LIF creation: "+val.Name)
					//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				}
				log.Info("NFS LIF creation attempt: " + val.Name)
				err = oc.CreateIpInterface(jsonPayload)
				if err != nil {
					log.Error(err, "Error occurred when creating NFS LIF: "+val.Name)
					//TODO: _ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				}
				log.Info("NFS LIF creation successful: " + val.Name)
				// err = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_TRUE)
				// if err != nil {
				// 	return nil //even though condition not create, don't reconcile again
				// }
			}

		} else {
			for index, val := range svmCR.Spec.NfsConfig.NfsLifs {
				if val.IPAddress != lifs.Records[index].Ip.Address || val.Name != lifs.Records[index].Name {
					//reset value
					var updateLif ontap.IpInterface = lifs.Records[index]
					updateLif.Name = val.Name
					updateLif.Ip.Address = val.IPAddress
					updateLif.Ip.Netmask = val.Netmask
					updateLif.Location.BroadcastDomain.Name = val.BroacastDomain
					updateLif.Location.HomeNode.Name = val.HomeNode
					updateLif.ServicePolicy.Name = "default-data-files" //special word
					updateLif.Enabled = true

					jsonPayload, err := json.Marshal(updateLif)
					if err != nil {
						//error creating the json body
						log.Error(err, "Error creating the json payload for NFS LIF update: "+val.Name)
						//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
						return err
					}
					log.Info("NFS LIF update attempt: " + val.Name)
					err = oc.PatchIpInterface(updateLif.Uuid, jsonPayload)
					if err != nil {
						log.Error(err, "Error occurred when updating NFS LIF: "+val.Name)
						//TODO: _ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
						return err
					}
					log.Info("NFS LIF update successful: " + val.Name)
				} else {
					log.Info("No changes detected for NFS LIf: " + val.Name)
				}
			}
		} // Checking for NFS LIFs updates
	} //LIFs defined in custom resource

	// Check to see if NFS rules are defined in custom resources
	if svmCR.Spec.NfsConfig.NfsRules == nil {
		// If not, exit with no error
		log.Info("No NFS export rules - skipping")
	} else {
		// If so, GET /protocols/nfs/export-policies compare rules with result based upon index/id
		// PATCH /protocols/nfs/export-policies/id if needed
		// If rule missing in ONTAP, POST /protocols/nfs/export-policies/
	}

	return nil
}
