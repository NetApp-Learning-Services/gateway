// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"golang.org/x/exp/slices"

	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/api/errors"
)

func (r *StorageVirtualMachineReconciler) reconcileManagementLifUpdate(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {

	log.Info("Step 11: Update management LIF")

	create := false
	lifUuid := ""

	if svmCR.Spec.ManagementLIF == nil {
		log.Info("No management LIF defined - skipping step")
		return nil
	}

	// Get current LIFs for SVM provided in UUID
	lifs, err := oc.GetInterfacesForSVMByUUID(uuid)
	if err != nil {
		log.Error(err, "Error retreiving LIFs for SVM by UUID")
	}

	if lifs.NumRecords == 0 {
		// no LIFs for the SVM provided in UUID
		// create new LIF
		log.Info("No LIFs defined for SVM: " + uuid + " - creating management LIF")
		create = true
	} else {
		log.Info(fmt.Sprintf("%#v", lifs))
	}

	var patchManagementLif ontap.IpInterface

	ipIndex := slices.IndexFunc(lifs.Records, func(i ontap.IpInterface) bool { return i.Ip.Address == svmCR.Spec.ManagementLIF.IPAddress })
	nameIndex := slices.IndexFunc(lifs.Records, func(i ontap.IpInterface) bool { return i.Name == svmCR.Spec.ManagementLIF.Name })

	if oc.Debug {
		log.Info("[DEBUG] ipIndex: " + fmt.Sprintf("%v", ipIndex))
		log.Info("[DEBUG] nameIndex: " + fmt.Sprintf("%v", nameIndex))
	}

	if ipIndex != -1 {
		// IP Address in custom resource found in SVM's LIFs
		if nameIndex != -1 {
			// LIF name in custom resource found in SVM's LIFs
			lifUuid = lifs.Records[ipIndex].Uuid
			if ipIndex == nameIndex {
				patchManagementLif.Name = svmCR.Spec.ManagementLIF.Name
				patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
				//same object
				if lifs.Records[ipIndex].Ip.Netmask != svmCR.Spec.ManagementLIF.Netmask {
					// need to update netmask
					patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
				}
				// if lifs.Records[ipIndex].Location.BroadcastDomain.Name != svmCR.Spec.ManagementLIF.BroacastDomain {
				// 	// need to update broadcast domain
				// 	patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
				// }
				// if lifs.Records[ipIndex].Location.HomeNode.Name != svmCR.Spec.ManagementLIF.HomeNode {
				// 	// need to update homenode
				// 	patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
				// }
			} else {
				// error state - don't know which one to choose the LIf with the correct IP address or the LIF with the correct name
				err := errors.NewBadRequest("Both Managment LIF name and IP address found on different LIFs")
				return err
			}
		} else {
			// ipIndex correct, no name
			// need to update name of lifs.Records[ipIndex]
			lifUuid = lifs.Records[ipIndex].Uuid
			patchManagementLif.Name = svmCR.Spec.ManagementLIF.Name
			patchManagementLif.Ip.Address = lifs.Records[ipIndex].Ip.Address

			if lifs.Records[ipIndex].Ip.Netmask != svmCR.Spec.ManagementLIF.Netmask {
				// need to update netmask
				patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
			}
			// if lifs.Records[ipIndex].Location.BroadcastDomain.Name != svmCR.Spec.ManagementLIF.BroacastDomain {
			// 	// need to update broadcast domain
			// 	patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
			// }
			// if lifs.Records[ipIndex].Location.HomeNode.Name != svmCR.Spec.ManagementLIF.HomeNode {
			// 	// need to update homenode
			// 	patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
			// }
		}

	} else {
		//IP not found
		if nameIndex != -1 {
			// name found but not IP
			// need to update IP of lifs.Records[nameIndex]
			lifUuid = lifs.Records[nameIndex].Uuid
			patchManagementLif.Name = lifs.Records[nameIndex].Name
			patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress

			if lifs.Records[nameIndex].Ip.Netmask != svmCR.Spec.ManagementLIF.Netmask {
				// need to update netmask
				patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
			}
			// if lifs.Records[nameIndex].Location.BroadcastDomain.Name != svmCR.Spec.ManagementLIF.BroacastDomain {
			// 	// need to update broadcast domain
			// 	patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
			// }
			// if lifs.Records[nameIndex].Location.HomeNode.Name != svmCR.Spec.ManagementLIF.HomeNode {
			// 	// need to update homenode
			// 	patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
			// }
		} else {
			// nothing defined in SVM create new management LIF
			create = true
			patchManagementLif.Name = svmCR.Spec.ManagementLIF.Name
			patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
			patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
			patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
			patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
		}

	}

	log.Info("SVM management LIF update payload: " + fmt.Sprintf("%#v\n", patchManagementLif))

	jsonPayload, err := json.Marshal(patchManagementLif)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for SVM managment LIF update")
		_ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}

	if !create {
		// After building update string execute it and check for errors
		log.Info("SVM management LIF update attempt of: " + lifUuid)
		err = oc.PatchInterface(lifUuid, jsonPayload)
		if err != nil {
			log.Error(err, "Error occurred when updating SVM management LIF")
			_ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}
		log.Info("SVM management LIF updated successful")
		err = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
		if err != nil {
			return nil //even though condition not create, don't reconcile again
		}
	} else {
		// Create new management LIF
		log.Info("SVM management LIF creation attempt")
		err = oc.CreateInterface(jsonPayload)
		if err != nil {
			log.Error(err, "Error occurred when creating SVM management LIF")
			_ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
			return err
		}
		log.Info("SVM management LIF creation successful")
		err = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_TRUE)
		if err != nil {
			return nil //even though condition not create, don't reconcile again
		}
	}

	return nil
}
