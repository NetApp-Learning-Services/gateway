package ontap

import (
	"encoding/json"
	"fmt"
	"strings"
)

type IpInterfaceCreation struct {
	Name          string   `json:"name,omitempty"`
	Ip            Ip       `json:"ip,omitempty"`
	Location      Location `json:"location,omitempty"`
	ServicePolicy string   `json:"service_policy,omitempty"`
}

type IpInterface struct {
	Name          string        `json:"name,omitempty"`
	Ip            Ip            `json:"ip,omitempty"`
	Location      Location      `json:"location,omitempty"`
	ServicePolicy ServicePolicy `json:"service_policy,omitempty"`
	State         string        `json:"state,omitempty"`
	Uuid          string        `json:"uuid,omitempty"`
	Scope         string        `json:"scope,omitempty"`
	Enabled       bool          `json:"enabled,omitempty"`
	Svm           SvmRef        `json:"svm,omitempty"`
}

type Ip struct {
	Address string `json:"address,omitempty"`
	Netmask string `json:"netmask,omitempty"`
	Family  string `json:"family,omitempty"`
}

type Location struct {
	BroadcastDomain BroadcastDomain `json:"broadcast_domain,omitempty"`
	HomeNode        HomeNode        `json:"home_node,omitempty"`
}

type BroadcastDomain struct {
	Name string `json:"name,omitempty"`
}

type HomeNode struct {
	Name string `json:"name,omitempty"`
}

type ServicePolicy struct {
	Links SelfLinks `json:"_links,omitempty"`
	Name  string    `json:"name,omitempty"`
}

type IpInterfacesResponse struct {
	BaseResponse
	Records []IpInterface `json:"records,omitempty"`
}

type IpServicePolicyResponse struct {
	BaseResponse
	Records []IpServicePolicy `json:"records,omitempty"`
}

type IpServicePolicy struct {
	Services []string `json:"services,omitempty"`
	Name     string   `json:"name,omitempty"`
	Ipspace  Ref      `json:"ipspace,omitempty"`
	Svm      SvmRef   `json:"svm,omitempty"`
	Scope    string   `json:"scope,omitempty"`
}

func (c *Client) GetIpInterfacesBySvmUuid(uuid string) (lifs IpInterfacesResponse, err error) {
	uri := "/api/network/ip/interfaces?svm.uuid=" + uuid

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

func (c *Client) GetIpInterfaceByLifUuid(uuid string) (lif IpInterface, err error) {
	uri := "/api/network/ip/interfaces/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return lif, &apiError{1, err.Error()}
	}

	var resp IpInterface
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateIpInterface(jsonPayload []byte) (err error) {
	uri := "/api/network/ip/interfaces"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) PatchIpInterface(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/network/ip/interfaces/" + uuid

	_, err = c.clientPatch(uri, jsonPayload)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return &apiError{404, fmt.Sprintf("LIF with UUID \"%s\" not found", uuid)}
		}
		if strings.Contains(err.Error(), "Duplicate") {
			return &apiError{1376963, err.Error()}
		}
		//miscellaneous errror
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) DeleteIpInterface(uuid string) (err error) {
	uri := "/api/network/ip/interfaces/" + uuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) CheckExistsInterfaceServicePolicyByName(servicePolicy string) (err error) {
	uri := "/api/network/ip/service-policies?name=" + servicePolicy

	data, err := c.clientGet(uri)
	if err != nil {
		//Error in GET request
		return &apiError{1, err.Error()}
	}
	var resp IpServicePolicyResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		//Error in response
		return &apiError{2, err.Error()}
	}
	if resp.NumRecords == 0 {
		//Service Policy not found
		return &apiError{3, "Lif service policy not found"}
	}

	// return nil if service policy name exists
	return nil
}

func (c *Client) CreateInterfaceServicePolicy(jsonPayload []byte) (err error) {
	uri := "/api/network/ip/service-policy"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}
