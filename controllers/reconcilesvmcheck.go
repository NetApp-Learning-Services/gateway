package controllers

import (
	"context"
	"strings"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmCheck(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client) (ontap.Svm, error) {
	log := log.FromContext(ctx)
	log.Info("reconcileSvmCheck started")

	// Check to see if SVM exists by the uuid in CR
	uuid := strings.TrimSpace(svmCR.Spec.SvmUuid)
	if uuid == "" {
		log.Info("SVM uuid retrieved from the custom resource is empty, need to create the SVM")
	} else {
		log.Info("SVM uuid retrieved from the custom resource: " + uuid + ", attempt to get the SVM")
	}

	var svm ontap.Svm
	if uuid != "" {
		// SvmUuid has a value in the custom resource
		// Check to see if SVM exists
		svm, err := oc.GetStorageVMByUUID(uuid)
		if err != nil {
			log.Error(err, "SVM UUID in the custom resource is invalid")
			return svm, err
		}
		log.Info("SVM UUID in the custom resource is valid", "svm retrieved: ", svm)
		return svm, err
	}

	return svm, errors.NewNotFound(schema.GroupResource{Group: "gatewayv1alpha1", Resource: "StorageVirtualMachine"}, "svm")

}
