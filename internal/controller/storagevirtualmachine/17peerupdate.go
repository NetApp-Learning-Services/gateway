package controller

import (
	"context"
	"encoding/json"
	"fmt"
	gateway "gateway/api/v1beta2"
	"gateway/internal/controller/ontap"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const InterclusterLifServicePolicy = "default-intercluster" //magic word
const InterclusterLifServicePolicyScope = "cluster"         //magic word
const ClusterPeerAvailable = "available"                    //magic word

func (r *StorageVirtualMachineReconciler) reconcilePeerUpdate(ctx context.Context, svmCR *gateway.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {

	log.Info("STEP 17: Update Peering service")

	// Check to see if Peer configuration is provided in custom resource
	if svmCR.Spec.PeerConfig == nil {
		// If not, exit with no error
		log.Info("No Peering service defined - skipping STEP 17")
		return nil
	}

	// Intercluster LIFS

	// Check to see if Intercluster interfaces are defined in custom resource
	if svmCR.Spec.PeerConfig.Lifs == nil {
		// If not, exit with no error
		log.Info("No Intercluster LIFs defined - skipping updates")
	} else {
		createInterclusterLifs := false

		// Check to see if Intercluster interfaces defined and compare to custom resource's definitions
		lifs, err := oc.GetIpInterfacesByServicePolicy(InterclusterLifServicePolicy)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error getting Intercluster LIFs for cluster: "+svmCR.Spec.ClusterManagementHost)
			_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if lifs.NumRecords == 0 {
			// no Intercluster LIFs i cluster
			// create new LIF(s)
			log.Info("No Intercluster LIFs defined for cluster: " + svmCR.Spec.ClusterManagementHost + " - creating Intercluster Lif(s)")
			createInterclusterLifs = true
		}

		if createInterclusterLifs {
			//creating lifs
			for _, val := range svmCR.Spec.PeerConfig.Lifs {
				err = CreateLif(val, InterclusterLifServicePolicy, InterclusterLifServicePolicyScope, uuid, oc, log)
				if err != nil {
					_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				}
			}

		} else {
			// update LIFs
			for index, val := range svmCR.Spec.PeerConfig.Lifs {

				// Check to see if lifs.Records[index] is out of index - if so, need to create LIF
				if index > lifs.NumRecords-1 {
					// Need to create LIF for val
					err = CreateLif(val, InterclusterLifServicePolicy, InterclusterLifServicePolicyScope, uuid, oc, log)
					if err != nil {
						_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "PeerCreationLifFailed", "Error: "+err.Error())
						return err
					}

				} else {

					if reflect.ValueOf(lifs.Records[index]).IsZero() {
						break
					}

					err = UpdateLif(val, lifs.Records[index], InterclusterLifServicePolicy, oc, log)
					if err != nil {
						_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "PeerUpdateLifFailed", "Error: "+err.Error())
						return err
					}
				}

			} // Need looping through Intercluster LIF definitions in custom resource

			// // Delete all SVM Intercluster LIFs that are not defined in the custom resource
			// for i := len(svmCR.Spec.PeerConfig.Lifs); i < lifs.NumRecords; i++ {
			// 	log.Info("Intercluster LIF delete attempt: " + lifs.Records[i].Name)
			// 	oc.DeleteIpInterface(lifs.Records[i].Uuid)
			// 	if err != nil {
			// 		log.Error(err, "Error occurred when deleting Intercluster LIF: "+lifs.Records[i].Name)
			// 		// don't requeue on failed delete request
			// 		// no condition error
			// 		// return err
			// 	} else {
			// 		log.Info("Intercluster LIF delete successful: " + lifs.Records[i].Name)
			// 	}
			// }

		} // Checking for NFS LIFs updates
		_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "PeerUpsertLifSucceeded", "Upserted Intercluster LIF(s) successfully")
	} // LIFs defined in custom resource

	// END NFS LIFS

	// CLUSTER PEERING SERVICE

	createPeeringService := false
	establishedPeeringService := false
	clusterPeerName := ""

	clusterPeerServices, err := oc.GetClusterPeerServicesForCluster(svmCR.Spec.PeerConfig.Remote.Clustername)
	if err != nil && errors.IsNotFound(err) {
		createPeeringService = true
	} else if err != nil {
		//some other error
		log.Error(err, "Error retrieving Cluster peering service - requeuing")
		return err
	}

	var upsertPeerService ontap.ClusterPeerService

	if createPeeringService {
		log.Info("No Cluster peering service defined for cluster: " + svmCR.Spec.PeerConfig.Remote.Clustername + " - creating Peer service")
		upsertPeerService.Name = svmCR.Spec.PeerConfig.Name
		upsertPeerService.Authentication.Passphrase = svmCR.Spec.PeerConfig.Passphrase
		upsertPeerService.Encryption.Proposed = svmCR.Spec.PeerConfig.Encryption
		for _, val := range svmCR.Spec.PeerConfig.Applications {
			upsertPeerService.Applications = append(upsertPeerService.Applications, val.App)
		}
		upsertPeerService.Remote.Name = svmCR.Spec.PeerConfig.Remote.Clustername
		for _, val := range svmCR.Spec.PeerConfig.Remote.Ipaddresses {
			upsertPeerService.Remote.Addresses = append(upsertPeerService.Remote.Addresses, val.IPAddress)
		}
		var localSVM ontap.SvmRef
		localSVM.Name = svmCR.Spec.SvmName
		if svmCR.Spec.SvmUuid != "" {
			localSVM.Uuid = svmCR.Spec.SvmUuid
		}
		upsertPeerService.InitialAllowedSVMs = append(upsertPeerService.InitialAllowedSVMs, localSVM)

		jsonPayload, err := json.Marshal(upsertPeerService)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for cluster peer service creation - requeuing")
			_ = r.setConditionPeerClusterService(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if oc.Debug {
			log.Info("[DEBUG] Cluster Peer service creation payload: " + fmt.Sprintf("%#v\n", upsertPeerService))
		}

		err = oc.CreateClusterPeerService(jsonPayload)
		if err != nil {
			log.Error(err, "Error creating the cluster peer service - requeuing")
			_ = r.setConditionPeerClusterService(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "PeerCreationFailed", "Error: "+err.Error())
			return err
		}
		log.Info("Cluster peer request created successful")
	}

	//Cluster Peering service already created

	if clusterPeerServices.NumRecords != 0 && svmCR.Spec.PeerConfig.Remote.ClusterUuid == "" {
		//Check to see if peer state is available
		for _, val := range clusterPeerServices.Records {
			if val.Status.State == ClusterPeerAvailable {
				log.Info("Remote cluster " + val.Remote.Name + " accepted")
				clusterPeerName = val.Remote.Name
				establishedPeeringService = true

				//Aadd the remote cluster uuid to CR
				patch := client.MergeFrom(svmCR.DeepCopy())
				svmCR.Spec.PeerConfig.Remote.ClusterUuid = val.Uuid
				err = r.Patch(ctx, svmCR, patch)
				if err != nil {
					log.Error(err, "Error patching the new cluster peer uuid in the custom resource - requeuing")
					r.Recorder.Event(svmCR, "Warning", "PeerCreationFailed", "Error: "+err.Error())
					_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
					return err
				}

				_ = r.setConditionPeerClusterService(ctx, svmCR, CONDITION_STATUS_TRUE)
				r.Recorder.Event(svmCR, "Normal", "ClusterPeerCreationSucceeded", "Created cluster peer service successfully")
				log.Info("Cluster peer service created successful with remote cluster " + clusterPeerName)
			}
		}
	}

	if !establishedPeeringService {
		//need to requeue or recheck
		err = errors.NewInvalid(svmCR.GroupVersionKind().GroupKind(), "Peer requested - waiting for respond", nil)
		log.Error(err, "Error - peer has NOT responded - requeuing")
		return err
	}

	// Should have an available cluster peer relationship
	log.Info("Proceeding with SVM peering")

	// END CLUSTER PEERING SERVICE

	return nil
}

