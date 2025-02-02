// From:  https://github.com/igor-feoktistov/go-ontap-rest/blob/main/ontap/svm.go
// From:  https://github.com/marstid/go-ontap-rest/blob/master/svm.go
// From:  https://github.com/marstid/go-ontap-rest/blob/master/svm_type.go

package ontap

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type AdDomain struct {
	OrganizationalUnit string `json:"organizational_unit,omitempty"`
	Fqdn               string `json:"fqdn,omitempty"`
	Password           string `json:"password,omitempty"`
	User               string `json:"user,omitempty"`
}

type Dns struct {
	Resource
	Domains []string `json:"domains,omitempty"`
	Servers []string `json:"servers,omitempty"`
}

type Ldap struct {
	Resource
	AdDomain string   `json:"ad_domain,omitempty"`
	BaseDn   string   `json:"base_dn,omitempty"`
	BindDn   string   `json:"bind_dn,omitempty"`
	Servers  []string `json:"servers,omitempty"`
	Enabled  bool     `json:"enabled"`
}

type Nfs struct {
	Resource
	Enabled bool `json:"enabled"`
}

type Iscsi struct {
	Resource
	Enabled bool `json:"enabled"`
}

type Fcp struct {
	Resource
	Enabled bool `json:"enabled"`
}

type Nvme struct {
	Resource
	Enabled bool `json:"enabled"`
}

type Cifs struct {
	Resource
	Name     string   `json:"name"`
	AdDomain AdDomain `json:"ad_domain,omitempty"`
	Enabled  bool     `json:"enabled"`
}

type Nis struct {
	Resource
	Domain  string   `json:"nis_domain,omitempty"`
	Servers []string `json:"nis_servers,omitempty"`
	Enabled bool     `json:"enabled"`
}

type NsSwitch struct {
	Resource
	Group    []string `json:"group,omitempty"`
	Hosts    []string `json:"hosts,omitempty"`
	NameMap  []string `json:"namemap,omitempty"`
	NetGroup []string `json:"netgroup,omitempty"`
	Passwd   []string `json:"passwd,omitempty"`
}

// type IpInfo struct {
// 	Address string `json:"address,omitempty"`
// 	Netmask string `json:"netmask,omitempty"`
// 	Family  string `json:"family,omitempty"`
// }

type FcPortReference struct {
	Resource
	Node string
}

type FcInterfaceSvm struct {
	Resource
	DataProtocol string `json:"data_protocol,omitempty"`
	Location     struct {
		Port FcPortReference `json:"port,omitempty"`
	} `json:"location,omitempty"`
}

// type NetworkRouteForSvmSvm struct {
// 	Gateway     string `json:"gateway,omitempty"`
// 	Destination IpInfo `json:"destination,omitempty"`
// }

type SnapMirror struct {
	IsProtected           bool `json:"is_protected"`
	ProtectedVolumesCount int  `json:"protected_volumes_count"`
}

type SelfLink struct {
	Self struct {
		Href string `json:"href"`
	} `json:"self"`
}

type SVMCreationPayload struct {
	Name         string                `json:"name,omitempty"`
	Comment      string                `json:"comment,omitempty"`
	State        string                `json:"state,omitempty"`
	IpInterfaces []IpInterfaceCreation `json:"ip_interfaces,omitempty"`
}

type SvmRef struct {
	Name string `json:"name,omitempty"`
	Uuid string `json:"uuid,omitempty"`
}

// type SvmResponse struct {
// 	Records []struct {
// 		Name  string `json:"name"`
// 		UUID  string `json:"uuid"`
// 		Links struct {
// 			Self struct {
// 				Href string `json:"href"`
// 			} `json:"self"`
// 		} `json:"_links"`
// 	} `json:"records"`
// 	NumRecords int `json:"num_records"`
// 	Links      struct {
// 		Self struct {
// 			Href string `json:"href"`
// 		} `json:"self"`
// 		Next struct {
// 			Href string `json:"href,omitempty"`
// 		} `json:"next,omitempty"`
// 	} `json:"_links"`
// }

type SvmResponse struct {
	BaseResponse
	Records []Svm `json:"records,omitempty"`
}

type Svm struct {
	Resource
	Uuid                   string        `json:"uuid,omitempty"`
	Name                   string        `json:"name,omitempty"`
	Subtype                string        `json:"subtype,omitempty"`
	Language               string        `json:"language,omitempty"`
	Aggregates             []Resource    `json:"aggregates,omitempty"`
	State                  string        `json:"state,omitempty"`
	Comment                string        `json:"comment,omitempty"`
	Ipspace                Resource      `json:"ipspace,omitempty"`
	Dns                    Dns           `json:"dns,omitempty"`
	Nsswitch               NsSwitch      `json:"nsswitch,omitempty"`
	Nis                    Nis           `json:"nis,omitempty"`
	Ldap                   Ldap          `json:"ldap,omitempty"`
	Nfs                    Nfs           `json:"nfs,omitempty"`
	Cifs                   Cifs          `json:"cifs,omitempty"`
	Iscsi                  Iscsi         `json:"iscsi,omitempty"`
	Fcp                    Fcp           `json:"fcp,omitempty"`
	Nvme                   Nvme          `json:"nvme,omitempty"`
	S3                     S3Service     `json:"s3,omitempty"`
	SnapMirror             SnapMirror    `json:"snapmirror,omitempty"`
	SnapshotPolicy         Resource      `json:"snapshot_policy,omitempty"`
	VolumeEfficiencyPolicy Resource      `json:"volume_efficiency_policy,omitempty"`
	IpInterfaces           []IpInterface `json:"ip_interfaces,omitempty"`
}

