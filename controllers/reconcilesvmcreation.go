// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmCreation(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client) (ctrl.Result, error) {

	log := log.FromContext(ctx)
	log.Info("reconcileSvmCreation started")

	var payload ontap.SVMCreationPayload
	payload.Name = svmCR.Spec.SvmName
	payload.Comment = "Created by Astra Gateway"
	if svmCR.Spec.ManagementLIF != nil {
		var ifpayload ontap.IpInterface
		ifpayload.Name = svmCR.Spec.ManagementLIF.Name
		ifpayload.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
		ifpayload.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
		ifpayload.ServicePolicy = "default-management" // special word

		var locpayload ontap.Location
		locpayload.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
		locpayload.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
		ifpayload.Location = locpayload
		payload.IpInterfaces = append(payload.IpInterfaces, ifpayload)
	}
	//log.Info("SVM creation payload", "payload:", payload)
	log.Info("SVM creation payload: " + fmt.Sprintf("%#v\n", payload))

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for SVM creation")
		return ctrl.Result{}, err
	}

	log.Info("SVM creation attempt")
	uuid, err := oc.CreateStorageVM(jsonPayload)
	if err != nil {
		log.Info("uuid received was: " + uuid)
		log.Error(err, "Error occurred when creating SVM")
		return ctrl.Result{}, err
	}

	// log.Info("Looking up UUID for the new SVM")
	// uuid, err := oc.GetStorageVmUUIDByName(svmCR.Spec.SvmName)
	// if err != nil {
	// 	log.Error(err, "Error occurred when creating SVM")
	// 	return ctrl.Result{}, err
	// }

	log.Info("SVM new uuid: " + uuid)
	//patch the new uuid on the custom resource
	patch := client.MergeFrom(svmCR.DeepCopy())
	svmCR.Spec.SvmUuid = uuid
	err = r.Patch(ctx, svmCR, patch)
	if err != nil {
		log.Error(err, "Error patching the new uuid in the custom resource")
		return ctrl.Result{}, err
	}

	//Check to see if need to create vsadmin
	if svmCR.Spec.VsadminCredentialSecret.Name != "" {
		// Look up vsadmin secret
		vsAdminSecret, err := r.reconcileSecret(ctx,
			svmCR.Spec.VsadminCredentialSecret.Name,
			svmCR.Spec.VsadminCredentialSecret.Namespace)
		if err != nil {
			// return ctrl.Result{}, nil // not a valid secret - ignore
		} else {
			r.reconcileSecurityAccount(ctx, svmCR, oc, vsAdminSecret)
		}

	}

	return ctrl.Result{}, nil
}
