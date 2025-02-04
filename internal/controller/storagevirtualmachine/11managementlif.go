// From:  https://github.com/nheidloff/operator-sample-go/blob/main/operator-application/controllers/application/deployment.go

package controller

import (
	"context"
	"encoding/json"
	"fmt"
	defaultLog "log"
	"strconv"

	gateway "gateway/api/v1beta3"
	"gateway/internal/controller/ontap"

	"golang.org/x/exp/slices"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/go-logr/logr"
)

const svmScope = "svm" //magic word

func (r *StorageVirtualMachineReconciler) reconcileManagementLifUpdate(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, uuid string, oc *ontap.Client, log logr.Logger) error {

	log.Info("STEP 11: Update management LIF")

	execute := false
	create := false
	lifUuid := ""

	if svmCR.Spec.ManagementLIF == nil {
		log.Info("No management LIF defined - skipping STEP 11")
		return nil
	}

	// Get current LIFs for SVM provided in UUID
	lifs, err := oc.GetIpInterfacesBySvmUuid(uuid)
	if err != nil {
		log.Error(err, "Error retrieving LIFs for SVM by UUID")
		return err
	}

	if lifs.NumRecords == 0 {
		// no LIFs for the SVM provided in UUID
		// create new LIF
		log.Info("No LIFs defined for SVM: " + uuid + " - creating management LIF")
		create = true
	}

	var upsertManagementLif ontap.IpInterface

	nameIndex := slices.IndexFunc(lifs.Records, func(i ontap.IpInterface) bool { return i.Name == svmCR.Spec.ManagementLIF.Name })

	if oc.Debug {
		defaultLog.Printf("%s", "[DEBUG] nameIndex: "+fmt.Sprintf("%v", nameIndex))
	}

	//IP not returned at least in 9.9.1 vsims - can only check the Name
	if nameIndex != -1 {
		// name found
		// need to update IP of lifs.Records[nameIndex]
		lifUuid = lifs.Records[nameIndex].Uuid

		// Get current LIF details by LIF UUID
		lifRetrieved, err := oc.GetIpInterfaceByLifUuid(lifUuid)
		if err != nil {
			log.Error(err, "Error retreiving LIF details by LIF UUID")
		}

		upsertManagementLif.Name = lifs.Records[nameIndex].Name

		if lifRetrieved.Ip.Address != svmCR.Spec.ManagementLIF.IPAddress {
			// need to update ip address
			execute = true
			upsertManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
			upsertManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
		}

		netmaskAsInt, _ := strconv.Atoi(lifRetrieved.Ip.Netmask)
		netmaskAsIP := NetmaskIntToString(netmaskAsInt)

		if oc.Debug {
			defaultLog.Printf("%s", "[DEBUG] netmaskAsInt: "+fmt.Sprintf("%v", netmaskAsInt))
			defaultLog.Printf("%s", "[DEBUG] netmaskAsIP: "+fmt.Sprintf("%v", netmaskAsIP))
		}

		if netmaskAsIP != svmCR.Spec.ManagementLIF.Netmask {
			// need to update netmask
			execute = true
			upsertManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
			upsertManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
		}
	} else {
		// nothing defined in SVM create new management LIF
		execute = true
		create = true
		upsertManagementLif.Name = svmCR.Spec.ManagementLIF.Name
		upsertManagementLif.Ip.Address = svmCR.Spec.ManagementLIF.IPAddress
		upsertManagementLif.Ip.Netmask = svmCR.Spec.ManagementLIF.Netmask
		upsertManagementLif.Location.BroadcastDomain.Name = svmCR.Spec.ManagementLIF.BroadcastDomain
		upsertManagementLif.Location.HomeNode.Name = svmCR.Spec.ManagementLIF.HomeNode
		upsertManagementLif.ServicePolicy.Name = managementLIFServicePolicy
		upsertManagementLif.Scope = svmScope
		upsertManagementLif.Svm.Uuid = uuid
	}

	if !execute {
		log.Info("No changes detected - skipping STEP 11")
		return nil
	}

	// otherwise changes need to be implemented
	if oc.Debug {
		defaultLog.Printf("%s", "[DEBUG] SVM management LIF update payload: "+fmt.Sprintf("%#v\n", upsertManagementLif))
	}

	jsonPayload, err := json.Marshal(upsertManagementLif)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for SVM managment LIF update")
		_ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
		return err
	}

	if !create {
		// After building update string execute it and check for errors
		log.Info("SVM management LIF update attempt of: " + lifUuid)
		err = oc.PatchIpInterface(lifUuid, jsonPayload)
		if err != nil {
			log.Error(err, "Error occurred when updating SVM management LIF")
			_ = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "SvmUpdateLIFFailed", "Error: "+err.Error())
			return err
		}
		log.Info("SVM management LIF updated successful")
		err = r.setConditionManagementLIFUpdate(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Warning", "SvmUpdateLIFSucceeded", "Updated SVM LIF successfully")
		if err != nil {
			return nil //even though condition not create, don't reconcile again
		}
	} else {
		// Create new management LIF
		log.Info("SVM management LIF creation attempt")
		err = oc.CreateIpInterface(jsonPayload)
		if err != nil {
			log.Error(err, "Error occurred when creating SVM management LIF")
			_ = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_FALSE)
			r.Recorder.Event(svmCR, "Warning", "SvmCreationLIFFailed", "Error: "+err.Error())
			return err
		}
		log.Info("SVM management LIF creation successful")
		err = r.setConditionManagementLIFCreation(ctx, svmCR, CONDITION_STATUS_TRUE)
		r.Recorder.Event(svmCR, "Warning", "SVMCreationLIFSucceeded", "Created SVM LIF successfully")
		if err != nil {
			return nil //even though condition not create, don't reconcile again
		}
	}

	return nil
}

