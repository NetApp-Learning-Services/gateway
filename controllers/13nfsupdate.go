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

const NfsLifType = "default-data-files" //magic word
const NfsLifScope = "svm"               //magic word

func (r *StorageVirtualMachineReconciler) reconcileNFSUpdate(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {

	log.Info("STEP 13: Update NFS service")

	// NFS SERVICE
	create := false
	updateNfsService := false

	// Check to see if nfs configuration is provided in custom resource
	if svmCR.Spec.NfsConfig == nil {
		// If not, exit with no error
		log.Info("No NFS service defined - skipping STEP 13")
		return nil
	}

	// Get the NFS configuration of SVM
	nfsService, err := oc.GetNfsServiceBySvmUuid(uuid)
	if err != nil && errors.IsNotFound(err) {
		create = true
	} else if err != nil {
		// some other error
		log.Error(err, "Error retrieving NFS service for SVM by UUID - requeuing")
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
			log.Error(err, "Error creating the json payload for NFS service creation - requeuing")
			_ = r.setConditionNfsService(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		err = oc.CreateNfsService(jsonPayload)
		if err != nil {
			log.Error(err, "Error creating the NFS service - requeuing")
			_ = r.setConditionNfsService(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "NfsCreationFailed", "Error: "+err.Error())
			return err
		}
		_ = r.setConditionNfsService(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "NfsCreationSucceeded", "Created NFS service successfully")
		log.Info("NFS service created successful")
	} else {

		// Compare enabled to custom resource enabled
		if *nfsService.Enabled != svmCR.Spec.NfsConfig.Enabled {
			updateNfsService = true
			upsertNfsService.Enabled = &svmCR.Spec.NfsConfig.Enabled
		}

		if *nfsService.Protocol.V3Enable != svmCR.Spec.NfsConfig.Nfsv3 {
			updateNfsService = true
			upsertNfsService.Protocol.V3Enable = &svmCR.Spec.NfsConfig.Nfsv3
		}

		if *nfsService.Protocol.V4Enable != svmCR.Spec.NfsConfig.Nfsv4 {
			updateNfsService = true
			upsertNfsService.Protocol.V4Enable = &svmCR.Spec.NfsConfig.Nfsv4
		}

		if *nfsService.Protocol.V41Enable != svmCR.Spec.NfsConfig.Nfsv41 {
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
				log.Error(err, "Error creating the json payload for NFS service update - requeuing")
				_ = r.setConditionNfsService(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

			//Patch Nfs service
			log.Info("NFS service update attempt for SVM: " + uuid)
			err = oc.PatchNfsService(uuid, jsonPayload)
			if err != nil {
				log.Error(err, "Error updating the NFS service - requeuing")
				_ = r.setConditionNfsService(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "NfsUpdateFailed", "Error: "+err.Error())
				return err
			}
			log.Info("NFS service updated successful")
			_ = r.setConditionNfsService(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.Recorder.Event(svmCR, "Normal", "NfsUpdateSucceeded", "Updated NFS service successfully")
		} else {
			log.Info("No NFS service changes detected - skip updating")
		}
	}
	// END NFS SERVICE

	// NFS LIFS
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
			_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_FALSE)
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
					_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_FALSE)
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
						_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "NfsCreationLifFailed", "Error: "+err.Error())
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
							log.Error(err, "Error creating the json payload for NFS LIF update: "+val.Name+" - requeuing")
							_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_FALSE)
							r.Recorder.Event(svmCR, "Warning", "NfsUpdateLifFailed", "Error: "+err.Error())
							return err
						}
						log.Info("NFS LIF update attempt: " + val.Name)
						err = oc.PatchIpInterface(lifs.Records[index].Uuid, jsonPayload)
						if err != nil {
							log.Error(err, "Error occurred when updating NFS LIF: "+val.Name+" - requeuing")
							_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_FALSE)
							r.Recorder.Event(svmCR, "Warning", "NfsUpdateLifFailed", "Error: "+err.Error())
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
					// don't requeue on failed delete request
					// no condition error
					// return err
				} else {
					log.Info("NFS LIF delete successful: " + lifs.Records[i].Name)
				}
			}

		} // Checking for NFS LIFs updates
		_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "NfsUpsertLifSucceeded", "Upserted NFS LIF(s) successfully")
	} // LIFs defined in custom resource
	// END NFS LIFS

	// NFS EXPORTS
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
			log.Error(err, "Error getting NFS export rules for SVM: "+uuid+" - requeuing")
			_ = r.setConditionNfsExport(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if exportRetrieved.NumRecords == 0 {
			// No exports for the SVM provided in UUID
			// create new export(s)
			log.Info("No exports defined for SVM: " + uuid + " - creating NFS export(s)")
			exportsCreate = true
		}

		if exportsCreate {
			// this will probably never happen because there is always a default export
			// creating export
			err = CreateExport(*svmCR.Spec.NfsConfig.Export, uuid, oc, log)
			if err != nil {
				_ = r.setConditionNfsExport(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

		} else {

			// if more than one export, delete anything after the first one
			// never delete the first export
			for i := 1; i < exportRetrieved.NumRecords; i++ {
				log.Info("NFS export delete attempt: " + exportRetrieved.Records[i].Name)
				oc.DeleteNfsExport(exportRetrieved.Records[i].Id)
				if err != nil {
					log.Error(err, "Error occurred when deleting NFS export: "+exportRetrieved.Records[i].Name+" - requeuing")
					// no condition error
					// don't requeue on failed delete request
					// return err
				} else {
					log.Info("NFS export delete successful: " + exportRetrieved.Records[i].Name)
				}
			}

			// check for updating export
			exportUpdate := false

			if exportRetrieved.Records[0].Name != svmCR.Spec.NfsConfig.Export.Name {
				//need to update it
				exportUpdate = true
			}

			for indx, val := range svmCR.Spec.NfsConfig.Export.Rules {

				if len(exportRetrieved.Records[0].Rules) == 0 {
					//nothing defined
					exportUpdate = true
				} else {
					if val.Anon != exportRetrieved.Records[0].Rules[indx].Anonuser {
						exportUpdate = true
					}

					if val.Clients != exportRetrieved.Records[0].Rules[indx].Clients[0].Match {
						exportUpdate = true
					}
					if val.Protocols != exportRetrieved.Records[0].Rules[indx].Protocols[0] {
						exportUpdate = true
					}

					if val.Ro != exportRetrieved.Records[0].Rules[indx].RoRule[0] {
						exportUpdate = true
					}
					if val.Rw != exportRetrieved.Records[0].Rules[indx].RwRule[0] {
						exportUpdate = true
					}
					if val.Superuser != exportRetrieved.Records[0].Rules[indx].Superuser[0] {
						exportUpdate = true
					}
				}

			}

			if exportUpdate {
				// Build out complete export if there is any change
				idToReplace := exportRetrieved.Records[0].Id // get the id
				var newExport ontap.ExportPolicy

				//NOT CHANGING NAME DURING UPDATE
				//newExport.Name = svmCR.Spec.NfsConfig.Export.Name

				for _, val := range svmCR.Spec.NfsConfig.Export.Rules {
					var newRule ontap.ExportRule
					newRule.Anonuser = val.Anon
					newRule.Protocols = append(newRule.Protocols, val.Protocols)
					newRule.RwRule = append(newRule.RwRule, val.Rw)
					newRule.RoRule = append(newRule.RoRule, val.Ro)
					newRule.Superuser = append(newRule.Superuser, val.Superuser)
					var match ontap.ExportMatch
					match.Match = val.Clients
					newRule.Clients = append(newRule.Clients, match)
					newExport.Rules = append(newExport.Rules, newRule)
				}

				// otherwise changes need to be implemented
				if oc.Debug {
					log.Info("[DEBUG] NFS export update payload: " + fmt.Sprintf("%#v\n", newExport))
				}

				jsonPayload, err := json.Marshal(newExport)
				if err != nil {
					//error creating the json body
					log.Error(err, "Error creating the json payload for NFS export update - requeuing")
					_ = r.setConditionNfsExport(ctx, svmCR, CONDITION_STATUS_FALSE)
					r.Recorder.Event(svmCR, "Warning", "NfsUpdateExportFailed", "Error: "+err.Error())
					return err
				}

				err = oc.PatchNfsExport(idToReplace, jsonPayload)
				if err != nil {
					log.Error(err, "Error occurred when updating NFS export - requeuing")
					_ = r.setConditionNfsExport(ctx, svmCR, CONDITION_STATUS_FALSE)
					r.Recorder.Event(svmCR, "Warning", "NfsUpdateExportFailed", "Error: "+err.Error())
					return err
				}
				log.Info("NFS export updated successful")
				_ = r.setConditionNfsExport(ctx, svmCR, CONDITION_STATUS_TRUE)
				r.Recorder.Event(svmCR, "Normal", "NfsUpdateExportSucceeded", "Updated NFS export(s) successfully")
			} else {
				log.Info("No NFS export rules changed detected - skipping")
			}

		} //Check for NFS export updates or deletion of none conforming exports

	} // NFS exports rules defined in custom resource
	// END NFS EXPORTS

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
		return err
	}
	log.Info("NFS LIF creation attempt: " + lifToCreate.Name)
	err = oc.CreateIpInterface(jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when creating NFS LIF: "+lifToCreate.Name)
		return err
	}
	log.Info("NFS LIF creation successful: " + lifToCreate.Name)

	return nil
}

func CreateExport(exportToCreate gatewayv1alpha1.NfsExport, uuid string, oc *ontap.Client, log logr.Logger) (err error) {
	var newExport ontap.ExportPolicy
	newExport.Name = exportToCreate.Name

	for _, val := range exportToCreate.Rules {
		var newRule ontap.ExportRule
		newRule.Anonuser = val.Anon
		newRule.Protocols = append(newRule.Protocols, val.Protocols)
		newRule.RwRule = append(newRule.RwRule, val.Rw)
		newRule.RoRule = append(newRule.RoRule, val.Ro)
		newRule.Superuser = append(newRule.Superuser, val.Superuser)
		var match ontap.ExportMatch
		match.Match = val.Clients
		newRule.Clients = append(newRule.Clients, match)
		newExport.Rules = append(newExport.Rules, newRule)
	}

	newExport.Svm.Uuid = uuid

	jsonPayload, err := json.Marshal(newExport)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for NFS export creation: "+exportToCreate.Name)
		return err
	}
	log.Info("NFS export creation attempt: " + exportToCreate.Name)
	err = oc.CreateIpInterface(jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when creating NFS export: "+exportToCreate.Name)
		return err
	}
	log.Info("NFS export creation successful: " + exportToCreate.Name)

	return nil
}
