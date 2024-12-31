package ontap

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type S3Service struct {
	Svm            SvmRef   `json:"svm,omitempty"`
	Certificate    Resource `json:"certificate,omitempty"`
	IsHttpEnabled  bool     `json:"is_http_enabled"`
	IsHttpsEnabled bool     `json:"is_https_enabled"`
	Port           int      `json:"port,omitempty"`
	SecurePort     int      `json:"secure_port,omitempty"`
	Enabled        bool     `json:"enabled"`
	Name           string   `json:"name"`
}

type S3User struct {
	Name      string `json:"name,omitempty"`
	Svm       SvmRef `json:"svm,omitempty"`
	AccessKey string `json:"access_key,omitempty"`
	SecretKey string `json:"secret_key,omitempty"`
}

type S3UsersResponse struct {
	BaseResponse
	Records []S3User `json:"records,omitempty"`
}

type S3Bucket struct {
	Name    string `json:"name,omitempty"`
	Svm     SvmRef `json:"svm,omitempty"`
	Size    int    `json:"size,omitempty"`
	Type    string `json:"type,omitempty"`
	Comment string `json:"comment,omitempty"`
	Uuid    string `json:"uuid,omitempty"`
}

type S3BucketsResponse struct {
	BaseResponse
	Records []S3Bucket `json:"records,omitempty"`
}

type S3BucketJobResponse struct {
	JobResponse
	Uuid string `json:"uuid,omitempty"`
}

const returnS3Records string = "?return_records=true"

func (c *Client) GetS3ServiceBySvmUuid(uuid string) (s3Service S3Service, err error) {
	uri := "/api/protocols/s3/services/" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		if strings.Contains(err.Error(), "exist") {
			return s3Service, errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "no s3")
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
	uri := "/api/protocols/s3/services" + returnS3Records
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
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
	uri := "/api/network/ip/interfaces" + returnNFSRecords + "&service_policy.name=" + servicePolicy + "&svm.uuid=" + uuid

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
		return &apiError{1, err.Error()}
	}
	return nil
}

func (c *Client) GetS3UsersBySvmUuid(uuid string) (users S3UsersResponse, err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/users"

	data, err := c.clientGet(uri)
	if err != nil {
		return users, &apiError{1, err.Error()}
	}

	var resp S3UsersResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateS3User(uuid string, jsonPayload []byte) (users S3UsersResponse, err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/users"

	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		return users, &apiError{1, err.Error()}
	}

	var resp S3UsersResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) DeleteS3User(uuid string, name string) (err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/users/" + name

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) GetS3BucketsBySvmUuid(uuid string) (users S3BucketsResponse, err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/buckets"

	data, err := c.clientGet(uri)
	if err != nil {
		return users, &apiError{1, err.Error()}
	}

	var resp S3BucketsResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateS3Bucket(uuid string, jsonPayload []byte) (err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/buckets"

	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
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
	}

	if createJob.State == "success" {
		//_, err = ParseUUID(createJob.Description, "/")
		return err
	}

	return nil
}

func (c *Client) DeleteS3Bucket(uuid string, bucketUuid string) (err error) {
	uri := "/api/protocols/s3/services/" + uuid + "/buckets/" + bucketUuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}
