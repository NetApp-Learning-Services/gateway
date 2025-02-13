package ontap

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type NFSService struct {
	Enabled  *bool       `json:"enabled,omitempty"`
	Protocol NFSProtocol `json:"protocol,omitempty"`
	Svm      SvmRef      `json:"svm,omitempty"`
}

type NFSProtocol struct {
	V3Enable  *bool `json:"v3_enabled,omitempty"`
	V4Enable  *bool `json:"v40_enabled,omitempty"`
	V41Enable *bool `json:"v41_enabled,omitempty"`
}

type ExportPolicyUpsert struct {
	Name  string       `json:"name,omitempty"`
	Svm   SvmRef       `json:"svm,omitempty"`
	Rules []ExportRule `json:"rules,omitempty"`
}

type ExportPolicy struct {
	Name  string       `json:"name,omitempty"`
	Svm   SvmRef       `json:"svm,omitempty"`
	Rules []ExportRule `json:"rules,omitempty"`
	Id    int          `json:"id,omitempty"`
}

// Requires ONTAP 9.10?
type ExportRule struct {
	Protocols []string      `json:"protocols,omitempty"`
	RwRule    []string      `json:"rw_rule,omitempty"`
	RoRule    []string      `json:"ro_rule,omitempty"`
	Superuser []string      `json:"superuser,omitempty"`
	Anonuser  string        `json:"anonymous_user,omitempty"`
	Clients   []ExportMatch `json:"clients,omitempty"`
}

type ExportMatch struct {
	Match string `json:"match,omitempty"`
}

type ExportResponse struct {
	BaseResponse
	Records []ExportPolicy `json:"records,omitempty"`
}

const returnNFSRecords string = "?return_timeout=120&max_records=40&fields=*"

func (c *Client) GetNfsServiceBySvmUuid(uuid string) (nfsService NFSService, err error) {
	uri := "/api/protocols/nfs/services/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		if strings.Contains(err.Error(), "entry doesn't exist") {
			return nfsService, errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "no nfs")
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

func (c *Client) GetNfsExportBySvmUuid(uuid string) (exports ExportResponse, err error) {
	uri := "/api/protocols/nfs/export-policies" + returnNFSRecords + "&svm.uuid=" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return exports, &apiError{1, err.Error()}
	}

	var resp ExportResponse
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

func (c *Client) PatchNfsExport(id int, jsonPayload []byte) (err error) {
	uri := "/api/protocols/nfs/export-policies/" + strconv.Itoa(id)

	_, err = c.clientPatch(uri, jsonPayload)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return &apiError{404, fmt.Sprintf("Export with ID \"%s\" not found", strconv.Itoa(id))}
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

func (c *Client) DeleteNfsExport(id int) (err error) {
	uri := "/api/protocols/nfs/export-policies/" + strconv.Itoa(id)

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) GetNfsInterfacesBySvmUuid(uuid string) (lifs IpInterfacesResponse, err error) {
	uri := "/api/network/ip/interfaces" + returnNFSRecords + "&service_policy.name=default-data-files&svm.uuid=" + uuid

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
