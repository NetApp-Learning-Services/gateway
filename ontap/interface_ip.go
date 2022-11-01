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
	Scope         string   `json:"scope,omitempty"`
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
	Svm           SvmId         `json:"svm,omitempty"`
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
	Uuid  string    `json:"uuid,omitempty"`
}

type IPInterfacesResponse struct {
	BaseResponse
	Records []IpInterface `json:"records,omitempty"`
}

type SvmId struct {
	Name string `json:"name,omitempty"`
	Uuid string `json:"uuid,omitempty"`
}

func (c *Client) GetIpInterfacesBySvmUuid(uuid string) (lifs IPInterfacesResponse, err error) {
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

	// var result JobResponse
	// json.Unmarshal(data, &result)

	// url := result.Job.Selflink.Self.Href

	// createJob, _ := c.GetJob(url)

	// for createJob.State == "running" {
	// 	time.Sleep(time.Second * 2)
	// 	createJob, _ = c.GetJob(url)
	// }

	// if createJob.State == "failure" {
	// 	return &apiError{int64(createJob.Code), createJob.Message}
	// 	//return fmt.Errorf("%d - %s", createJob.Code, createJob.Message)
	// }

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

	// var result map[string]interface{}
	// err = json.Unmarshal(data, &result)
	// if err != nil {
	// 	return &apiError{2, err.Error()}
	// }

	// job := result["job"].(map[string]interface{})
	// link := job["_links"].(map[string]interface{})
	// href := link["self"].(map[string]interface{})
	// url := href["href"].(string)

	// patchJob, _ := c.GetJob(url)

	// for patchJob.State == "running" {
	// 	time.Sleep(time.Second)
	// 	patchJob, _ = c.GetJob(url)
	// }

	// if patchJob.State == "failure" {
	// 	return &apiError{int64(patchJob.Code), patchJob.Message}
	// }

	// if patchJob.State == "success" {
	// 	//uuid, err = ParseUUID(patchJob.Description, " ")
	// 	//return uuid, err
	// 	return nil
	// }

	return nil
}
