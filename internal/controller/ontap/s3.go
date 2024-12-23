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
	IsHttpEnabled  *bool    `json:"is_http_enabled"`
	IsHttpsEnabled *bool    `json:"is_https_enabled"`
	Port           int      `json:"port"`
	SecurePort     int      `json:"secure_port"`
	Enabled        *bool    `json:"enabled"`
	Name           *string  `json:"name"`
}

type UserResponse struct {
	BaseResponse
	Records []S3User `json:"records,omitempty"`
}

type S3User struct {
	Name      string `json:"name,omitempty"`
	Svm       SvmRef `json:"svm,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

const returnS3Qs string = "?return_records=true"

func (c *Client) GetS3ServiceBySvmUuid(uuid string) (s3Service S3Service, err error) {
	uri := "/api/protocols/s3/services/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		if strings.Contains(err.Error(), "entry doesn't exist") {
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

func (c *Client) CreateS3ServicePolicy(jsonPayload []byte) (err error) {
	uri := "/api/network/ip/service-policies"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}
	return nil
}

func (c *Client) CreateS3User(uuid string, jsonPayload []byte) (users UserResponse, err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/users"

	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return users, &apiError{1, err.Error()}
	}

	var resp UserResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) GetS3UsersBySvmUuid(uuid string) (users UserResponse, err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/users"

	data, err := c.clientGet(uri)
	if err != nil {
		return users, &apiError{1, err.Error()}
	}

	var resp UserResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}
