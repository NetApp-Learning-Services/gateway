package controller

import (
	"encoding/json"
	"fmt"
	gateway "gateway/api/v1beta2"
	"gateway/internal/controller/ontap"
	"strconv"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
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

func CreateLif(lifToCreate gateway.LIF, lifServicePolicy string, lifServicePolicyScope string, uuid string, oc *ontap.Client, log logr.Logger) (err error) {
	var newLif ontap.IpInterface
	newLif.Name = lifToCreate.Name
	newLif.Ip.Address = lifToCreate.IPAddress
	newLif.Ip.Netmask = lifToCreate.Netmask
	newLif.Location.BroadcastDomain.Name = lifToCreate.BroacastDomain
	newLif.Location.HomeNode.Name = lifToCreate.HomeNode
	newLif.ServicePolicy.Name = lifServicePolicy
	newLif.Scope = lifServicePolicyScope
	if lifServicePolicyScope == "cluster" { //magic word
		if lifToCreate.Ipspace != "" {
			newLif.Ipspace.Name = lifToCreate.Ipspace
		} else {
			//required for cluster-scoped lifs
			newLif.Ipspace.Name = "Default" //magic word
		}
	} else {
		//scope is svm
		newLif.Svm.Uuid = uuid
	}

	jsonPayload, err := json.Marshal(newLif)
	if err != nil {
		//error creating the json body
		log.Error(err, fmt.Sprintf("Error creating the json payload for LIF creation: %v of type %v", lifToCreate.Name, lifServicePolicy))
		return err
	}
	log.Info("LIF creation attempt: " + lifToCreate.Name)
	err = oc.CreateIpInterface(jsonPayload)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred when creating LIF: %v of type %v", lifToCreate.Name, lifServicePolicy))
		return err
	}
	log.Info(fmt.Sprintf("LIF creation successful: %v of type %v", lifToCreate.Name, lifServicePolicy))

	return nil
}

func UpdateLif(lifDefinition gateway.LIF, lifToUpdate ontap.IpInterface, lifServicePolicy string, oc *ontap.Client, log logr.Logger) (err error) {

	netmaskAsInt, _ := strconv.Atoi(lifToUpdate.Ip.Netmask)
	netmaskAsIP := NetmaskIntToString(netmaskAsInt)

	if lifToUpdate.Ip.Address != lifDefinition.IPAddress ||
		lifToUpdate.Name != lifDefinition.Name ||
		netmaskAsIP != lifDefinition.Netmask ||
		lifToUpdate.ServicePolicy.Name != lifServicePolicy ||
		!lifToUpdate.Enabled {
		//reset value
		var updateLif ontap.IpInterface
		updateLif.Name = lifDefinition.Name
		updateLif.Ip.Address = lifDefinition.IPAddress
		updateLif.Ip.Netmask = lifDefinition.Netmask
		//updateLif.Location.BroadcastDomain.Name = lifDefinition.BroacastDomain
		//updateLif.Location.HomeNode.Name = lifDefinition.HomeNode
		updateLif.ServicePolicy.Name = lifServicePolicy
		updateLif.Enabled = true

		jsonPayload, err := json.Marshal(updateLif)
		if err != nil {
			//error creating the json body
			log.Error(err, fmt.Sprintf("Error creating json payload occurred when updating LIF: %v of type %v", lifToUpdate.Name, lifServicePolicy))
			return &apiError{1, err.Error()}
		}
		log.Info(fmt.Sprintf("LIF update attempt:  %v of type %v", lifToUpdate.Name, lifServicePolicy))
		err = oc.PatchIpInterface(lifToUpdate.Uuid, jsonPayload)
		if err != nil {
			log.Error(err, fmt.Sprintf("Error occurred when updating LIF: %v of type %v", lifToUpdate.Name, lifServicePolicy))
			return &apiError{2, err.Error()}
		}

		log.Info(fmt.Sprintf("LIF update successful: %v of type %v", lifToUpdate.Name, lifServicePolicy))

	} else {
		log.Info(fmt.Sprintf("No changes detected for LIf: %v of type %v", lifToUpdate.Name, lifServicePolicy))
	}
	return nil
}

func CreateUser(userToCreate gateway.S3User, uuid string, oc *ontap.Client, log logr.Logger) (user ontap.S3UsersResponse, err error) {
	var newUser ontap.S3User
	newUser.Name = userToCreate.Name

	jsonPayload, err := json.Marshal(newUser)
	if err != nil {
		//error creating the json body
		log.Error(err, fmt.Sprintf("Error creating the json payload for S3 User creation: %v", userToCreate.Name))
		return user, err
	}
	log.Info("S3 User creation attempt: " + userToCreate.Name)
	user, err = oc.CreateS3User(uuid, jsonPayload)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred when creating S3 User: %v", userToCreate.Name))
		return user, err
	}
	log.Info(fmt.Sprintf("S3 User creation successful: %v", userToCreate.Name))

	return user, nil
}

