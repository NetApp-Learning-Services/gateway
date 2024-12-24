package ontap

import "encoding/json"

type Certificate struct {
	Svm                SvmRef  `json:"svm,omitempty"`
	Type               *string `json:"type,omitempty"`
	PublicCertificate  *string `json:"public_certificate,omitempty"`
	PrivateCertificate *string `json:"private_certificate,omitempty"`
	KeySize            int     `json:"key_size,omitempty"`
	ExpiryTime         *string `json:"expiry_time,omitempty"`
	Name               *string `json:"name"`
	CommonName         *string `json:"common_name"`
	SerialNumber       *string `json:"serial_number,omitempty"`
}

type SelfSigningRequest struct {
	SubjectName *string `json:"subject_name"`
}

type CertificateResponse struct {
	BaseResponse
	Records []Certificate `json:"records,omitempty"`
}

//const returnCertificateQs string = "?return_records=true"

func (c *Client) GetCertificatesBySvmUuid(uuid string, commonName string) (certs CertificateResponse, err error) {
	uri := "/api/securtiy/certificates" + qs + "common_name=" + commonName + "&svm.uuid=" + uuid

	data, err := c.clientGet(uri)
	if err != nil {
		return certs, &apiError{1, err.Error()}
	}

	var resp CertificateResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}

func (c *Client) CreateCertificate(jsonPayload []byte) (err error) {
	uri := "/api/security/certificates"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) CreateSelfSigningRequest(jsonPayload []byte) (err error) {
	uri := "/api/security/certificate-signing-request"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) CreateSignedCertificate(jsonPayload []byte, uuid string) (err error) {
	uri := "/api/security/certificate/" + uuid + "/sign"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}
