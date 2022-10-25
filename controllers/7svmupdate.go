// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"golang.org/x/exp/slices"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *StorageVirtualMachineReconciler) reconcileSvmUpdate(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine,
	svmRetrieved ontap.Svm, oc *ontap.Client, log logr.Logger) (ctrl.Result, error) {

	log.Info("Step 7: Update SVM")

	var patchSVM ontap.SvmPatch

	// interate over custom resoource svmCR and look for differences in retrieved SVM
	// if svmCR.Spec.SvmName != svmRetrieved.Name {
	// 	//update name
	// 	patchSVM.Name = svmCR.Spec.SvmName
	// }
	// always set name
	patchSVM.Name = svmCR.Spec.SvmName

	if svmCR.Spec.SvmComment != "" && svmCR.Spec.SvmComment != svmRetrieved.Comment {
		//update comment
		patchSVM.Comment = svmCR.Spec.SvmComment
	}

	if svmCR.Spec.ManagementLIF != nil {

		var patchManagementLif ontap.IpInterface

		ipIndex := slices.IndexFunc(svmRetrieved.IpInterfaces, func(i ontap.IpInterface) bool { return i.Ip.Address == svmCR.Spec.ManagementLIF.IPAddress })
		nameIndex := slices.IndexFunc(svmRetrieved.IpInterfaces, func(i ontap.IpInterface) bool { return i.Name == svmCR.Spec.ManagementLIF.Name })

		log.Info("ipIndex: " + fmt.Sprintf("%v", ipIndex))
		log.Info("nameIndex: " + fmt.Sprintf("%v", nameIndex))

		if ipIndex != -1 {
			if nameIndex != -1 {
				if ipIndex == nameIndex {
					patchManagementLif.Name = svmCR.Spec.ManagementLIF.Name
					patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
					//same object
					if svmRetrieved.IpInterfaces[ipIndex].Ip.Netmask != svmCR.Spec.ManagementLIF.Netmask {
						// need to update netmask
						patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
					}
					if svmRetrieved.IpInterfaces[ipIndex].Location.BroadcastDomain.Name != svmCR.Spec.ManagementLIF.BroacastDomain {
						// need to update broadcast domain
						patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
					}
					if svmRetrieved.IpInterfaces[ipIndex].Location.HomeNode.Name != svmCR.Spec.ManagementLIF.HomeNode {
						// need to update homenode
						patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
					}
				} else {
					// error state - don't know which one to choose the LIf with the correct IP address or the LIF with the correct name
					err := errors.NewBadRequest("Both Managment LIF name and IP address found on different LIFs")
					return ctrl.Result{}, err
				}
			} else {
				// ipIndex correct, no name
				// need to update name of svmRetrieved.IpInterfaces[ipIndex]
				patchManagementLif.Name = svmCR.Spec.ManagementLIF.Name
				patchManagementLif.Ip.Address = svmRetrieved.IpInterfaces[ipIndex].Ip.Address

				if svmRetrieved.IpInterfaces[ipIndex].Ip.Netmask != svmCR.Spec.ManagementLIF.Netmask {
					// need to update netmask
					patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
				}
				if svmRetrieved.IpInterfaces[ipIndex].Location.BroadcastDomain.Name != svmCR.Spec.ManagementLIF.BroacastDomain {
					// need to update broadcast domain
					patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
				}
				if svmRetrieved.IpInterfaces[ipIndex].Location.HomeNode.Name != svmCR.Spec.ManagementLIF.HomeNode {
					// need to update homenode
					patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
				}
			}

		} else {
			//IP not found
			if nameIndex != -1 {
				// name found but not IP
				// need to update IP of svmRetrieved.IpInterfaces[nameIndex]
				patchManagementLif.Name = svmRetrieved.IpInterfaces[nameIndex].Name
				patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress

				if svmRetrieved.IpInterfaces[nameIndex].Ip.Netmask != svmCR.Spec.ManagementLIF.Netmask {
					// need to update netmask
					patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
				}
				if svmRetrieved.IpInterfaces[nameIndex].Location.BroadcastDomain.Name != svmCR.Spec.ManagementLIF.BroacastDomain {
					// need to update broadcast domain
					patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
				}
				if svmRetrieved.IpInterfaces[nameIndex].Location.HomeNode.Name != svmCR.Spec.ManagementLIF.HomeNode {
					// need to update homenode
					patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
				}
			} else {
				// nothing defined in SVM create new
				patchManagementLif.Name = svmCR.Spec.ManagementLIF.Name
				patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
				patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
				patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
				patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
			}

		}

		// add patch management LIF to patchSVM
		patchSVM.IpInterfaces = append(patchSVM.IpInterfaces, patchManagementLif)

	}

	log.Info("SVM update payload: " + fmt.Sprintf("%#v\n", patchSVM))

	jsonPayload, err := json.Marshal(patchSVM)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for SVM update")
		_ = r.setConditionSVMUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, err
	}

	// After building update string execute it and check for errors
	log.Info("SVM update attempt of: " + svmRetrieved.Uuid)
	err = oc.PatchStorageVM(svmRetrieved.Uuid, jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when updating SVM")
		_ = r.setConditionSVMCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
		return ctrl.Result{}, err
	}

	err = r.setConditionSVMUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
	if err != nil {
		return ctrl.Result{}, nil //even though condition not create, don't reconcile again
	}

	return ctrl.Result{}, nil
}
