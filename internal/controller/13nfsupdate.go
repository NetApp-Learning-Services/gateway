package controller

import (
	"context"
	"encoding/json"
	"fmt"
	gateway "gateway/api/v1beta2"
	"gateway/internal/controller/ontap"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const NfsLifType = "default-data-files" //magic word
const NfsLifScope = "svm"               //magic word

func (r *StorageVirtualMachineReconciler) reconcileNfsUpdate(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, uuid string, oc *ontap.Client, log logr.Logger) error {

	log.Info("STEP 13: Update NFS service")

	// NFS SERVICE

	createNfsService := false
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
		createNfsService = true
	} else if err != nil {
		// some other error
		log.Error(err, "Error retrieving NFS service for SVM by UUID - requeuing")
		return err
	}

	var upsertNfsService ontap.NFSService

	if createNfsService {
		log.Info("No NFS service defined for SVM: " + uuid + " - creating NFS service")
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

		if oc.Debug {
			log.Info("[DEBUG] NFS service creation payload: " + fmt.Sprintf("%#v\n", upsertNfsService))
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

		if oc.Debug && updateNfsService {
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
		createNfsLifs := false

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
			createNfsLifs = true
		}

		if createNfsLifs {
			//creating lifs
			for _, val := range svmCR.Spec.NfsConfig.Lifs {
				err = CreateLif(val, NfsLifType, uuid, oc, log)
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
					err = CreateLif(val, NfsLifType, uuid, oc, log)
					if err != nil {
						_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "NfsCreationLifFailed", "Error: "+err.Error())
						return err
					}

				} else {

					if reflect.ValueOf(lifs.Records[index]).IsZero() {
						break
					}

					err = UpdateLif(val, lifs.Records[index], NfsLifType, oc, log)
					if err != nil {
						_ = r.setConditionNfsLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "NfsUpdateLifFailed", "Error: "+err.Error())
						//e := err.(*apiError)
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
		createNfsExports := false

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
			createNfsExports = true
		}

		if createNfsExports {
			// this will probably never happen because there is always a default export
			// creating export
			err = CreateNfsExport(*svmCR.Spec.NfsConfig.Export, uuid, oc, log)
			if err != nil {
				_ = r.setConditionNfsExport(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}

		} else {

			// if more than one export, delete anything after the first one
			// never delete the first export
			for i := 1; i < exportRetrieved.NumRecords; i++ {
				log.Info("NFS export delete attempt: " + exportRetrieved.Records[i].Name)
				err = oc.DeleteNfsExport(exportRetrieved.Records[i].Id)
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
			updateNfsExports := false

			if exportRetrieved.Records[0].Name != svmCR.Spec.NfsConfig.Export.Name {
				//need to update it
				updateNfsExports = true
			}

			for indx, val := range svmCR.Spec.NfsConfig.Export.Rules {

				if len(exportRetrieved.Records[0].Rules) == 0 {
					//nothing defined
					updateNfsExports = true
				} else {
					if val.Anon != exportRetrieved.Records[0].Rules[indx].Anonuser {
						updateNfsExports = true
					}

					if val.Clients != exportRetrieved.Records[0].Rules[indx].Clients[0].Match {
						updateNfsExports = true
					}
					if val.Protocols != exportRetrieved.Records[0].Rules[indx].Protocols[0] {
						updateNfsExports = true
					}

					if val.Ro != exportRetrieved.Records[0].Rules[indx].RoRule[0] {
						updateNfsExports = true
					}
					if val.Rw != exportRetrieved.Records[0].Rules[indx].RwRule[0] {
						updateNfsExports = true
					}
					if val.Superuser != exportRetrieved.Records[0].Rules[indx].Superuser[0] {
						updateNfsExports = true
					}
				}

			}

			if updateNfsExports {
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

func CreateNfsExport(exportToCreate gateway.NfsExport, uuid string, oc *ontap.Client, log logr.Logger) (err error) {
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

// STEP 13
// NFS update
// Note: Status of NFS_SERVICE can only be true or false
const CONDITION_TYPE_NFS_SERVICE = "13NFSservice"
const CONDITION_REASON_NFS_SERVICE = "NFSservice"
const CONDITION_MESSAGE_NFS_SERVICE_TRUE = "NFS service configuration succeeded"
const CONDITION_MESSAGE_NFS_SERVICE_FALSE = "NFS service configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNfsService(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NFS_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NFS_SERVICE, CONDITION_REASON_NFS_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_SERVICE, CONDITION_MESSAGE_NFS_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_SERVICE, CONDITION_MESSAGE_NFS_SERVICE_FALSE)
	}
	return nil
}

const CONDITION_REASON_NFS_LIF = "NFSlif"
const CONDITION_MESSAGE_NFS_LIF_TRUE = "NFS LIF configuration succeeded"
const CONDITION_MESSAGE_NFS_LIF_FALSE = "NFS LIF configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNfsLif(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NFS_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NFS_SERVICE, CONDITION_REASON_NFS_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_LIF, CONDITION_MESSAGE_NFS_LIF_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_LIF, CONDITION_MESSAGE_NFS_LIF_FALSE)
	}
	return nil
}

const CONDITION_REASON_NFS_EXPORT = "NFSexport"
const CONDITION_MESSAGE_NFS_EXPORT_TRUE = "NFS export configuration succeeded"
const CONDITION_MESSAGE_NFS_EXPORT_FALSE = "NFS export configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionNfsExport(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_NFS_EXPORT) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_NFS_SERVICE, CONDITION_REASON_NFS_EXPORT)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_EXPORT, CONDITION_MESSAGE_NFS_EXPORT_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_NFS_SERVICE, status,
			CONDITION_REASON_NFS_EXPORT, CONDITION_MESSAGE_NFS_EXPORT_FALSE)
	}
	return nil
}
