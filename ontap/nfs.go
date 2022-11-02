package ontap

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NFSService struct {
	Enabled  *bool       `json:"enabled,omitempty"`
	Protocol NFSProtocol `json:"protocol,omitempty"`
	Svm      NfsSvm      `json:"svm,omitempty"`
}

type NFSProtocol struct {
	V3Enable  *bool `json:"v3_enabled,omitempty"`
	V4Enable  *bool `json:"v40_enabled,omitempty"`
	V41Enable *bool `json:"v41_enabled,omitempty"`
}

type NfsSvm struct {
	Name string `json:"name,omitempty"`
	Uuid string `json:"uuid,omitempty"`
}

type ExportPolicyUpsert struct {
	Name  string       `json:"name,omitempty"`
	Svm   NfsSvm       `json:"svm,omitempty"`
	Rules []ExportRule `json:"rules,omitempty"`
}

type ExportPolicy struct {
	Name  string       `json:"name,omitempty"`
	Svm   NfsSvm       `json:"svm,omitempty"`
	Rules []ExportRule `json:"rules,omitempty"`
	Id    int16        `json:"id,omitempty"`
}

// Requires ONTAP 9.10?
type ExportRule struct {
	Protocols string `json:"protocols,omitempty"`
	RwRule    string `json:"rw_rule,omitempty"`
	RoRule    string `json:"ro_rule,omitempty"`
	Superuser string `json:"superuser,omitempty"`
	Anonuser  string `json:"anonymous_user,omitempty"`
}

type ExportResponse struct {
	BaseResponse
	Records []Svm `json:"records,omitempty"`
}

func (c *Client) GetNfsServiceBySvmUuid(uuid string) (nfsService NFSService, err error) {
	uri := "/api/protocols/nfs/services/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		if strings.Contains(err.Error(), "entry doesn't exist") {
			return nfsService, errors.NewNotFound(schema.GroupResource{Group: "gatewayv1alpha1", Resource: "StorageVirtualMachine"}, "no nfs")
		}
		return nfsService, &apiError{1, err.Error()}
	}

	var resp NFSService
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateNfsService(jsonPayload []byte) (err error) {
	uri := "/api/protocols/nfs/services"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) PatchNfsService(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/protocols/nfs/services/" + uuid

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

func (c *Client) DeleteNfsService(uuid string) (err error) {
	uri := "/api/protocols/nfs/services/" + uuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) GetNfsExportBySvmUuid(uuid string) (exports ExportPolicy, err error) {
	uri := "/api/protocols/nfs/export-policies?svm.uuid=" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return exports, &apiError{1, err.Error()}
	}

	var resp ExportPolicy
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateNfsExport(jsonPayload []byte) (err error) {
	uri := "/api/protocols/nfs/export-policies"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) PatchNfsExport(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/protocols/nfs/export-policies/" + uuid

	_, err = c.clientPatch(uri, jsonPayload)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return &apiError{404, fmt.Sprintf("Export with UUID \"%s\" not found", uuid)}
		}
		if strings.Contains(err.Error(), "Failed to rename") {
			return &apiError{1703950, err.Error()}
		}
		if strings.Contains(err.Error(), "No spaces") {
			return &apiError{1703952, err.Error()}
		}
		//miscellaneous errror
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) DeleteNfsExport(uuid string) (err error) {
	uri := "/api/protocols/nfs/export-policies/" + uuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) GetNfsInterfacesBySvmUuid(uuid string) (lifs IpInterfacesResponse, err error) {
	uri := "/api/network/ip/interfaces?service_policy.name=default-data-files&fields=enabled&svm.uuid=" + uuid

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
