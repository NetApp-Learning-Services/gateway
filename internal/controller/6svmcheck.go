// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controller

import (
	"context"
	"strings"

	gateway "gateway/api/v1beta1"
	"gateway/internal/controller/ontap"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmCheck(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, oc *ontap.Client, log logr.Logger) (ontap.SvmByUUID, error) {

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

// STEP 6
// SVM Lookup
// Note: Status of SVM_FOUND can only be true, false, or unknown
const CONDITION_TYPE_SVM_FOUND = "6SVMDiscovered"
const CONDITION_REASON_SVM_FOUND = "SVMFound"
const CONDITION_MESSAGE_SVM_FOUND_TRUE = "UUID maps to SVM"
const CONDITION_MESSAGE_SVM_FOUND_FALSE = "NO UUID"
const CONDITION_MESSAGE_SVM_FOUND_UNKNOWN = "UUID does NOT map to SVM"

func (reconciler *StorageVirtualMachineReconciler) setConditionSVMFound(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(svmCR, CONDITION_REASON_SVM_FOUND) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_SVM_FOUND, CONDITION_REASON_SVM_FOUND)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_FOUND, status,
			CONDITION_REASON_SVM_FOUND, CONDITION_MESSAGE_SVM_FOUND_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_FOUND, status,
			CONDITION_REASON_SVM_FOUND, CONDITION_MESSAGE_SVM_FOUND_FALSE)
	}

	if status == CONDITION_STATUS_UNKNOWN {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_SVM_FOUND, status,
			CONDITION_REASON_SVM_FOUND, CONDITION_MESSAGE_SVM_FOUND_UNKNOWN)
	}
	return nil
}
