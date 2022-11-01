package controllers

import (
	"context"
	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"github.com/go-logr/logr"
)

func (r *StorageVirtualMachineReconciler) reconcileNFSUpdate(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {

	// Check to see if nfs configuration is provided in custom resource

	// If not, exit with no error

	// Get the NFS configuration of SVM
	// GET /procotols/nfs/services/uuid

	// Compare enabled to custom resource enabled
	// If not and enabled in custom resource, POST /procotols/nfs/services

	// Compare v3, v4, v41 to custom resource enabled
	// If not and enabled in custom resource, PATCH /procotols/nfs/services

	// Check to see if NFS interfaces are defined in custom resource

	// If so, check to see if NFS interfaces defined and compare to custom resource's definitions

	// Check to see if NFS rules are defined in custom resources

	// If so, GET /protocols/nfs/export-policies compare rules with result based upon index/id
	// PATCH /protocols/nfs/export-policies/id if needed
	// If rule missing in ONTAP, POST /protocols/nfs/export-policies/

	return nil
}
