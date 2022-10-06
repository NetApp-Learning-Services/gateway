/*
Copyright 2022.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gatewayv1alpha1 "github.com/NetApp-Learning-Services/gateway/api/v1alpha1"
)

// StorageVirtualMachineReconciler reconciles a StorageVirtualMachine object
type StorageVirtualMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gateway.netapp.com,resources=storagevirtualmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.netapp.com,resources=storagevirtualmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.netapp.com,resources=storagevirtualmachines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// the StorageVirtualMachine object against the actual ontap cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *StorageVirtualMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO: Check out this: https://github.com/kubernetes-sigs/kubebuilder/issues/618
	// TODO: Check for existing of CRD object - if doesn't exist, log error, event, status, then exit reconcile

	// TODO: Check to see if SVM exists by the uuid in CR 
	// TODO: If SVM exists, then check to see if svmName needs to be updated
	// TODO: If SVM exists, then check to see if managment LIF needs to be created or updated
	// TODO: If SVM exists, then check to see if vsadmin needs to be created/updated
	// TODO: If SVM !exists, then create SVM and update CR with new uuid
	
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageVirtualMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha1.StorageVirtualMachine{}).
		Complete(r)
}
