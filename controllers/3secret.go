// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/secret.go

package controllers

import (
	"context"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

func (r *StorageVirtualMachineReconciler) reconcileSecret(ctx context.Context,
	name string, namespace string, log logr.Logger) (*corev1.Secret, error) {

	secret := &corev1.Secret{}
	err := r.Get(ctx, types.NamespacedName{
		Name:      name,
		Namespace: namespace,
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

	//log.Info("username: " + string(secret.Data["username"]))
	//log.Info("password: " + string(secret.Data["password"]))

	return secret, nil
}
