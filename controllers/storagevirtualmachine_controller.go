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
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"
)

// StorageVirtualMachineReconciler reconciles a StorageVirtualMachine object
type StorageVirtualMachineReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gateway.netapp.com,resources=storagevirtualmachines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gateway.netapp.com,resources=storagevirtualmachines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gateway.netapp.com,resources=storagevirtualmachines/finalizers,verbs=update

// ADDED to support access to secrets
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.

// the StorageVirtualMachine object against the actual ontap cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *StorageVirtualMachineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Create log from the context
	log := log.FromContext(ctx).WithValues("Request.Namespace", req.Namespace, "Request.Name", req.Name)
	log.Info("Reconcile started")

	// TODO: Check out this: https://github.com/kubernetes-sigs/kubebuilder/issues/618

	// Check for existing of CR object -
	// if doesn't exist or error retrieving, log error and exit reconcile
	svmCR := &gatewayv1alpha1.StorageVirtualMachine{}
	err := r.Get(ctx, req.NamespacedName, svmCR)
	if err != nil && errors.IsNotFound(err) {
		log.Info("StorageVirtualMachine custom resource not found. Ignoring since object must be deleted.")
		return ctrl.Result{}, nil
	} else if err != nil {
		log.Error(err, "Failed to get StorageVirtualMachine custom resource. Re-running reconile.")
		return ctrl.Result{}, err
	}

	//Set condition for CR found
	err = r.setConditionResourceFound(ctx, svmCR)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Get cluster management url
	clusterUrl, err := r.reconcileClusterUrl(ctx, svmCR)
	if err != nil {
		return ctrl.Result{}, nil // not a valid cluster Url - stop reconcile
	}

	log.Info("Using cluster management URL: " + clusterUrl.String())

	// Look up adminSecret
	adminSecret, err := r.reconcileSecret(ctx, svmCR)
	if err != nil {
		return ctrl.Result{}, nil // not a valid secret - stop reconcile
	}

	log.Info("Secret username: " + string(adminSecret.Data["username"]))
	log.Info("Secret password: " + string(adminSecret.Data["password"]))

	//create ONTAP client
	oc := ontap.NewClient(
		clusterUrl.String(),
		&ontap.ClientOptions{
			BasicAuthUser:     string(adminSecret.Data["username"]),
			BasicAuthPassword: string(adminSecret.Data["password"]),
			SSLVerify:         false,
			Debug:             true,
			Timeout:           60 * time.Second,
		},
	)

	//define variable whether to create svm or update it - default to false
	create := false

	// Check to see if svmCR has uuid and then check if svm can be looked up on that uuid
	svm, err := r.reconcileSvmCheck(ctx, svmCR, oc)
	if err != nil && errors.IsNotFound(err) {
		create = true
	} else {
		// some other error
		return ctrl.Result{}, err // got another error - re-reconcile
	}

	if create == false {
		log.Info("Reconciling SVM update")
		log.Info("create == false: ", svm)
		// reconcile SVM update

	} else {
		// reconcile SVM creation
		log.Info("Reconciling SVM creation")
		_, err = r.reconcileSvmCreation(ctx, svmCR, oc)
		if err != nil {
			log.Error(err, "Error during reconciling SVM creation")
			return ctrl.Result{}, err //got another error - re-reconcile
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageVirtualMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha1.StorageVirtualMachine{}).
		// Owns(&corev1.Secret{}).
		Complete(r)
}