type Aggregate struct {
	Name string `json:"name"`
	UUID string `json:"uuid"`
}

type SvmByUUID struct {
	Uuid       string      `json:"uuid"`
	Name       string      `json:"name"`
	Subtype    string      `json:"subtype"`
	Language   string      `json:"language"`
	Aggregates []Aggregate `json:"aggregates"`
	State      string      `json:"state"`
	Comment    string      `json:"comment"`
	Ipspace    struct {
		Name  string `json:"name"`
		Uuid  string `json:"uuid"`
		Links struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"_links"`
	} `json:"ipspace"`
	SnapshotPolicy struct {
		Uuid  string `json:"uuid"`
		Name  string `json:"name"`
		Links struct {
			Self struct {
				Href string `json:"href"`
			} `json:"self"`
		} `json:"_links"`
	} `json:"snapshot_policy"`
	Nsswitch struct {
		Hosts    []string `json:"hosts"`
		Group    []string `json:"group"`
		Passwd   []string `json:"passwd"`
		Netgroup []string `json:"netgroup"`
		Namemap  []string `json:"namemap"`
	} `json:"nsswitch"`
	Nis struct {
		Enabled bool `json:"enabled"`
	} `json:"nis"`
	Ldap struct {
		Enabled bool `json:"enabled"`
	} `json:"ldap"`
	Nfs struct {
		Enabled bool `json:"enabled"`
	} `json:"nfs"`
	Cifs struct {
		Enabled bool `json:"enabled"`
	} `json:"cifs"`
	Iscsi struct {
		Enabled bool `json:"enabled"`
	} `json:"iscsi"`
	Fcp struct {
		Enabled bool `json:"enabled"`
	} `json:"fcp"`
	Nvme struct {
		Allowed bool `json:"allowed"`
		Enabled bool `json:"enabled"`
	} `json:"nvme"`
	Links SelfLinks `json:"_links"`
}

// In 9.11.1:
// Missing QOS
// Mssing Certificate
// Missing anti_ransomware_default_volume_state
// Missing is space reporting logical
// Missing is space enforcement logical
// Missing max volume
type SvmPatch struct {
	Resource
	Name         string        `json:"name,omitempty"`
	Language     string        `json:"language,omitempty"`
	Aggregates   []Resource    `json:"aggregates,omitempty"`
	State        string        `json:"state,omitempty"`
	Comment      string        `json:"comment,omitempty"`
	IpInterfaces []IpInterface `json:"ip_interfaces,omitempty"`
	Nvme         NvmePatch     `json:"nvme,omitempty"`
}

type NvmePatch struct {
	Resource
	Allowed bool `json:"allowed"`
}

// In 9.11.1:
// Missing QOS
// Mssing Certificate
// Missing anti_ransomware_default_volume_state
// Missing is space reporting logical
// Missing is space enforcement logical
// Missing max volume
type SvmAggregatePatch struct {
	Resource
	Aggregates []Resource `json:"aggregates,omitempty"`
}

// Return svm uuid from name
func (c *Client) GetStorageVmUUIDByName(name string) (uuid string, err error) {
	uri := "/api/svm/svms?name=" + name
	data, err := c.clientGet(uri)
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	records := result["records"].([]interface{})
	for _, v := range records {
		rec := v.(map[string]interface{})
		if rec["name"] == name {
			return rec["uuid"].(string), nil
		}

	}

	//return "", fmt.Errorf("0 - Storage VM with name %s not found", name)
	return "", &apiError{0, "0 - Storage VM with name " + name + "%s not found"}
}

// Return a SVM by UUID
func (c *Client) GetStorageVMByUUID(uuid string) (svm SvmByUUID, err error) {
	uri := "/api/svm/svms/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return svm, &apiError{1, err.Error()}
	}

	var resp SvmByUUID
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

// Create SVM
func (c *Client) CreateStorageVM(jsonPayload []byte) (uuid string, err error) {
	uri := "/api/svm/svms"
	r := ""
	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return r, &apiError{1, err.Error()}
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
		return r, &apiError{int64(createJob.Code), createJob.Message}
		//return fmt.Errorf("%d - %s", createJob.Code, createJob.Message)
	}

	if createJob.State == "success" {
		uuid, err = ParseUUID(createJob.Description, "/")
		return uuid, err
	}

	return r, nil
}

func (c *Client) PatchStorageVM(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/svm/svms/" + uuid

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

func (c *Client) DeleteStorageVM(uuid string) (err error) {
	uri := "/api/svm/svms/" + uuid

	data, err := c.clientDelete(uri)
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

	deleteJob, _ := c.GetJob(url)

	for deleteJob.State == "running" {
		time.Sleep(time.Second)
		deleteJob, _ = c.GetJob(url)
	}

	if deleteJob.State == "failure" {
		return &apiError{int64(deleteJob.Code), deleteJob.Message}
	}

	return nil
}

func ParseUUID(input string, char string) (string, error) {
	if len(input) == 0 {
		return "", &apiError{5, "UUID length is zero"}
	}

	//doesn't work with /auuid
	//split := strings.Split(input, " ")
	//return strings.Trim(split[1], trim), nil

	idx := strings.LastIndex(input, char)
	return input[idx+1:], nil //return the string starting at the last instance of char + 1 (not returning char)

}
