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
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const defaultComment = "Created by Astra Gateway"       //magic word
const managementLIFServicePolicy = "default-management" //magic word

func (r *StorageVirtualMachineReconciler) reconcileSvmCreation(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client, log logr.Logger) (ctrl.Result, error) {

	log.Info("STEP 7: Create SVM")

	var payload ontap.SVMCreationPayload
	payload.Name = svmCR.Spec.SvmName
	payload.Comment = defaultComment
	if svmCR.Spec.ManagementLIF != nil {
		var ifpayload ontap.IpInterfaceCreation
		ifpayload.Name = svmCR.Spec.ManagementLIF.Name
		ifpayload.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
		ifpayload.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
		ifpayload.ServicePolicy = managementLIFServicePolicy

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
		log.Error(err, "Error creating the json payload for SVM creation - requeuing")
		r.Recorder.Event(svmCR, "Warning", "SvmCreationFailed", "Error: "+err.Error())
		_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, err
	}

	log.Info("SVM creation attempt")
	uuid, err := oc.CreateStorageVM(jsonPayload)
	if err != nil {
		log.Info("Uuid received was: " + uuid)
		log.Error(err, "Error occurred when creating SVM - requeuing")
		r.Recorder.Event(svmCR, "Warning", "SvmCreationFailed", "Error: "+err.Error())
		r.Recorder.Event(svmCR, "Warning", "SvmCreationFailed", "Error: "+err.Error())
		_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, err
	}

	log.Info("SVM new uuid: " + uuid)
	//patch the new uuid on the custom resource
	patch := client.MergeFrom(svmCR.DeepCopy())
	svmCR.Spec.SvmUuid = uuid
	err = r.Patch(ctx, svmCR, patch)
	if err != nil {
		log.Error(err, "Error patching the new uuid in the custom resource - requeuing")
		r.Recorder.Event(svmCR, "Warning", "SvmCreationFailed", "Error: "+err.Error())
		_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, err
	}

	//Set condition for SVM create
	_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_TRUE)

	// Set finalizer
	_, err = r.addFinalizer(ctx, svmCR)
	if err != nil {
		log.Error(err, "Error adding the finalizer to the custom resource - requeuing")
		r.Recorder.Event(svmCR, "Warning", "SvmCreationFailed", "Error: "+err.Error())
		return ctrl.Result{}, err //got another error - re-reconcile
	}
	r.Recorder.Event(svmCR, "Normal", "SvmCreationSuccesful", "SVM created with UUID: "+uuid)
	log.Info("SVM created")
	return ctrl.Result{}, nil
}

func (r *StorageVirtualMachineReconciler) addFinalizer(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine) (ctrl.Result, error) {
	if !controllerutil.ContainsFinalizer(svmCR, finalizerName) {
		controllerutil.AddFinalizer(svmCR, finalizerName)
		err := r.Update(ctx, svmCR)
		if err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}
