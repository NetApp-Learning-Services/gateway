/*
Copyright 2022.
Created by Curtis Burchett
*/

package controllers

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

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
	log.Info("RECONCILE START")

	// This works.
	// It is a hack to stop the second reconcile that occurrs
	// immediately after the first reconcile.
	// If this is not present it causes errors while updating conditions.
	// TODO: Check this statement: https://groups.google.com/g/kubebuilder/c/tULj-TRM9ts
	time.Sleep(1 * time.Second)

	// STEP 1
	// Check for existing of CR object -
	// if doesn't exist or error retrieving, log error and exit reconcile
	// if discovered, write condition and move on
	svmCR, err := r.reconcileDiscoverObject(ctx, req, log)
	if err != nil && errors.IsNotFound(err) {
		return ctrl.Result{Requeue: false}, nil
	} else if err != nil {
		log.Error(err, "Error during custom resource discovery - requeuing")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err //re-reconcile
	}

	// STEP 2
	// Get cluster management host
	host, err := r.reconcileClusterHost(ctx, svmCR, log)
	if err != nil {
		return ctrl.Result{Requeue: false}, nil // not a valid cluster Url - stop reconcile
	}

	// STEP 3
	// Look up cluster admin secret
	adminSecret, err := r.reconcileSecret(ctx,
		svmCR.Spec.ClusterCredentialSecret.Name,
		svmCR.Spec.ClusterCredentialSecret.Namespace, log)
	if err != nil {
		log.Info("Cluster admin credentials NOT available")
		_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{Requeue: false}, nil // not a valid secret - stop reconcile
	} else {
		log.Info("Cluster admin credentials available")
		_ = r.setConditionClusterSecretLookup(ctx, svmCR, CONDITION_STATUS_TRUE)
	}

	// STEP 4
	//create ONTAP client
	oc, err := r.reconcileGetClient(ctx, svmCR, adminSecret, host, trustSSL, log)
	if err != nil {
		log.Error(err, "Error during creation of ONTAP client - requeuing")
		return ctrl.Result{RequeueAfter: 30 * time.Second}, err //got another error - re-reconcile
	} else {
		log.Info("ONTAP client created")
	}

	// STEP 5
	// Check to see if deleting custom resource and handle the deletion
	isSMVMarkedToBeDeleted := svmCR.GetDeletionTimestamp() != nil
	if isSMVMarkedToBeDeleted {
		_, err = r.tryDeletions(ctx, svmCR, oc, log)
		if err != nil {
			log.Error(err, "Error during custom resource deletion - requeuing")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err //got another error - re-reconcile
		} else {
			log.Info("SVM deleted, removed finalizer, cleaning up custom resource")
			return ctrl.Result{Requeue: false}, nil //stop reconcile
		}
	}

	// STEP 6
	// Check to see if svmCR has uuid and then check if svm can be looked up on that uuid
	create := false // Define variable whether to create svm or update it - default to false
	svmRetrieved, err := r.reconcileSvmCheck(ctx, svmCR, oc, log)
	if err != nil {
		if errors.IsNotFound(err) {
			create = true
		} else {
			// some other error
			log.Error(err, "Some other error while trying to retrieve SVM, other then SVM not created - requeuing")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err // got another error - re-reconcile
		}
	}

	if create {
		// STEP 7
		// reconcile SVM creation
		_, err = r.reconcileSvmCreation(ctx, svmCR, oc, log)
		if err != nil {
			log.Error(err, "Error during reconciling SVM creation - requeuing")
			return ctrl.Result{RequeueAfter: 30 * time.Second}, err //got another error - re-reconcile
		} else {
			log.Info("SVM created")
		}
	} else {
		log.Info("Step 7: Create SVM - skipped because already created")
	}

	// STEP 8
	// Check to see if SVM management credentials is available
	if svmCR.Spec.VsadminCredentialSecret.Name != "" {
		// Look up SVM management credentials secret
		vsAdminSecret, err := r.reconcileSecret(ctx,
			svmCR.Spec.VsadminCredentialSecret.Name,
			svmCR.Spec.VsadminCredentialSecret.Namespace, log)
		if err != nil {
			r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_FALSE)
			log.Info("SVM managment credentials NOT available")
			return ctrl.Result{Requeue: false}, nil // not a valid secret - ignore
		} else {
			log.Info("SVM managment credentials available")
			r.setConditionVsadminSecretLookup(ctx, svmCR, CONDITION_STATUS_TRUE)
			// STEP 9
			// Create or update SVM management credentials
			err = r.reconcileSecurityAccount(ctx, svmCR, oc, vsAdminSecret, log)
			if err != nil {
				log.Error(err, "Error while updating SVM management credentials - requeuing")
				return ctrl.Result{Requeue: true}, err
			}
		}
	}

	// Check whether we need to update the SVM
	if !create {
		// STEP 10
		// reconcile SVM update
		err = r.reconcileSvmUpdate(ctx, svmCR, svmRetrieved, oc, log)
		if err != nil {
			log.Error(err, "Error during reconciling SVM update - requeuing")
			return ctrl.Result{Requeue: true}, err
		}

		if svmRetrieved.Uuid != "" {
			// STEP 11
			// Reconcile Management LIF information
			err = r.reconcileManagementLifUpdate(ctx, svmCR, svmRetrieved.Uuid, oc, log)
			if err != nil {
				if strings.Contains(err.Error(), "Duplicate IP") {
					log.Error(err, "Duplicated IP Address - stop reconcile")
					return ctrl.Result{Requeue: false}, nil
				}
				log.Error(err, "Error during reconciling management LIF - requeuing")
				return ctrl.Result{Requeue: true}, err
			}

			// STEP 12
			// Reconcile NFS information
			err = r.reconcileNFSUpdate(ctx, svmCR, svmRetrieved.Uuid, oc, log)
			if err != nil {
				log.Error(err, "Error during reconciling NFS update - requeuing")
				return ctrl.Result{Requeue: true}, err
			}
		}

	}

	log.Info("RECONCILE END")
	return ctrl.Result{Requeue: false}, nil //no error - end reconcile
}

// SetupWithManager sets up the controller with the Manager.
// Adding predicate to prevent hotlooping when the status conditions are updated
// From this: https://github.com/kubernetes-sigs/kubebuilder/issues/618

func (r *StorageVirtualMachineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gatewayv1alpha1.StorageVirtualMachine{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		// Owns(&corev1.Secret{}).
		Complete(r)
}
