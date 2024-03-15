package ontap

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NvmeService struct {
	Svm     SvmRef `json:"svm,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}

const returnNvmeQs string = "?return_records=true"

func (c *Client) GetNvmeServiceBySvmUuid(uuid string) (nvmeService NvmeService, err error) {
	uri := "/api/protocols/nvme/services/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		if strings.Contains(err.Error(), "An NVMe service does not exist") {
			return nvmeService, errors.NewNotFound(schema.GroupResource{Group: "gatewayv1beta1", Resource: "StorageVirtualMachine"}, "no nvme")
		}
		return nvmeService, &apiError{1, err.Error()}
	}

	var resp NvmeService
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateNvmeService(jsonPayload []byte) (err error) {
	uri := "/api/protocols/nvme/services" + returnNvmeQs
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) PatchNvmeService(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/protocols/nvme/services/" + uuid

	_, err = c.clientPatch(uri, jsonPayload)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return &apiError{404, fmt.Sprintf("SVM with UUID \"%s\" not found", uuid)}
		}
		if strings.Contains(err.Error(), "not running") {
			return &apiError{3276916, err.Error()}
		}
		//miscellaneous errror
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) DeleteNvmeService(uuid string) (err error) {
	uri := "/api/protocols/nvme/services/" + uuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) GetNvmeInterfacesBySvmUuid(uuid string, servicePolicy string) (lifs IpInterfacesResponse, err error) {
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

func (c *Client) GetNvmeServicePolicyByName(servicePolicy string) (err error) {
	uri := "/api/network/ip/service-policies?name=" + servicePolicy

	_, err = c.clientGet(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) CreateNvmeServicePolicy(jsonPayload []byte) (err error) {
	uri := "/api/network/ip/service-policies"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}
	return nil
}
