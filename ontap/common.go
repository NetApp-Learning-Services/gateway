package ontap

import "net/http"

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Code    string `json:"code"`
		Target  string `json:"target:omitempty"`
	} `json:"error"`
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

type RestResponse struct {
	ErrorResponse ErrorResponse
	HttpResponse  *http.Response
}

type JobResponse struct {
	Job struct {
		Selflink SelfLink `json:"_links"`
		Uuid     string   `json:"uuid"`
	} `json:"job"`
}

type SelfLinks struct {
	Self struct {
		Href string `json:"href,omitempty"`
	} `json:"self,omitempty"`
}