func CreateLifServicePolicy(servicePolicyName string, servicePolicyScope string, uuid string, oc *ontap.Client, log logr.Logger) (err error) {
	var newServicePolicy ontap.IpServicePolicy
	newServicePolicy.Name = servicePolicyName
	newServicePolicy.Scope = servicePolicyScope
	newServicePolicy.Svm.Uuid = uuid
	servicesToAdd := [3]string{"data-core", "data-s3-server", "data-dns-server"}
	newServicePolicy.Services = servicesToAdd[:]

	jsonPayload, err := json.Marshal(newServicePolicy)
	if err != nil {
		//error creating the json body
		log.Error(err, fmt.Sprintf("Error creating the json payload for LIF S3 Service Policy %v", newServicePolicy.Name))
		return err
	}
	log.Info("LIF service policy creation attempt: " + newServicePolicy.Name)
	err = oc.CreateInterfaceServicePolicy(jsonPayload)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred when creating LIF S3 Service Policy: %v", newServicePolicy.Name))
		return err
	}
	log.Info(fmt.Sprintf("LIF S3 Service Policy creation successful: %v", newServicePolicy.Name))

	return nil
}

func CreateServerCertificate(commonName string, catype string, expiryTime string, uuid string, svmName string, oc *ontap.Client, log logr.Logger) (returnCert ontap.Certificate, err error) {

	createNewCACertificate := false
	var cert ontap.Certificate
	//check if exists

	log.Info("Checking for a CA certificate " + commonName)

	resp, err := oc.GetCertificatesBySvmUuid(uuid, commonName, catype)

	if err != nil {
		if errors.IsNotFound((err)) {
			createNewCACertificate = true
		} else {
			//unknown error
			log.Error(err, fmt.Sprintf("Error while checking for the S3 certificate required: %v", err.Error()))
			return returnCert, err
		}
	}

	if resp.NumRecords != 0 {
		cert = resp.Records[0]
	}

	if createNewCACertificate {

		// Create a self-sign root CA certificate
		var newCertificate ontap.Certificate
		newCertificate.CommonName = commonName
		newCertificate.Svm.Uuid = uuid
		newCertificate.Type = catype
		if expiryTime != "" {
			newCertificate.ExpiryTime = expiryTime
		} else {
			newCertificate.ExpiryTime = "P725DT" //magic words
		}

		jsonPayload, err := json.Marshal(newCertificate)
		if err != nil {
			//error creating the json body
			log.Error(err, fmt.Sprintf("Error creating the json payload for CA certificate creation %v", newCertificate.CommonName))
			return returnCert, err
		}

		log.Info("CA certificate creation attempt: " + newCertificate.CommonName)
		resp, err = oc.CreateCertificate(jsonPayload)
		if err != nil {
			log.Error(err, fmt.Sprintf("Error occurred when creating CA certificate: %v", newCertificate.CommonName))
			return returnCert, err
		}
		if resp.NumRecords != 0 {
			cert = resp.Records[0]
		}
	}

	// Certificate Signing Request
	var newCRS ontap.CertificateSigningRequest
	newCRS.SubjectName = "C=US,O=GATEWAY.NETAPP.COM,CN=" + svmName
	jsonPayload, err := json.Marshal(newCRS)
	if err != nil {
		//error creating the json body
		log.Error(err, fmt.Sprintf("Error creating the json payload for Certificate Signing Request %v", newCRS.SubjectName))
		return returnCert, err
	}

	log.Info("Certificate signing request creation attempt: " + newCRS.SubjectName)
	csr, err := oc.CreateCertificateSigningRequest(jsonPayload)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred when creating Certificate Signing Request: %v", newCRS.SubjectName))
		return returnCert, err
	}

	// Sign the CSR using certificate created earlier
	var newCRSSign ontap.CertificateSignRequest
	newCRSSign.SigningRequest = csr.Csr
	jsonPayload, err = json.Marshal(newCRSSign)
	if err != nil {
		//error creating the json body
		log.Error(err, fmt.Sprintf("Error creating the json payload for signing the Certificate Signing Request %v", newCRS.SubjectName))
		return returnCert, err
	}

	log.Info("Certificate Signing Request sign attempt: " + newCRS.SubjectName)
	signedCert, err := oc.CreateSignedCertificate(jsonPayload, cert.Uuid)
	if err != nil {
		log.Error(err, fmt.Sprintf("Error occurred when signing the Certificate Signing Request: %v", newCRS.SubjectName))
		return returnCert, err
	}

	//Install the signed certificate
	var newServerCertificate ontap.Certificate
	newServerCertificate.PublicCertificate = signedCert.PublicCertificate
	newServerCertificate.Svm.Uuid = uuid
	newServerCertificate.Type = "server" //magic words
	newServerCertificate.PrivateKey = csr.GeneratedPrivateKey
	jsonPayload, err = json.Marshal(newServerCertificate)
	if err != nil {
		//error creating the json body
		log.Error(err, "Error creating the json payload for server certificate creation")
		return returnCert, err
	}

	log.Info("Server certificate creation attempt")
	resp, err = oc.CreateCertificate(jsonPayload)
	if err != nil {
		log.Error(err, "Error occurred when creating server certificate")
		return returnCert, err
	}
	var finalCert ontap.Certificate
	if resp.NumRecords != 0 {
		finalCert = resp.Records[0]
	}

	return finalCert, nil
}
