package controller

import (
	"context"
	"encoding/json"
	"fmt"
	gateway "gateway/api/v1beta2"
	"gateway/internal/controller/ontap"
	"strings"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const InterclusterLifServicePolicy = "default-intercluster" //magic word
const InterclusterLifServicePolicyScope = "cluster"         //magic word
const ClusterPeerAvailable = "available"                    //magic word
const SvmPeerPending = "pending"                            //magic word
const SvmPeerPeered = "peered"                              //magic word

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
			createLif := true
			currentLifIndex := -1
			for _, val := range svmCR.Spec.PeerConfig.Lifs {

				for i, lif := range lifs.Records {
					if val.Name == lif.Name {
						//skip this one
						createLif = false
						currentLifIndex = i
						log.Info("Intercluster lif " + val.Name + " with IP address " + val.IPAddress + " exists. Updating.")
					}
				}

				if createLif {
					// Need to create LIF for val
					err = CreateLif(val, InterclusterLifServicePolicy, InterclusterLifServicePolicyScope, uuid, oc, log)
					if err != nil {
						_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "PeerCreationLifFailed", "Error: "+err.Error())
						return err
					}

				} else {
					err = UpdateLif(val, lifs.Records[currentLifIndex], InterclusterLifServicePolicy, oc, log)
					if err != nil {
						_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "PeerUpdateLifFailed", "Error: "+err.Error())
						return err
					}
					createLif = true
					currentLifIndex = -1
				}

			}

		} // Checking for NFS LIFs updates
		_ = r.setConditionPeerLif(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Normal", "PeerUpsertLifSucceeded", "Upserted Intercluster LIF(s) successfully")
	} // LIFs defined in custom resource

	// END NFS LIFS

	// CLUSTER PEERING

	log.Info("Check Cluster peer relationship")
	createClusterPeer := true //default true

	clusterPeers, err := oc.GetClusterPeers()
	if err != nil && errors.IsNotFound(err) {
		createClusterPeer = true
	} else if err != nil {
		//some other error
		log.Error(err, "Error retrieving Cluster peer - requeuing")
		return err
	}

	if clusterPeers.NumRecords != 0 {
		for _, val := range clusterPeers.Records {
			if val.Name == svmCR.Spec.PeerConfig.Name {
				for _, ip := range val.Remote.Addresses {
					if ip == svmCR.Spec.PeerConfig.Remote.Ipaddress {
						createClusterPeer = false
					}
				}
			}
		}
	}

	var upsertClusterPeer ontap.ClusterPeer

	if createClusterPeer {

		log.Info("No Cluster peer defined for cluster: " + svmCR.Spec.PeerConfig.Remote.Ipaddress + " - creating Cluster Peer")
		upsertClusterPeer.Name = svmCR.Spec.PeerConfig.Name
		upsertClusterPeer.Authentication.Passphrase = svmCR.Spec.PeerConfig.Passphrase
		upsertClusterPeer.Encryption.Proposed = svmCR.Spec.PeerConfig.Encryption
		for _, val := range svmCR.Spec.PeerConfig.Applications {
			upsertClusterPeer.Applications = append(upsertClusterPeer.Applications, val.App)
		}

		upsertClusterPeer.Remote.Addresses = append(upsertClusterPeer.Remote.Addresses, svmCR.Spec.PeerConfig.Remote.Ipaddress)

		var localSVM ontap.SvmRef
		localSVM.Name = svmCR.Spec.SvmName
		upsertClusterPeer.InitialAllowedSVMs = append(upsertClusterPeer.InitialAllowedSVMs, localSVM)

		jsonPayload, err := json.Marshal(upsertClusterPeer)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for cluster peer creation - requeuing")
			_ = r.setConditionPeerClusterService(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		if oc.Debug {
			log.Info("[DEBUG] Cluster peer creation payload: " + fmt.Sprintf("%#v\n", upsertClusterPeer))
		}

		err = oc.CreateClusterPeer(jsonPayload)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "An introductory RPC to the peer address") {
				log.Info("Waiting for cluster peer to respond")
				return err
			} else {
				log.Error(err, "Error creating the cluster peer - requeuing")
				_ = r.setConditionPeerClusterService(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "ClusterPeerCreationFailed", "Error: "+err.Error())
				return err
			}

		}
		log.Info("Cluster peer request created successful - requeuing to wait for respond")
		return errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "waiting for cluster peer")
	} else {
		if clusterPeers.NumRecords != 0 {
			requeue := true
			for _, val := range clusterPeers.Records {
				if val.Status.State == ClusterPeerAvailable {
					requeue = false

					if svmCR.Spec.PeerConfig.Remote.Clustername == "" {
						log.Info("Remote cluster " + val.Remote.Name + " accepted")
						//Add the remote cluster name to CR
						patch := client.MergeFrom(svmCR.DeepCopy())
						svmCR.Spec.PeerConfig.Remote.Clustername = val.Remote.Name
						err = r.Patch(ctx, svmCR, patch)
						if err != nil {
							log.Error(err, "Error patching the new cluster peer uuid in the custom resource - requeuing")
							r.Recorder.Event(svmCR, "Warning", "ClusterPeerCreationFailed", "Error: "+err.Error())
							_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
							return err
						}
						_ = r.setConditionPeerClusterService(ctx, svmCR, CONDITION_STATUS_TRUE)
						r.Recorder.Event(svmCR, "Normal", "ClusterPeerCreationSucceeded", "Created cluster peer successfully")
					}

					log.Info("Cluster peer created successful with remote cluster " + val.Remote.Name)
				}
			}
			if requeue {
				log.Info("Waiting for cluster peer to be available - requeuing")
				return errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "waiting for cluster peer")
			}
		} else {
			log.Info("Cluster peer not created - requeuing")
			return errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "waiting for cluster peer")
		}
	}

	// END CLUSTER PEERING

	// SVM PEERING

	// Should have an available cluster peer relationship
	log.Info("Checking SVM peer relationship")
	createSvmPeer := true //default true

	svmPeers, err := oc.GetSvmPeers(svmCR.Spec.SvmName)
	if err != nil && errors.IsNotFound(err) {
		createSvmPeer = true
	} else if err != nil {
		//some other error
		log.Error(err, "Error retrieving SVM peer - requeuing")
		return err
	}

	if svmPeers.NumRecords != 0 {
		for _, val := range svmPeers.Records {
			if val.Peer.Cluster.Name == svmCR.Spec.PeerConfig.Remote.Clustername {
				createSvmPeer = false
			}
		}
	}

	var upsertSvmPeer ontap.SvmPeer

	if createSvmPeer {

		log.Info("No SVM peer for remote cluster " + svmCR.Spec.PeerConfig.Remote.Clustername + " and local SVM " + svmCR.Spec.SvmName + " - creating SVM peer")
		for _, val := range svmCR.Spec.PeerConfig.Applications {
			upsertSvmPeer.Applications = append(upsertSvmPeer.Applications, val.App)
		}
		upsertSvmPeer.LocalSvm.Name = svmCR.Spec.SvmName
		upsertSvmPeer.Peer.Cluster.Name = svmCR.Spec.PeerConfig.Name
		upsertSvmPeer.Peer.Svm.Name = svmCR.Spec.PeerConfig.Remote.Svmname

		jsonPayload, err := json.Marshal(upsertSvmPeer)
		if err != nil {
			//error creating the json body
			log.Error(err, "Error creating the json payload for SVM peer creation - requeuing")
			_ = r.setConditionPeerSvmService(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}

		//if oc.Debug {
		log.Info("[DEBUG] SVM Peer creation payload: " + fmt.Sprintf("%#v\n", upsertSvmPeer))
		//}

		err = oc.CreateSvmPeer(jsonPayload)
		if err != nil {
			if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "An introductory RPC to the peer address") {
				log.Info("Waiting for SVM peer to respond")
				return err
			} else {
				log.Error(err, "Error creating the SVM peer - requeuing")
				_ = r.setConditionPeerSvmService(ctx, svmCR, CONDITION_STATUS_FALSE)
				r.Recorder.Event(svmCR, "Warning", "SvmPeerCreationFailed", "Error: "+err.Error())
				return err
			}

		}
		log.Info("SVM peer request created successful - requeuing to wait for respond")
		return errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "waiting for SVM peer")
	} else {
		if svmPeers.NumRecords != 0 {
			requeue := true
			for _, val := range svmPeers.Records {
				if val.State == SvmPeerPending && val.Peer.Cluster.Name == svmCR.Spec.PeerConfig.Remote.Clustername {
					log.Info("Remote SVM " + val.Peer.Svm.Name + "peer request pending - patching")

					//PATCH
					var patchSvmPeer ontap.SvmPeerPatch
					patchSvmPeer.State = SvmPeerPeered
					jsonPayload, err := json.Marshal(patchSvmPeer)
					if err != nil {
						//error creating the json body
						log.Error(err, "Error creating the json payload for SVM peer patch - requeuing")
						_ = r.setConditionPeerSvmService(ctx, svmCR, CONDITION_STATUS_FALSE)
						return err
					}

					if oc.Debug {
						log.Info("[DEBUG] SVM Peer patch payload: " + fmt.Sprintf("%#v\n", patchSvmPeer))
					}

					err = oc.PatchSvmPeer(jsonPayload, val.Uuid)
					if err != nil {
						log.Error(err, "Error patching the SVM peer - requeuing")
						_ = r.setConditionPeerSvmService(ctx, svmCR, CONDITION_STATUS_FALSE)
						r.Recorder.Event(svmCR, "Warning", "SvmPeerPatchFailed", "Error: "+err.Error())
						return err
					}
					log.Info("SVM peer patch created successful - requeuing to verify SVM peer")
					return errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "waiting for SVM peer")
				} else if val.State == SvmPeerPeered {
					requeue = false
					_ = r.setConditionPeerSvmService(ctx, svmCR, CONDITION_STATUS_TRUE)
					r.Recorder.Event(svmCR, "Normal", "SvmPeerCreationSucceeded", "Created SVM peer successfully")
					log.Info("SVM peer created successful with remote SVM: " + svmCR.Spec.PeerConfig.Remote.Svmname)
				}

			}
			if requeue {
				log.Info("Waiting for SVM peer to be peered - requeuing")
				return errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "waiting for SVM peer")
			}
		} else {
			log.Info("SVM peer not created - requeuing")
			return errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "waiting for SVM peer")
		}
	}

	// END SVM PEERING SERVICE

	return nil
}

