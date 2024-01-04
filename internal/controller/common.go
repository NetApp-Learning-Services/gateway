package controller

import (
	"encoding/json"
	"fmt"
	gateway "gateway/api/v1beta1"
	"gateway/internal/controller/ontap"
	"strconv"

	"github.com/go-logr/logr"
)

type apiError struct {
	errorCode int64
	err       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%d - API Error - %s", e.errorCode, e.err)
}
func (e *apiError) ErrorCode() int64 {
	return e.errorCode
}

func NetmaskIntToString(mask int) (netmaskstring string) {
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

func CreateLif(lifToCreate gateway.LIF, lifType string, uuid string, oc *ontap.Client, log logr.Logger) (err error) {
	var newLif ontap.IpInterface
	newLif.Name = lifToCreate.Name
	newLif.Ip.Address = lifToCreate.IPAddress
	newLif.Ip.Netmask = lifToCreate.Netmask
	newLif.Location.BroadcastDomain.Name = lifToCreate.BroacastDomain
	newLif.Location.HomeNode.Name = lifToCreate.HomeNode
	newLif.ServicePolicy.Name = lifType
	newLif.Scope = NfsLifScope
	newLif.Svm.Uuid = uuid

	jsonPayload, err := json.Marshal(newLif)
	if err != nil {
		//error creating the json body
		log.Error(err, fmt.Sprintf("Error creating the json payload for LIF creation: %v of type %v", lifToCreate.Name, lifType))
		return err
	}
	log.Info("LIF creation attempt: " + lifToCreate.Name)
	err = oc.CreateIpInterface(jsonPayload)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred when creating LIF: %v of type %v", lifToCreate.Name, lifType))
		return err
	}
	log.Info(fmt.Sprintf("LIF creation successful: %v of type %v", lifToCreate.Name, lifType))

	return nil
}

func UpdateLif(lifDefinition gateway.LIF, lifToUpdate ontap.IpInterface, lifType string, oc *ontap.Client, log logr.Logger) (err error) {

	netmaskAsInt, _ := strconv.Atoi(lifToUpdate.Ip.Netmask)
	netmaskAsIP := NetmaskIntToString(netmaskAsInt)

	if lifToUpdate.Ip.Address != lifDefinition.IPAddress ||
		lifToUpdate.Name != lifDefinition.Name ||
		netmaskAsIP != lifDefinition.Netmask ||
		lifToUpdate.ServicePolicy.Name != lifType ||
		!lifToUpdate.Enabled {
		//reset value
		var updateLif ontap.IpInterface
		updateLif.Name = lifDefinition.Name
		updateLif.Ip.Address = lifDefinition.IPAddress
		updateLif.Ip.Netmask = lifDefinition.Netmask
		//updateLif.Location.BroadcastDomain.Name = lifDefinition.BroacastDomain
		//updateLif.Location.HomeNode.Name = lifDefinition.HomeNode
		updateLif.ServicePolicy.Name = lifType
		updateLif.Enabled = true

		jsonPayload, err := json.Marshal(updateLif)
		if err != nil {
			//error creating the json body
			log.Error(err, fmt.Sprintf("Error creating json payload occurred when updating LIF: %v of type %v", lifToUpdate.Name, lifType))
			return &apiError{1, err.Error()}
		}
		log.Info(fmt.Sprintf("LIF update attempt:  %v of type %v", lifToUpdate.Name, lifType))
		err = oc.PatchIpInterface(lifToUpdate.Uuid, jsonPayload)
		if err != nil {
			log.Error(err, fmt.Sprintf("Error occurred when updating LIF: %v of type %v", lifToUpdate.Name, lifType))
			return &apiError{2, err.Error()}
		}

		log.Info(fmt.Sprintf("LIF update successful: %v of type %v", lifToUpdate.Name, lifType))

	} else {
		log.Info(fmt.Sprintf("No changes detected for LIf: %v of type %v", lifToUpdate.Name, lifType))
	}
	return nil
}
