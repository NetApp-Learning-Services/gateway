// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"strings"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmCheck(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client, log logr.Logger) (ontap.SvmByUUID, error) {

	log.Info("STEP 6: Check for a valid SVM")

	var svm ontap.SvmByUUID

	// Check to see if SVM exists by the uuid in CR
	uuid := strings.TrimSpace(svmCR.Spec.SvmUuid)
	if uuid == "" {
		log.Info("SVM uuid retrieved from the custom resource is empty, need to create the SVM")
		_ = r.setConditionSVMFound(ctx, svmCR, CONDITION_STATUS_FALSE)
		return svm, errors.NewNotFound(schema.GroupResource{Group: "gatewayv1alpha1", Resource: "StorageVirtualMachine"}, "svm")
	} else {
		log.Info("SVM uuid retrieved from the custom resource: " + uuid + ", attempt to get the SVM")

		// SvmUuid has a value in the custom resource
		// Check to see if SVM exists
		svm, err := oc.GetStorageVMByUUID(uuid)
		if err != nil {
			log.Error(err, "SVM uuid in the custom resource is invalid - not requeuing")
			_ = r.setConditionSVMFound(ctx, svmCR, CONDITION_STATUS_UNKNOWN)
			return svm, nil
		}
		log.Info("SVM uuid in the custom resource is valid", "svm retrieved: ", svm)
		_ = r.setConditionSVMFound(ctx, svmCR, CONDITION_STATUS_TRUE)
		return svm, nil
	}

}
