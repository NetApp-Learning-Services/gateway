package controller

import (
	"context"
	gateway "gateway/api/v1beta2"
	"gateway/internal/controller/ontap"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *StorageVirtualMachineReconciler) reconcileGetClient(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine,
	adminSecret *corev1.Secret, host string, trustSSL bool,
	log logr.Logger) (*ontap.Client, error) {

	log.Info("STEP 4: Create ONTAP client")

	oc, err := ontap.NewClient(
		string(adminSecret.Data["username"]),
		string(adminSecret.Data["password"]),
		host, svmCR.Spec.SvmDebug, trustSSL)

	if err != nil {
		log.Error(err, "Error creating ONTAP client - requeueing")
		_ = r.setConditionONTAPCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return oc, err
	}

	log.Info("ONTAP client created")
	_ = r.setConditionONTAPCreation(ctx, svmCR, CONDITION_STATUS_TRUE)

	cluster, err := oc.GetCluster()
	if err != nil {
		log.Error(err, "Error retrieving cluster - requeuing")
		return oc, err
	}

	log.Info("Connected to cluster: " + host)
	log.Info("Cluster reporting ONTAP version: " + cluster.Version.Full)

	return oc, nil

}

// STEP 4
// ONTAP client Creation
// Note: Status of ONTAP_CREATED can only be true or false
const CONDITION_TYPE_ONTAP_CREATED = "4CreatedONTAPClient"
const CONDITION_REASON_ONTAP_CREATED = "ONTAPClientCreation"
const CONDITION_MESSAGE_ONTAP_CREATED_TRUE = "ONTAP client created"
const CONDITION_MESSAGE_ONTAP_CREATED_FALSE = "ONTAP client failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionONTAPCreation(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(svmCR, CONDITION_REASON_ONTAP_CREATED) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_ONTAP_CREATED, CONDITION_REASON_ONTAP_CREATED)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ONTAP_CREATED, status,
			CONDITION_REASON_ONTAP_CREATED, CONDITION_MESSAGE_ONTAP_CREATED_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_ONTAP_CREATED, status,
			CONDITION_REASON_ONTAP_CREATED, CONDITION_MESSAGE_ONTAP_CREATED_FALSE)
	}
	return nil
}
