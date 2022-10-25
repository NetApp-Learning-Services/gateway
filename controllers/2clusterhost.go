package controllers

import (
	"context"
	"net"
	"net/url"

	gatewayv1alpha1 "gateway/api/v1alpha1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *StorageVirtualMachineReconciler) reconcileClusterHost(ctx context.Context,
	svmCR *gatewayv1alpha1.StorageVirtualMachine, log logr.Logger) (string, error) {

	log.Info("STEP 2: Identify Cluster Host")

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

		/*
			1.6667198119695444e+09 INFO STEP 2: Identify Cluster Host {"controller": "storagevirtualmachine", "controllerGroup": "gateway.netapp.com", "controllerKind": "StorageVirtualMachine", "storageVirtualMachine": {"name":"storagevirtualmachine-sample","namespace":"gateway-system"}, "namespace": "gateway-system", "name": "storagevirtualmachine-sample", "reconcileID": "d1369493-9ad0-4269-933d-fa7ae312a0f7", "Request.Namespace": "gateway-system", "Request.Name": "storagevirtualmachine-sample"}
			1.666719811969549e+09 INFO Using cluster management host: 192.168.0.102 {"controller": "storagevirtualmachine", "controllerGroup": "gateway.netapp.com", "controllerKind": "StorageVirtualMachine", "storageVirtualMachine": {"name":"storagevirtualmachine-sample","namespace":"gateway-system"}, "namespace": "gateway-system", "name": "storagevirtualmachine-sample", "reconcileID": "d1369493-9ad0-4269-933d-fa7ae312a0f7", "Request.Namespace": "gateway-system", "Request.Name": "storagevirtualmachine-sample"}
			1.6667198119721894e+09 ERROR Custom resource status update failed {"controller": "storagevirtualmachine", "controllerGroup": "gateway.netapp.com", "controllerKind": "StorageVirtualMachine", "storageVirtualMachine": {"name":"storagevirtualmachine-sample","namespace":"gateway-system"}, "namespace": "gateway-system", "name": "storagevirtualmachine-sample", "reconcileID": "d1369493-9ad0-4269-933d-fa7ae312a0f7", "error": "Operation cannot be fulfilled on storagevirtualmachines.gateway.netapp.com \"storagevirtualmachine-sample\": the object has been modified; please apply your changes to the latest version and try again"}
			gateway/controllers.(*StorageVirtualMachineReconciler).setConditionHostFound
			/workspace/controllers/conditions.go:50
			gateway/controllers.(*StorageVirtualMachineReconciler).reconcileClusterHost
			/workspace/controllers/2clusterhost.go:46
			gateway/controllers.(*StorageVirtualMachineReconciler).Reconcile
			/workspace/controllers/storagevirtualmachine_controller.go:69
			sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Reconcile
			/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.2/pkg/internal/controller/controller.go:121
			sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).reconcileHandler
			/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.2/pkg/internal/controller/controller.go:320
			sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).processNextWorkItem
			/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.2/pkg/internal/controller/controller.go:273
			sigs.k8s.io/controller-runtime/pkg/internal/controller.(*Controller).Start.func2.2
			/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.12.2/pkg/internal/controller/controller.go:234
		*/
		return name, nil //TODO: To prevent forever loop
	}

	return name, nil

}
