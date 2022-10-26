package controllers

import (
	"context"
	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

func (r *StorageVirtualMachineReconciler) reconcileGetClient(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine,
	adminSecret *corev1.Secret, host string, debugOn bool, trustSSL bool,
	log logr.Logger) (*ontap.Client, error) {

	log.Info("Step 4: Create ONTAP client")

	oc, err := ontap.NewClient(
		string(adminSecret.Data["username"]),
		string(adminSecret.Data["password"]),
		host, debugOn, trustSSL)

	if err != nil {
		log.Error(err, "Error creating ONTAP client")
		_ = r.setConditionONTAPCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return oc, err
	}

	_ = r.setConditionONTAPCreation(ctx, svmCR, CONDITION_STATUS_TRUE)

	return oc, nil

}
