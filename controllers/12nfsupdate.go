package controllers

import (
	"context"
	"encoding/json"
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

		upsertNfsService.Enabled = svmCR.Spec.NfsConfig.NfsEnabled
		upsertNfsService.Protocol.V3Enable = svmCR.Spec.NfsConfig.Nfsv3
		upsertNfsService.Protocol.V4Enable = svmCR.Spec.NfsConfig.Nfsv4
		upsertNfsService.Protocol.V41Enable = svmCR.Spec.NfsConfig.Nfsv41
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
		if nfsService.Enabled != svmCR.Spec.NfsConfig.NfsEnabled {
			updateNfsService = true
			upsertNfsService.Enabled = svmCR.Spec.NfsConfig.NfsEnabled
		}

		if nfsService.Protocol.V3Enable != svmCR.Spec.NfsConfig.Nfsv3 {
			updateNfsService = true
			upsertNfsService.Protocol.V3Enable = svmCR.Spec.NfsConfig.Nfsv3
		}

		if nfsService.Protocol.V4Enable != svmCR.Spec.NfsConfig.Nfsv4 {
			updateNfsService = true
			upsertNfsService.Protocol.V4Enable = svmCR.Spec.NfsConfig.Nfsv4
		}

		if nfsService.Protocol.V41Enable != svmCR.Spec.NfsConfig.Nfsv41 {
			updateNfsService = true
			upsertNfsService.Protocol.V41Enable = svmCR.Spec.NfsConfig.Nfsv41
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
			err = oc.PatchNfsService(uuid, jsonPayload)
			if err != nil {
				log.Error(err, "Error updating the NFS service")
				//TODO: _ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
				return err
			}
			log.Info("NFS service updated successful")
		} else {
			log.Info("No changes detected - NFS service")
		}
	}

	// Check to see if NFS interfaces are defined in custom resource

	// If so, check to see if NFS interfaces defined and compare to custom resource's definitions

	// Check to see if NFS rules are defined in custom resources

	// If so, GET /protocols/nfs/export-policies compare rules with result based upon index/id
	// PATCH /protocols/nfs/export-policies/id if needed
	// If rule missing in ONTAP, POST /protocols/nfs/export-policies/

	return nil
}
