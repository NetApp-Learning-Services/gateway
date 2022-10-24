package controllers

import (
	"context"
	"net/url"

	gatewayv1alpha1 "gateway/api/v1alpha1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *StorageVirtualMachineReconciler) reconcileClusterHost(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, log logr.Logger) (string, error) {

	// Get cluster management url
	host := svmCR.Spec.ClusterManagementHost
	var clusterUrl *url.URL
	if host == "" {
		err := errors.NewBadRequest("No Cluster Management LIF provided")
		log.Error(err, "The custom resource has no clusterHost")
		_ = r.setConditionHostFound(ctx, svmCR, CONDITION_STATUS_FALSE)
		return clusterUrl.Host, err
	}

	clusterUrl, err := url.ParseRequestURI(host)
	if err != nil {
		log.Error(err, "clusterHost in the custom resource is invalid")
		_ = r.setConditionHostFound(ctx, svmCR, CONDITION_STATUS_UNKNOWN)
		return clusterUrl.Host, err
	}

	log.Info("Using cluster management host: " + clusterUrl.Host)

	//Set condition for CR found
	err = r.setConditionHostFound(ctx, svmCR, CONDITION_STATUS_TRUE)
	if err != nil {
		return clusterUrl.Host, err
	}

	return clusterUrl.Host, err

}
