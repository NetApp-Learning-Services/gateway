package controllers

import (
	"context"
	"net/url"

	gatewayv1alpha1 "gateway/api/v1alpha1"

	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (r *StorageVirtualMachineReconciler) reconcileClusterUrl(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine) (*url.URL, error) {
	log := log.FromContext(ctx)

	// Get cluster management url
	u := svmCR.Spec.ClusterManagementUrl
	var clusterUrl *url.URL
	if u == "" {
		err := errors.NewBadRequest("No Cluster Management LIF provided")
		log.Error(err, "Custom resource has no clusterUrl")
		return clusterUrl, err
	}

	clusterUrl, err := url.ParseRequestURI(u)
	if err != nil {
		log.Error(err, "clusterUrl in custom resource is invalid")
		u = ""
		return clusterUrl, err
	}

	return clusterUrl, err

}
