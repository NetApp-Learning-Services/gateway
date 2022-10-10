package controllers

import (
	"context"
	"net/url"

	gatewayv1alpha1 "gateway/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileClusterHost(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine) (string, error) {
	log := log.FromContext(ctx)

	// Get cluster management url
	u := svmCR.Spec.ClusterManagementHost
	var clusterUrl *url.URL
	if u == "" {
		err := errors.NewBadRequest("No Cluster Management LIF provided")
		log.Error(err, "The custom resource has no clusterHost")
		return clusterUrl.Host, err
	}

	clusterUrl, err := url.ParseRequestURI(u)
	if err != nil {
		log.Error(err, "clusterHost in the custom resource is invalid")
		u = ""
		return clusterUrl.Host, err
	}

	return clusterUrl.Host, err

}
