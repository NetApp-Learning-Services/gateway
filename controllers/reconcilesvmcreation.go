package controllers

import (
	"context"
	"encoding/json"

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

	var payload map[string]interface{}
	payload = make(map[string]interface{})
	payload["name"] = svmCR.Spec.SvmName
	payload["comment"] = "Created by Astra Gateway"
	if svmCR.Spec.ManagementLIF != nil {
		var ifpayload map[string]interface{}
		ifpayload = make(map[string]interface{})

		ifpayload["name"] = svmCR.Spec.ManagementLIF.Name
		ifpayload["ip"] = svmCR.Spec.ManagementLIF.IPAddress
		ifpayload["netmask"] = svmCR.Spec.ManagementLIF.Netmask
		payload["ip_interface"] = ifpayload

		var locpayload map[string]interface{}
		locpayload = make(map[string]interface{})
		locpayload["broadcast_domain"] = svmCR.Spec.ManagementLIF.BroacastDomain
		locpayload["home_node"] = svmCR.Spec.ManagementLIF.HomeNode
		locpayload["service_policy"] = "default-management" // special word
		payload["location"] = locpayload
	}
	log.Info("SVM creation payload", "payload:", payload)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		//error creating the json body
		return ctrl.Result{}, err
	}

	log.Info("SVM creation attempt")
	res, _ := oc.CreateStorageVM(jsonPayload)

	log.Info("SVM new uuid: " + res.Uuid)
	//patch the new uuid on the custom resource
	patch := client.MergeFrom(svmCR.DeepCopy())
	svmCR.Spec.SvmUuid = res.Uuid
	err = r.Patch(ctx, svmCR, patch)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
