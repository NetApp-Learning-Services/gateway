package ontap

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
