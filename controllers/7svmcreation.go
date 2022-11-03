// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	defaultLog "log"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmCreation(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client, log logr.Logger) (ctrl.Result, error) {

	log.Info("Step 7: Create SVM")

	var payload ontap.SVMCreationPayload
	payload.Name = svmCR.Spec.SvmName
	payload.Comment = "Created by Astra Gateway"
	if svmCR.Spec.ManagementLIF != nil {
		var ifpayload ontap.IpInterfaceCreation
		ifpayload.Name = svmCR.Spec.ManagementLIF.Name
		ifpayload.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
		ifpayload.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
		ifpayload.ServicePolicy = "default-management" // special word
		//ifpayload.Scope = "svm"                        //special word

		var locpayload ontap.Location
		locpayload.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
		locpayload.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
		ifpayload.Location = locpayload
		payload.IpInterfaces = append(payload.IpInterfaces, ifpayload)
	}

	if oc.Debug {
		defaultLog.Printf("[DEBUG] SVM creation payload: " + fmt.Sprintf("%#v\n", payload))
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for SVM creation")
		_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, err
	}

	log.Info("SVM creation attempt")
	uuid, err := oc.CreateStorageVM(jsonPayload)
	if err != nil {
		log.Info("Uuid received was: " + uuid)
		log.Error(err, "Error occurred when creating SVM")
		_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
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
		_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, err
	}

	//Set condition for SVM create
	err = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_TRUE)
	if err != nil {
		return ctrl.Result{}, nil //even though condition not create, don't reconcile again
	}

	// Set finalizer
	_, err = r.addFinalizer(ctx, svmCR)
	if err != nil {
		return ctrl.Result{}, err //got another error - re-reconcile
	}

	return ctrl.Result{}, nil
}
