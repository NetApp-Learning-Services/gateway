package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
)

const NfsLifType = "default-data-files" //special word
const NfsLifScope = "svm"               //special word

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

		upsertNfsService.Enabled = &svmCR.Spec.NfsConfig.Enabled
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
		if nfsService.Enabled != &svmCR.Spec.NfsConfig.Enabled {
			updateNfsService = true
			upsertNfsService.Enabled = &svmCR.Spec.NfsConfig.Enabled
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
	if svmCR.Spec.NfsConfig.Lifs == nil {
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
			for _, val := range svmCR.Spec.NfsConfig.Lifs {
				err = CreateLIF(val, uuid, oc, log)
				if err != nil {
					return err
				}
			}

		} else {
			// update LIFs
			for index, val := range svmCR.Spec.NfsConfig.Lifs {

				// Check to see if lifs.Records[index] is out of index - if so, need to create LIF
				if index > lifs.NumRecords-1 {
					// Need to create LIF for val
					err = CreateLIF(val, uuid, oc, log)
					if err != nil {
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
							log.Error(err, "Error creating the json payload for NFS LIF update: "+val.Name)
							//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
							return err
						}
						log.Info("NFS LIF update attempt: " + val.Name)
						err = oc.PatchIpInterface(lifs.Records[index].Uuid, jsonPayload)
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

			} // Need looping through NFS LIF definitions in custom resource

			// Delete all SVM data LIFs that are not defined in the custom resource
			for i := len(svmCR.Spec.NfsConfig.Lifs); i < lifs.NumRecords; i++ {
				log.Info("NFS LIF delete attempt: " + lifs.Records[i].Name)
				oc.DeleteIpInterface(lifs.Records[i].Uuid)
				if err != nil {
					log.Error(err, "Error occurred when deleting NFS LIF: "+lifs.Records[i].Name)
					//TODO: _ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
					// don't requeue on failed delete request
					// return err
				} else {
					log.Info("NFS LIF delete successful: " + lifs.Records[i].Name)
				}
			}

		} // Checking for NFS LIFs updates
	} // LIFs defined in custom resource

	// Check to see if NFS rules are defined in custom resources
	if svmCR.Spec.NfsConfig.Export == nil {
		// If not, exit with no error
		log.Info("No NFS export rules defined - skipping")
	} else {
		exportsCreate := false

		// Check to see if NFS interfaces defined and compare to custom resource's definitions
		exportRetrieved, err := oc.GetNfsExportBySvmUuid(uuid)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error getting NFS export rules for SVM: "+uuid)
			//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if exportRetrieved.NumRecords == 0 {
			// no data LIFs for the SVM provided in UUID
			// create new LIF(s)
			log.Info("No LIFs defined for SVM: " + uuid + " - creating NFS Lif(s)")
			exportsCreate = true
		}

		if exportsCreate {
			// creating export
			err = CreateExport(*svmCR.Spec.NfsConfig.Export, uuid, oc, log)
			if err != nil {
				return err
			}

		} else {

			// if more than one export, delete anything after the first one
			for i := 1; i < exportRetrieved.NumRecords; i++ {
				log.Info("NFS export delete attempt: " + exportRetrieved.Records[i].Name)
				oc.DeleteNfsExport(exportRetrieved.Records[i].Id)
				if err != nil {
					log.Error(err, "Error occurred when deleting NFS export: "+exportRetrieved.Records[i].Name)
					//TODO: _ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
					// don't requeue on failed delete request
					// return err
				} else {
					log.Info("NFS LIF delete successful: " + exportRetrieved.Records[i].Name)
				}
			}

			// check for updating export
			exportUpdate := false

			if exportRetrieved.Records[0].Name != svmCR.Spec.NfsConfig.Export.Name {
				//need to update it
				exportUpdate = true
			}

			for indx, val := range svmCR.Spec.NfsConfig.Export.Rules {

				if val.Anon != exportRetrieved.Records[0].Rules[indx].Anonuser {
					exportUpdate = true
				}

				if val.Client != exportRetrieved.Records[0].Rules[indx].Anonuser {
					exportUpdate = true
				}
				if val.Protocols != exportRetrieved.Records[0].Rules[indx].Protocols {
					exportUpdate = true
				}

				if val.Ro != exportRetrieved.Records[0].Rules[indx].RoRule {
					exportUpdate = true
				}
				if val.Rw != exportRetrieved.Records[0].Rules[indx].RwRule {
					exportUpdate = true
				}
				if val.Superuser != exportRetrieved.Records[0].Rules[indx].Superuser {
					exportUpdate = true
				}

			}

			if exportUpdate {
				// Build out complete export if there is any change
				idToReplace := exportRetrieved.Records[0].Id // get the id
				var exportUpdateVal ontap.ExportPolicy
				exportUpdateVal.Name = svmCR.Spec.NfsConfig.Export.Name

				for _, val := range svmCR.Spec.NfsConfig.Export.Rules {
					var exportRuleToAdd ontap.ExportRule
					exportRuleToAdd.Anonuser = val.Anon
					exportRuleToAdd.Protocols = val.Protocols
					exportRuleToAdd.RwRule = val.Rw
					exportRuleToAdd.Superuser = val.Superuser
					exportUpdateVal.Rules = append(exportUpdateVal.Rules, exportRuleToAdd)
				}

				// otherwise changes need to be implemented
				if oc.Debug {
					log.Info("[DEBUG] NFS export update payload: " + fmt.Sprintf("%#v\n", exportUpdateVal))
				}

				jsonPayload, err := json.Marshal(exportUpdateVal)
				if err != nil {
					//error creating the json body
					log.Error(err, "Error creating the json payload for NFS export update")
					//ToDO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				}
				err = oc.PatchNfsExport(idToReplace, jsonPayload)
				if err != nil {
					log.Error(err, "Error occurred when updating NFS export")
					//Todo: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				}
				log.Info("NFS export updated successful")
				// err = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
				// if err != nil {
				// 	return nil //even though condition not create, don't reconcile again
				// }

			} else {
				log.Info("No NFS export rules changed detected - skipping")
			}

		} //Check for NFS export updates or deletion of none conforming exports

	} // NFS exports rules defined in custom resource

	return nil
}

func CreateLIF(lifToCreate gatewayv1alpha1.LIF, uuid string, oc *ontap.Client, log logr.Logger) (err error) {
	var newLif ontap.IpInterface
	newLif.Name = lifToCreate.Name
	newLif.Ip.Address = lifToCreate.IPAddress
	newLif.Ip.Netmask = lifToCreate.Netmask
	newLif.Location.BroadcastDomain.Name = lifToCreate.BroacastDomain
	newLif.Location.HomeNode.Name = lifToCreate.HomeNode
	newLif.ServicePolicy.Name = NfsLifType
	newLif.Scope = NfsLifScope
	newLif.Svm.Uuid = uuid

	jsonPayload, err := json.Marshal(newLif)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for NFS LIF creation: "+lifToCreate.Name)
		//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}
	log.Info("NFS LIF creation attempt: " + lifToCreate.Name)
	err = oc.CreateIpInterface(jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when creating NFS LIF: "+lifToCreate.Name)
		//TODO: _ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}
	log.Info("NFS LIF creation successful: " + lifToCreate.Name)
	// err = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_TRUE)
	// if err != nil {
	// 	return nil //even though condition not create, don't reconcile again
	// }

	return nil
}

func CreateExport(exportToCreate gatewayv1alpha1.NfsExport, uuid string, oc *ontap.Client, log logr.Logger) (err error) {
	var newExport ontap.ExportPolicy
	newExport.Name = exportToCreate.Name

	for _, val := range exportToCreate.Rules {
		var newRule ontap.ExportRule
		newRule.Protocols = val.Protocols
		newRule.RwRule = val.Rw
		newRule.RoRule = val.Ro
		newRule.Anonuser = val.Anon
		newRule.Superuser = val.Superuser
		newExport.Rules = append(newExport.Rules, newRule)
	}

	newExport.Svm.Uuid = uuid

	jsonPayload, err := json.Marshal(newExport)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for NFS export creation: "+exportToCreate.Name)
		//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}
	log.Info("NFS export creation attempt: " + exportToCreate.Name)
	err = oc.CreateIpInterface(jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when creating NFS export: "+exportToCreate.Name)
		//TODO: _ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}
	log.Info("NFS export creation successful: " + exportToCreate.Name)
	// err = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_TRUE)
	// if err != nil {
	// 	return nil //even though condition not create, don't reconcile again
	// }

	return nil
}
