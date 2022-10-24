/*
Copyright 2022.
Created by Curtis Burchett
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	gatewayv1alpha1 "gateway/api/v1alpha1"
)

const (
	trustSSL = true
	debugOn  = true
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

	// STEP 1
	// Check for existing of CR object -
	// if doesn't exist or error retrieving, log error and exit reconcile
	// if discovered, write condition and move on
	svmCR, err := r.reconcileDiscoverObject(ctx, req, log)
	if err != nil && errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err //re-reconcile
	}

	// STEP 2
	// Get cluster management host
	host, err := r.reconcileClusterHost(ctx, svmCR, log)
	if err != nil {
		return ctrl.Result{}, nil // not a valid cluster Url - stop reconcile
	}

	// STEP 3
	// Look up cluster admin secret
	adminSecret, err := r.reconcileSecret(ctx,
		svmCR.Spec.ClusterCredentialSecret.Name,
		svmCR.Spec.ClusterCredentialSecret.Namespace, log)
	if err != nil {
		err = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, nil // not a valid secret - stop reconcile
	} else {
		err = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_TRUE)

	}

	// STEP 4
	//create ONTAP client
	oc, err := r.reconcileGetClient(ctx, svmCR, adminSecret, host, debugOn, trustSSL, log)
	if err != nil {
		return ctrl.Result{}, err //got another error - re-reconcile
	}

	// STEP 5
	// Check to see if deleting custom resource and handle the deletion
	isSMVMarkedToBeDeleted := svmCR.GetDeletionTimestamp() != nil
	if isSMVMarkedToBeDeleted {
		_, err = r.tryDeletions(ctx, svmCR, oc, log)
		if err != nil {
			log.Error(err, "Error during svmCR deletion")
			return ctrl.Result{}, err //got another error - re-reconcile
		} else {
			log.Info("SVM deleted, removed finalizer, cleaning up custom resource")
			return ctrl.Result{}, nil //stop reconcile
		}
	}

	// STEP 6
	// Check to see if svmCR has uuid and then check if svm can be looked up on that uuid
	create := false // Define variable whether to create svm or update it - default to false
	svmRetrieved, err := r.reconcileSvmCheck(ctx, svmCR, oc, log)
	if err != nil && errors.IsNotFound(err) {
		create = true
	} else {
		// some other error
		return ctrl.Result{}, err // got another error - re-reconcile
	}

	// Check whether we need to update or create an SVM
	if create == false {
		// STEP 7
		// reconcile SVM update
		log.Info("Reconciling SVM update")
		_, err = r.reconcileSvmUpdate(ctx, svmCR, svmRetrieved, oc, log)
		if err != nil {

		}

	} else {
		// STEP 8
		// reconcile SVM creation
		log.Info("Reconciling SVM creation")
		_, err = r.reconcileSvmCreation(ctx, svmCR, oc, log)
		if err != nil {
			log.Error(err, "Error during reconciling SVM creation")
			return ctrl.Result{}, err //got another error - re-reconcile
		}
	}

	// STEP 9
	//Check to see if need to create vsadmin
	if svmCR.Spec.VsadminCredentialSecret.Name != "" {
		// Look up vsadmin secret
		vsAdminSecret, err := r.reconcileSecret(ctx,
			svmCR.Spec.VsadminCredentialSecret.Name,
			svmCR.Spec.VsadminCredentialSecret.Namespace, log)
		if err != nil {
			// return ctrl.Result{}, nil // not a valid secret - ignore
			r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		} else {
			r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_TRUE)
			r.reconcileSecurityAccount(ctx, svmCR, oc, vsAdminSecret, log)
		}

	}

	return ctrl.Result{}, nil //no error - end reconcile
}

// SetupWithManager sets up the controller with the Manager.
func (r *StorageVirtualMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha1.StorageVirtualMachine{}).
		// Owns(&corev1.Secret{}).
		Complete(r)
}
