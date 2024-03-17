// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/secret.go

package controller

import (
	"context"
	gateway "gateway/api/v1beta1"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const clusterAdminRequest = "ontap-cluster-admin" // magic word
const svmAdminRequest = "ontap-svm-admin"         // magic word

func (r *StorageVirtualMachineReconciler) reconcileSecret(ctx context.Context, secretType string,
	name string, namespace string, svmCR *gateway.StorageVirtualMachine, log logr.Logger) (*corev1.Secret, error) {

	if secretType == clusterAdminRequest {
		log.Info("STEP 3: Resolve cluster admin secret")
	} else if secretType == svmAdminRequest {
		log.Info("STEP 8: Resolve SVM management secret")
	} else {
		log.Info("STEP ?: Resolve ? secret")
		return nil, nil
	}

	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
	}, secret)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Secret does not exist - not requeuing")
		if secretType == clusterAdminRequest {
			_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		} else if secretType == svmAdminRequest {
			_ = r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		}
		return nil, err
	} else if err != nil {
		log.Error(err, "Failed to get secret - not requeuing")
		if secretType == clusterAdminRequest {
			_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		} else if secretType == svmAdminRequest {
			_ = r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		}
		return nil, err
	}

	if strings.TrimSpace(string(secret.Data["username"])) == "" {

		log.Error(errors.NewBadRequest("Missing username"), secret.Name+"has no username - not requeuing")
		if secretType == clusterAdminRequest {
			_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		} else if secretType == svmAdminRequest {
			_ = r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		}
		return nil, nil
	}

	if strings.TrimSpace(string(secret.Data["password"])) == "" {
		log.Error(errors.NewBadRequest("Missing password"), secret.Name+"has no password - not requeuing")
		if secretType == clusterAdminRequest {
			_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		} else if secretType == svmAdminRequest {
			_ = r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		}
		return nil, nil
	}

	//log.Info("username: " + string(secret.Data["username"]))
	//log.Info("password: " + string(secret.Data["password"]))

	if secretType == clusterAdminRequest {
		log.Info("Cluster admin credentials available")
		_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_TRUE)
	} else if secretType == svmAdminRequest {
		log.Info("SVM managment credentials available")
		_ = r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_TRUE)
	}

	return secret, nil
}

// STEP 3
// Resolve Secret
// Note: Status of CLUSTER_SECRET_LOOKUP can only be true or false
const CONDITION_TYPE_CLUSTER_SECRET_LOOKUP = "3ClusterAdminSecretLookup"
const CONDITION_REASON_CLUSTER_SECRET_LOOKUP = "ClusterAdminSecretLookup"
const CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_TRUE = "Cluster Admin credentials available"
const CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_FALSE = "Cluster Admin credentials NOT available"

func (reconciler *StorageVirtualMachineReconciler) setConditionClusterSecretLookup(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(svmCR, CONDITION_REASON_CLUSTER_SECRET_LOOKUP) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_CLUSTER_SECRET_LOOKUP, CONDITION_REASON_CLUSTER_SECRET_LOOKUP)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_CLUSTER_SECRET_LOOKUP, status,
			CONDITION_REASON_CLUSTER_SECRET_LOOKUP, CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_CLUSTER_SECRET_LOOKUP, status,
			CONDITION_REASON_CLUSTER_SECRET_LOOKUP, CONDITION_MESSAGE_CLUSTER_SECRET_LOOKUP_FALSE)
	}
	return nil
}

// STEP 8
// VSADMIN LOOKUP
// Note: Status of VSADMIN_SECRET_LOOKUP can only be true or false
const CONDITION_TYPE_VSADMIN_SECRET_LOOKUP = "8VsAdminSecretLookup"
const CONDITION_REASON_VSADMIN_SECRET_LOOKUP = "VsAdminSecretLookup"
const CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_TRUE = "SVM Admin credentials available"
const CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_FALSE = "SVM Admin credentials NOT available"

func (reconciler *StorageVirtualMachineReconciler) setConditionVsadminSecretLookup(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(svmCR, CONDITION_REASON_VSADMIN_SECRET_LOOKUP) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_VSADMIN_SECRET_LOOKUP, CONDITION_REASON_VSADMIN_SECRET_LOOKUP)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_VSADMIN_SECRET_LOOKUP, status,
			CONDITION_REASON_VSADMIN_SECRET_LOOKUP, CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_VSADMIN_SECRET_LOOKUP, status,
			CONDITION_REASON_VSADMIN_SECRET_LOOKUP, CONDITION_MESSAGE_VSADMIN_SECRET_LOOKUP_FALSE)
	}
	return nil
}
