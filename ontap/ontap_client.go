// From https://github.com/igor-feoktistov/go-ontap-rest/blob/main/ontap/client.go

package ontap

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	libraryVersion = "1.0.0"
	userAgent      = "go-ontap-rest/" + libraryVersion
)

type Client struct {
	client          *http.Client
	BaseURL         *url.URL
	UserAgent       string
	options         *ClientOptions
	ResponseTimeout time.Duration
}

type ClientOptions struct {
	BasicAuthUser     string
	BasicAuthPassword string
	SSLVerify         bool
	Debug             bool
	Timeout           time.Duration
}

type Resource struct {
	Name  string `json:"name,omitempty"`
	Uuid  string `json:"uuid,omitempty"`
	Links *struct {
		Self struct {
			Href string `json:"href,omitempty"`
		} `json:"self,omitempty"`
	} `json:"_links,omitempty"`
}

type NameReference struct {
	Name string `json:"name,omitempty"`
	Uuid string `json:"uuid,omitempty"`
}

func (r *Resource) GetRef() string {
	if r.Links != nil {
		return r.Links.Self.Href
	} else {
		return ""
	}
}

type BaseResponse struct {
	NumRecords int `json:"num_records"`
	Links      struct {
		Self struct {
			Href string `json:"href,omitempty"`
		} `json:"self,omitempty"`
		Next struct {
			Href string `json:"href,omitempty"`
		} `json:"next,omitempty"`
	} `json:"_links,omitempty"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
		Target  string `json:"target"`
	} `json:"error"`
}

type RestResponse struct {
	ErrorResponse ErrorResponse
	HttpResponse  *http.Response
}

func (res *BaseResponse) IsPaginate() bool {
	if res.NumRecords > 0 && len(res.Links.Next.Href) > 0 {
		return true
	} else {
		return false
	}
}

func (res *BaseResponse) GetNextRef() string {
	return res.Links.Next.Href
}

func DefaultOptions() *ClientOptions {
	return &ClientOptions{
		SSLVerify: true,
		Debug:     false,
		Timeout:   60 * time.Second,
	}
}

func NewClient(endpoint string, options *ClientOptions) *Client {
	if options == nil {
		options = DefaultOptions()
	}
	httpClient := &http.Client{
		Timeout: options.Timeout,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: !options.SSLVerify,
			},
		},
	}
	if !strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint + "/"
	}
	baseURL, _ := url.Parse(endpoint)
	c := &Client{
		client:          httpClient,
		BaseURL:         baseURL,
		UserAgent:       userAgent,
		options:         options,
		ResponseTimeout: options.Timeout,
	}
	return c
}

func (c *Client) NewRequest(method string, apiPath string, parameters []string, body interface{}) (req *http.Request, err error) {
	var payload io.Reader
	var extendedPath string
	if len(parameters) > 0 {
		extendedPath = fmt.Sprintf("%s?%s", apiPath, strings.Join(parameters, "&"))
	} else {
		extendedPath = apiPath
	}
	u, _ := c.BaseURL.Parse(extendedPath)
	if body != nil {
		buf, err := json.MarshalIndent(body, "", "  ")
		if err != nil {
			return nil, err
		}
		if c.options.Debug {
			log.Printf("[DEBUG] request JSON:\n%v\n\n", string(buf))
		}
		payload = bytes.NewBuffer(buf)
	}
	req, err = http.NewRequest(method, u.String(), payload)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", c.UserAgent)
	if c.options.BasicAuthUser != "" && c.options.BasicAuthPassword != "" {
		req.SetBasicAuth(c.options.BasicAuthUser, c.options.BasicAuthPassword)
	}
	if c.options.Debug {
		dump, _ := httputil.DumpRequestOut(req, true)
		log.Printf("[DEBUG] request dump:\n%q\n\n", dump)
	}
	return
}

