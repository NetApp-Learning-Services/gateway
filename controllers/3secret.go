// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/secret.go

package controllers

import (
	"context"
	gatewayv1alpha2 "gateway/api/v1alpha2"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const clusterAdminRequest = "ontap-cluster-admin" // magic word
const svmAdminRequest = "ontap-svm-admin"         // magic word

func (r *StorageVirtualMachineReconciler) reconcileSecret(ctx context.Context, secretType string,
	name string, namespace string, svmCR *gatewayv1alpha2.StorageVirtualMachine, log logr.Logger) (*corev1.Secret, error) {

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
		return nil, nil
	} else if err != nil {
		log.Error(err, "Failed to get secret - not requeuing")
		if secretType == clusterAdminRequest {
			_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		} else if secretType == svmAdminRequest {
			_ = r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		}
		return nil, nil
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
