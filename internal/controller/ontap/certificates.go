package ontap

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Certificate struct {
	Svm                SvmRef `json:"svm,omitempty"`
	Type               string `json:"type,omitempty"`
	PublicCertificate  string `json:"public_certificate,omitempty"`
	PrivateCertificate string `json:"private_certificate,omitempty"`
	KeySize            int    `json:"key_size,omitempty"`
	ExpiryTime         string `json:"expiry_time,omitempty"`
	Name               string `json:"name"`
	CommonName         string `json:"common_name"`
	SerialNumber       string `json:"serial_number,omitempty"`
	Uuid               string `json:"uuid"`
}

type CertificateSigningRequest struct {
	SubjectName string `json:"subject_name"`
}

type CertificateSigningResponse struct {
	Csr                 string `json:"csr"`
	GeneratedPrivateKey string `json:"generated_private_key"`
	SubjectName         string `json:"subject_name"`
}

type CertificateSignRequest struct {
	SigningRequest string `json:"signing_request"`
}

type CertificateSignResponse struct {
	PublicCertificate string `json:"public_certificate"`
}

type CertificateResponse struct {
	BaseResponse
	Records []Certificate `json:"records,omitempty"`
}

const returnCertificateQs string = "?return_timeout=120&max_records=40&fields="

func (c *Client) GetCertificatesBySvmUuid(uuid string, commonName string, caType string) (certs CertificateResponse, err error) {
	uri := "/api/security/certificates" + returnCertificateQs + "common_name=" + commonName + "&svm.uuid=" + uuid + "&type=" + caType

	data, err := c.clientGet(uri)
	if err != nil {
		return certs, &apiError{1, err.Error()}
	}

	var resp CertificateResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	if resp.NumRecords == 0 {
		//No certificate found
		return resp, errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "no S3 certificate")
	}

	return resp, nil
}

func (c *Client) CreateCertificate(jsonPayload []byte) (cert Certificate, err error) {
	uri := "/api/security/certificates"
	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return cert, &apiError{1, err.Error()}
	}

	var resp Certificate
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateCertificateSigningRequest(jsonPayload []byte) (csr CertificateSigningResponse, err error) {
	uri := "/api/security/certificate-signing-request"
	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return csr, &apiError{1, err.Error()}
	}

	var resp CertificateSigningResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateSignedCertificate(jsonPayload []byte, ca_uuid string) (cert CertificateSignResponse, err error) {
	uri := "/api/security/certificate/" + ca_uuid + "/sign"
	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return cert, &apiError{1, err.Error()}
	}

	var resp CertificateSignResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}