// STEP 17
// Peer update
// Note: Status of PEER_SERVICE can only be true or false
const CONDITION_TYPE_PEERCLUSTER_SERVICE = "17PeerCluster"
const CONDITION_REASON_PEERCLUSTER_SERVICE = "PeerCluster"
const CONDITION_MESSAGE_PEERCLUSTER_SERVICE_TRUE = "Cluster peer configuration succeeded"
const CONDITION_MESSAGE_PEERCLUSTER_SERVICE_FALSE = "Cluster peer configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionPeerClusterService(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_PEERCLUSTER_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_PEERCLUSTER_SERVICE, CONDITION_REASON_PEERCLUSTER_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEERCLUSTER_SERVICE, status,
			CONDITION_REASON_PEERCLUSTER_SERVICE, CONDITION_MESSAGE_PEERCLUSTER_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEERCLUSTER_SERVICE, status,
			CONDITION_REASON_PEERCLUSTER_SERVICE, CONDITION_MESSAGE_PEERCLUSTER_SERVICE_FALSE)
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
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEERCLUSTER_SERVICE, status,
			CONDITION_REASON_PEER_LIF, CONDITION_MESSAGE_PEER_LIF_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEERCLUSTER_SERVICE, status,
			CONDITION_REASON_PEER_LIF, CONDITION_MESSAGE_PEER_LIF_FALSE)
	}
	return nil
}

const CONDITION_TYPE_PEERSVM_SERVICE = "17PeerSvm"
const CONDITION_REASON_PEERSVM_SERVICE = "PeerSvm"
const CONDITION_MESSAGE_PEERSVM_SERVICE_TRUE = "SVM peer configuration succeeded"
const CONDITION_MESSAGE_PEERSVM_SERVICE_FALSE = "Cluster peer configuration failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionPeerSvmService(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_PEERSVM_SERVICE) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_PEERSVM_SERVICE, CONDITION_REASON_PEERSVM_SERVICE)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEERSVM_SERVICE, status,
			CONDITION_REASON_PEERSVM_SERVICE, CONDITION_MESSAGE_PEERSVM_SERVICE_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_PEERSVM_SERVICE, status,
			CONDITION_REASON_PEERSVM_SERVICE, CONDITION_MESSAGE_PEERSVM_SERVICE_FALSE)
	}
	return nil
}