// STEP 17
// Peer update
// Note: Status of PEER_SERVICE can only be true or false
const CONDITION_TYPE_PEER_SERVICE = "17Peerservice"
const CONDITION_REASON_PEER_SERVICE = "PeerClusterservice"
const CONDITION_MESSAGE_PEER_SERVICE_TRUE = "Cluster peer service configuration succeeded"
const CONDITION_MESSAGE_PEER_SERVICE_FALSE = "Cluster peer service configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionPeerClusterService(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_PEER_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_PEER_SERVICE, CONDITION_REASON_PEER_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEER_SERVICE, status,
			CONDITION_REASON_PEER_SERVICE, CONDITION_MESSAGE_PEER_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEER_SERVICE, status,
			CONDITION_REASON_PEER_SERVICE, CONDITION_MESSAGE_PEER_SERVICE_FALSE)
	}
	return nil
}

const CONDITION_REASON_PEER_LIF = "Peerlif"
const CONDITION_MESSAGE_PEER_LIF_TRUE = "Peer LIF configuration succeeded"
const CONDITION_MESSAGE_PEER_LIF_FALSE = "Peer LIF configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionPeerLif(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_PEER_LIF) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_PEER_SERVICE, CONDITION_REASON_PEER_LIF)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEER_SERVICE, status,
			CONDITION_REASON_PEER_LIF, CONDITION_MESSAGE_PEER_LIF_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEER_SERVICE, status,
			CONDITION_REASON_PEER_LIF, CONDITION_MESSAGE_PEER_LIF_FALSE)
	}
	return nil
}
