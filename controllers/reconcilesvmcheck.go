package controllers

import (
	"context"
	"strings"

	gatewayv1alpha1 "github.com/NetApp-Learning-Services/gateway/api/v1alpha1"
	"github.com/NetApp-Learning-Services/gateway/ontap"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmCheck(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, oc *ontap.Client) (*ontap.Svm, error) {
	log := log.FromContext(ctx)

	// Check to see if SVM exists by the uuid in CR
	uuid := strings.TrimSpace(svmCR.Spec.SvmUuid)
	if uuid != "" {
		// SvmUuid has a value
		// Check to see if SVM exists
		path := "/api/svm/svms/" + uuid
		parameters := []string{""} //[]string{"fields=ip_interfaces"}
		svm, _, err := oc.SvmGet(path, parameters)
		if err != nil {
			log.Error(err, "Invalid SVM UUID in custom resource")
			return nil, err
		}
		return svm, err
	}

	return nil, errors.NewNotFound(schema.GroupResource{Group: "gatewayv1alpha1", Resource: "StorageVirtualMachine"}, "svm")

}
