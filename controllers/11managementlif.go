// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	gatewayv1alpha1 "gateway/api/v1alpha1"
	"gateway/ontap"

	"golang.org/x/exp/slices"

	"github.com/go-logr/logr"
)

func (r *StorageVirtualMachineReconciler) reconcileManagementLifUpdate(ctx context.Context, svmCR *gatewayv1alpha1.StorageVirtualMachine,
	uuid string, oc *ontap.Client, log logr.Logger) error {

	log.Info("Step 11: Update management LIF")

	execute := false
	create := false
	lifUuid := ""

	if svmCR.Spec.ManagementLIF == nil {
		log.Info("No management LIF defined - skipping Step 11")
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
	}

	var patchManagementLif ontap.IpInterface

	nameIndex := slices.IndexFunc(lifs.Records, func(i ontap.IpInterface) bool { return i.Name == svmCR.Spec.ManagementLIF.Name })

	if oc.Debug {
		log.Info("[DEBUG] nameIndex: " + fmt.Sprintf("%v", nameIndex))
	}

	//IP not returned at least in 9.9.1 vsims - can only check the Name
	if nameIndex != -1 {
		// name found
		// need to update IP of lifs.Records[nameIndex]
		lifUuid = lifs.Records[nameIndex].Uuid

		// Get current LIF details by LIF UUID
		lifRetrieved, err := oc.GetInterfaceByUUID(lifUuid)
		if err != nil {
			log.Error(err, "Error retreiving LIF details by LIF UUID")
		}

		patchManagementLif.Name = lifs.Records[nameIndex].Name

		if lifRetrieved.Ip.Address != svmCR.Spec.ManagementLIF.IPAddress {
			// need to update ip address
			execute = true
			patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
			patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
		}

		netmaskAsInt, _ := strconv.Atoi(lifRetrieved.Ip.Netmask)
		netmaskAsIP := NetMaskToString(netmaskAsInt)
		if oc.Debug {
			log.Info("[DEBUG] netmaskAsInt: " + fmt.Sprintf("%v", netmaskAsInt))
			log.Info("[DEBUG] netmaskAsIP: " + fmt.Sprintf("%v", netmaskAsIP))
		}

		if netmaskAsIP != svmCR.Spec.ManagementLIF.Netmask {
			// need to update netmask
			execute = true
			patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
			patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
		}
	} else {
		// nothing defined in SVM create new management LIF
		execute = true
		create = true
		patchManagementLif.Name = svmCR.Spec.ManagementLIF.Name
		patchManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
		patchManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
		patchManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroacastDomain
		patchManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
		patchManagementLif.Scope = "svm" //special word
	}

	if !execute {
		log.Info("No changes detected - skipping Step 11")
		return nil
	}

	// otherwise changes need to be implemented
	if oc.Debug {
		log.Info("[DEBUG] SVM management LIF update payload: " + fmt.Sprintf("%#v\n", patchManagementLif))
	}

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

func NetMaskToString(mask int) (netmaskstring string) {
	var binarystring string

	for ii := 1; ii <= mask; ii++ {
		binarystring = binarystring + "1"
	}
	for ii := 1; ii <= (32 - mask); ii++ {
		binarystring = binarystring + "0"
	}
	oct1 := binarystring[0:8]
	oct2 := binarystring[8:16]
	oct3 := binarystring[16:24]
	oct4 := binarystring[24:]

	ii1, _ := strconv.ParseInt(oct1, 2, 64)
	ii2, _ := strconv.ParseInt(oct2, 2, 64)
	ii3, _ := strconv.ParseInt(oct3, 2, 64)
	ii4, _ := strconv.ParseInt(oct4, 2, 64)
	netmaskstring = strconv.Itoa(int(ii1)) + "." + strconv.Itoa(int(ii2)) + "." + strconv.Itoa(int(ii3)) + "." + strconv.Itoa(int(ii4))
	return
}
