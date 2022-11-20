package controllers

import (
	"context"
	gatewayv1alpha2 "gateway/api/v1alpha2"
	"gateway/ontap"

	"github.com/go-logr/logr"
)

func (r *StorageVirtualMachineReconciler) reconcileIscsiUpdate(ctx context.Context, svmCR *gatewayv1alpha2.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {
	log.Info("STEP 14: Update iSCSI service")

	// iSCSI SERVICE
	// create := false
	// updateIscsiService := false

	// Check to see if iscsi configuration is provided in custom resource
	if svmCR.Spec.IscsiConfig == nil {
		// If not, exit with no error
		log.Info("No iSCSI service defined - skipping STEP 14")
		return nil
	}

	return nil
}
