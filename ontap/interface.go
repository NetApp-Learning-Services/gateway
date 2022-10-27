package ontap

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type IpInterfaceCreation struct {
	Name          string   `json:"name,omitempty"`
	Ip            Ip       `json:"ip,omitempty"`
	Location      Location `json:"location,omitempty"`
	ServicePolicy string   `json:"service_policy,omitempty"`
}

type IpInterface struct {
	Name          string        `json:"name,omitempty"`
	Ip            Ip            `json:"ip"`
	Location      Location      `json:"location,omitempty"`
	ServicePolicy ServicePolicy `json:"service_policy,omitempty"`
	State         string        `json:"state,omitempty"`
	Uuid          string        `json:"uuid,omitempty"`
	Scope         string        `json:"scope,omitempty"`
	Enabled       bool          `json:"enabled,omitempty"`
}

type Ip struct {
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Family  string `json:"family"`
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
	Links SelfLinks `json:"_links"`
	Name  string    `json:"name,omitempty"`
	Uuid  string    `json:"uuid,omitempty"`
}

type IPInterfacesResponse struct {
	BaseResponse
	Records []IpInterface `json:"records,omitempty"`
}

func (c *Client) GetInterfacesForSVMByUUID(uuid string) (lifs IPInterfacesResponse, err error) {
	uri := "/api/network/ip/interfaces?svm.uuid=" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return lifs, &apiError{1, err.Error()}
	}

	var resp IPInterfacesResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) GetInterfacesByUUID(uuid string) (lifs IPInterfacesResponse, err error) {
	uri := "/api/network/ip/interfaces?svm.uuid=" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return lifs, &apiError{1, err.Error()}
	}

	var resp IPInterfacesResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateInterface(jsonPayload []byte) (err error) {
	uri := "/api/network/ip/interfaces"
	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	var result JobResponse
	json.Unmarshal(data, &result)

	url := result.Job.Selflink.Self.Href

	createJob, _ := c.GetJob(url)

	for createJob.State == "running" {
		time.Sleep(time.Second * 2)
		createJob, _ = c.GetJob(url)
	}

	if createJob.State == "failure" {
		return &apiError{int64(createJob.Code), createJob.Message}
		//return fmt.Errorf("%d - %s", createJob.Code, createJob.Message)
	}

	return nil
}

func (c *Client) PatchInterface(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/network/ip/interfaces/" + uuid

	data, err := c.clientPatch(uri, jsonPayload)
	if err != nil {
		if strings.Contains(err.Error(), "Error-4") {
			return &apiError{4, fmt.Sprintf("SVM with UUID \"%s\" not found", uuid)}
		}
		return &apiError{1, err.Error()}
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return &apiError{2, err.Error()}
	}

	job := result["job"].(map[string]interface{})
	link := job["_links"].(map[string]interface{})
	href := link["self"].(map[string]interface{})
	url := href["href"].(string)

	patchJob, _ := c.GetJob(url)

	for patchJob.State == "running" {
		time.Sleep(time.Second)
		patchJob, _ = c.GetJob(url)
	}

	if patchJob.State == "failure" {
		return &apiError{int64(patchJob.Code), patchJob.Message}
	}

	if patchJob.State == "success" {
		//uuid, err = ParseUUID(patchJob.Description, " ")
		//return uuid, err
		return nil
	}

	return nil
}