// STEP 11
// Management LIF Upsert
// Note: Status of MANGEMENTLIF_UPSERT can only be true or false
const CONDITION_TYPE_MANGEMENTLIF_UPSERT = "11UpsertdManagementLIF"
const CONDITION_REASON_MANGEMENTLIF_UPDATED = "ManagementLIFUpdate"
const CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_TRUE = "Management LIF update succeeded"
const CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_FALSE = "Management LIF update failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionManagementLIFUpdate(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_MANGEMENTLIF_UPSERT) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPDATED, CONDITION_REASON_MANGEMENTLIF_UPDATED)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_UPDATED, CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_UPDATED, CONDITION_MESSAGE_MANGEMENTLIF_UPDATED_FALSE)
	}
	return nil
}

const CONDITION_REASON_MANGEMENTLIF_CREATION = "ManagementLIFCreation"
const CONDITION_MESSAGE_MANGEMENTLIF_CREATION_TRUE = "Management LIF creation succeeded"
const CONDITION_MESSAGE_MANGEMENTLIF_CREATION_FALSE = "Management LIF creation failed"

func (reconciler *StorageVirtualMachineReconciler) setConditionManagementLIFCreation(ctx context.Context,
	svmCR *gateway.StorageVirtualMachine, status metav1.ConditionStatus) error {

	// I don't want to delete old references to updates to make a history
	// if reconciler.containsCondition(ctx, svmCR, CONDITION_REASON_MANGEMENTLIF_CREATION) {
	// 	reconciler.deleteCondition(ctx, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, CONDITION_REASON_MANGEMENTLIF_CREATION)
	// }

	if status == CONDITION_STATUS_TRUE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_CREATION, CONDITION_MESSAGE_MANGEMENTLIF_CREATION_TRUE)
	}

	if status == CONDITION_STATUS_FALSE {
		return appendCondition(ctx, reconciler.Client, svmCR, CONDITION_TYPE_MANGEMENTLIF_UPSERT, status,
			CONDITION_REASON_MANGEMENTLIF_CREATION, CONDITION_MESSAGE_MANGEMENTLIF_CREATION_FALSE)
	}
	return nil
}
