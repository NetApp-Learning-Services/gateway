package controller

import (
	"context"
	"net"
	"net/url"

	gateway "gateway/api/v1beta1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *StorageVirtualMachineReconciler) reconcileClusterHost(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, log logr.Logger) (string, error) {

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

// STEP 2
// Host name discovery
// Note: Status of HOST_FOUND can only be true, false, unknown
const CONDITION_TYPE_HOST_FOUND = "2HostDiscovered"
const CONDITION_REASON_HOST_FOUND = "HostFound"
const CONDITION_MESSAGE_HOST_FOUND_TRUE = "A valid host found"
const CONDITION_MESSAGE_HOST_FOUND_FALSE = "A valid host was not found"

func (reconciler *StorageVirtualMachineReconciler) setConditionHostFound(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	if reconciler.containsCondition(svmCR, CONDITION_REASON_HOST_FOUND) {
		reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_HOST_FOUND, CONDITION_REASON_HOST_FOUND)
	}

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_HOST_FOUND, status,
			CONDITION_REASON_HOST_FOUND, CONDITION_MESSAGE_HOST_FOUND_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_HOST_FOUND, status,
			CONDITION_REASON_HOST_FOUND, CONDITION_MESSAGE_HOST_FOUND_FALSE)
	}

	if status == CONDITION_STATUS_UNKNOWN {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_HOST_FOUND, status,
			CONDITION_REASON_HOST_FOUND, CONDITION_MESSAGE_HOST_FOUND_FALSE)
	}

	return nil
}
