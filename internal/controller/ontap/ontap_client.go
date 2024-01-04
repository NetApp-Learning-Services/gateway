// From: https://github.com/marstid/go-ontap-rest/blob/05bcd8ec8c6d11a265a64d8f1187024400e0f1f5/ontap.go#L28
// From: https://github.com/igor-feoktistov/go-ontap-rest/blob/main/ontap/client.go

package ontap

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

const libraryVersion = "0.1"                        //special key
const userAgent = "astra.gateway/" + libraryVersion //special key
const defaultTimeout = 3                            // special key
const contentType = "application/hal+json"          //special key

type Client struct {
	UserName    string
	Password    string
	Host        string
	Debug       bool
	TrustSSL    bool
	TimeOut     time.Duration
	UserAgent   string
	ContentType string
}

type apiError struct {
	errorCode int64
	err       string
}

func (e *apiError) Error() string {
	return fmt.Sprintf("%d - API Error - %s", e.errorCode, e.err)
}
func (e *apiError) ErrorCode() int64 {
	return e.errorCode
}

func NewClient(user, password, host string, debug, trustSsl bool) (client *Client, error error) {

	return &Client{
		UserName:    user,
		Password:    password,
		Host:        host,
		Debug:       debug,
		TrustSSL:    trustSsl,
		TimeOut:     defaultTimeout,
		UserAgent:   userAgent,
		ContentType: contentType,
	}, error
}

// HTTP VERB FUNCS

func (c *Client) clientGet(uri string) (data []byte, err error) {

	url := "https://" + c.Host + uri

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] request type: GET")
		log.Printf("[DEBUG] request url: " + url)
	}

	response, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] response type: GET")
		log.Printf("[DEBUG] response url: " + url)
		log.Printf("[DEBUG] response body: " + fmt.Sprintf("%v", string(response[:])))
	}

	return response, nil
}

func (c *Client) clientPost(uri string, json []byte) (data []byte, err error) {

	url := "https://" + c.Host + uri

	payload := bytes.NewReader(json)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] request type: POST")
		log.Printf("[DEBUG] request url: " + url)
		log.Printf("[DEBUG] request payload: " + string(json))
	}

	response, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] response type: POST")
		log.Printf("[DEBUG] response url: " + url)
		log.Printf("[DEBUG] response body: " + fmt.Sprintf("%v", string(response[:])))
	}

	return response, nil
}

func (c *Client) clientPatch(uri string, json []byte) (data []byte, err error) {

	url := "https://" + c.Host + uri

	payload := bytes.NewReader(json)

	req, err := http.NewRequest("PATCH", url, payload)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] request type: PATCH")
		log.Printf("[DEBUG] request url: " + url)
		log.Printf("[DEBUG] request payload: " + string(json))
	}

	response, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] response type: PATCH")
		log.Printf("[DEBUG] response url: " + url)
		log.Printf("[DEBUG] response body: " + fmt.Sprintf("%v", string(response[:])))
	}

	return response, nil
}

func (c *Client) clientDelete(uri string) (data []byte, err error) {

	url := "https://" + c.Host + uri

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] request type: DELETE")
		log.Printf("[DEBUG] request url: " + url)
	}

	response, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if c.Debug {
		log.Printf("[DEBUG] response type: DELETE")
		log.Printf("[DEBUG] response url: " + url)
		log.Printf("[DEBUG] response body: " + fmt.Sprintf("%v", string(response[:])))
	}

	return response, nil
}

// Unified Do func

func (c *Client) doRequest(req *http.Request) ([]byte, error) {
	req.SetBasicAuth(c.UserName, c.Password)
	req.Header.Set("Content-Type", c.ContentType)
	req.Header.Set("UserAgent", c.UserAgent)

	httpClient := &http.Client{
		Timeout: time.Second * c.TimeOut,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.TrustSSL,
			},
		},
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {

		if strings.Contains(string(body), "message") {
			var jec ErrorResponse
			json.Unmarshal(body, &jec)
			if jec.Error.Code == "4" {
				return nil, fmt.Errorf("error-%s %s %s", jec.Error.Code, jec.Error.Target, jec.Error.Message)
			}
			//fmt.Println(string(body)) // todo: don't seem to need this
			return nil, fmt.Errorf("%s", jec.Error.Message)
		}
		return nil, fmt.Errorf("%s", body)

	}

	return body, nil

}
