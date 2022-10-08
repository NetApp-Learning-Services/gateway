package controllers

import (
	"context"
	"strings"

	gatewayv1alpha1 "gateway/api/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileSecret(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine) (*corev1.Secret, error) {
	log := log.FromContext(ctx)
	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      svmCR.Spec.ClusterCredentialSecret.Name,
		Namespace: svmCR.Spec.ClusterCredentialSecret.Namespace,
	}, secret)
	if err != nil && errors.IsNotFound(err) {
		log.Error(err, "Secret does not exist")
		return nil, nil
	} else if err != nil {
		log.Error(err, "Failed to get secret ")
		return nil, err
	}

	if strings.TrimSpace(string(secret.Data["username"])) == "" {
		log.Error(errors.NewBadRequest("Missing username"), secret.Name+"has no username")
		return nil, nil
	}

	if strings.TrimSpace(string(secret.Data["password"])) == "" {
		log.Error(errors.NewBadRequest("Missing password"), secret.Name+"has no password")
		return nil, nil
	}

	return secret, nil
}
