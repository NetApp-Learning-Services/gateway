package controller

import (
	"context"
	"net"
	"net/url"

	gatewayv1alpha2 "gateway/api/v1alpha2"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *StorageVirtualMachineReconciler) reconcileClusterHost(ctx context.Context,
	svmCR *gatewayv1alpha2.StorageVirtualMachine, log logr.Logger) (string, error) {

	log.Info("STEP 2: Identify cluster host")

	// Get cluster management url
	host := svmCR.Spec.ClusterManagementHost
	name := ""
	if host == "" {
		err := errors.NewBadRequest("No Cluster Management LIF provided")
		log.Error(err, "The custom resource has no clusterHost")
		_ = r.setConditionHostFound(ctx, svmCR, CONDITION_STATUS_FALSE)
		return host, err
	}

	addr := net.ParseIP(host)
	if addr == nil {
		log.Info("clusterHost was not a IP address")
		clusterUrl, err := url.Parse(host)
		if err != nil {
			log.Error(err, "clusterHost in the custom resource is invalid")
			_ = r.setConditionHostFound(ctx, svmCR, CONDITION_STATUS_UNKNOWN)
			return clusterUrl.Host, err
		}
		name = clusterUrl.Host
	} else {
		name = addr.String()
	}

	log.Info("Using cluster management host: " + name)

	//Set condition for CR found
	err := r.setConditionHostFound(ctx, svmCR, CONDITION_STATUS_TRUE)
	if err != nil {
		return name, nil
	}

	return name, nil

}
