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

type IpInterface struct {
	Name          string   `json:"name,omitempty"`
	Ip            Ip       `json:"ip,omitempty"`
	Location      Location `json:"location,omitempty"`
	ServicePolicy string   `json:"service_policy,omitempty"`
}

type Ip struct {
	Address string `json:"address,omitempty"`
	Netmask string `json:"netmask,omitempty"`
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
