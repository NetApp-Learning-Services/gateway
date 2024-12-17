package ontap

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type S3Service struct {
	Svm            SvmRef   `json:"svm,omitempty"`
	Certificate    Resource `json:"certificate,omitempty"`
	IsHttpEnabled  bool     `json:"is_http_enabled"`
	IsHttpsEnabled bool     `json:"is_https_enabled"`
	Port           int      `json:"port"`
	SecurePort     int      `json:"secure_port"`
	Enabled        *bool    `json:"enabled"`
}

const returnS3Qs string = "?return_records=true"

func (c *Client) GetS3ServiceBySvmUuid(uuid string) (s3Service S3Service, err error) {
	uri := "/api/protocols/s3/services/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		if strings.Contains(err.Error(), "An S3 service does not exist") {
			return s3Service, errors.NewNotFound(schema.GroupResource{Group: "gatewayv1beta2", Resource: "StorageVirtualMachine"}, "no s3")
		}
		return s3Service, &apiError{1, err.Error()}
	}

	var resp S3Service
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateS3Service(jsonPayload []byte) (err error) {
	uri := "/api/protocols/s3/services" + returnS3Qs
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) PatchS3Service(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/protocols/s3/services/" + uuid

	_, err = c.clientPatch(uri, jsonPayload)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return &apiError{404, fmt.Sprintf("SVM with UUID \"%s\" not found", uuid)}
		}
		if strings.Contains(err.Error(), "name") {
			return &apiError{92405790, err.Error()}
		}
		//miscellaneous errror
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) DeleteS3Service(uuid string) (err error) {
	uri := "/api/protocols/s3/services/" + uuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) GetS3InterfacesBySvmUuid(uuid string, servicePolicy string) (lifs IpInterfacesResponse, err error) {
	uri := "/api/network/ip/interfaces" + qs + "&service_policy.name=" + servicePolicy + "&svm.uuid=" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return lifs, &apiError{1, err.Error()}
	}

	var resp IpInterfacesResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) GetS3ServicePolicyByName(servicePolicy string) (err error) {
	uri := "/api/network/ip/service-policies?name=" + servicePolicy

	_, err = c.clientGet(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) CreateS3ServicePolicy(jsonPayload []byte) (err error) {
	uri := "/api/network/ip/service-policies"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}
	return nil
}
