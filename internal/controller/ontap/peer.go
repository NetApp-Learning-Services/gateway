package ontap

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClusterPeerService struct {
	Name               string             `json:"name"`
	Uuid               string             `json:"uuid,omitempty"`
	InitialAllowedSVMs []SvmRef           `json:"initial_allowed_svms,omitempty"`
	Remote             PeerRemote         `json:"remote,omitempty"`
	Applications       []string           `json:"peer_applications,omitempty"`
	Status             PeerStatus         `json:"status,omitempty"`
	Encryption         PeerEncryption     `json:"encryption,omitempty"`
	Authentication     PeerAuthentication `json:"authentication,omitempty"`
}

type PeerRemote struct {
	Name      string   `json:"name,omitempty"`
	Uuid      string   `json:"uuid,omitempty"`
	Addresses []string `json:"ip_addresses,omitempty"`
}

type PeerStatus struct {
	State      string `json:"state,omitempty"`
	UpdateTime string `json:"update_time,omitempty"`
}

type PeerEncryption struct {
	Proposed string `json:"proposed,omitempty"`
	State    string `json:"state,omitempty"`
}

type PeerAuthentication struct {
	Passphrase string `json:"passphrase,omitempty"`
	ExpiryTime string `json:"expiry_time,omitempty"`
	State      string `json:"state,omitempty"`
}

type ClusterPeersResponse struct {
	BaseResponse
	Records []ClusterPeerService `json:"records,omitempty"`
}

//const returnPeerRecords string = "?return_records=true"

func (c *Client) GetClusterPeerServicesForCluster(remoteIp string) (clusterPeers ClusterPeersResponse, err error) {
	uri := "/api/cluster/peers?ip_address=" + remoteIp

	data, err := c.clientGet(uri)
	if err != nil {
		return clusterPeers, &apiError{1, err.Error()}
	}

	var resp ClusterPeersResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	if resp.NumRecords == 0 {
		return clusterPeers, errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "no peers")
	}

	return resp, nil
}

func (c *Client) CreateClusterPeerService(jsonPayload []byte) (err error) {
	uri := "/api/cluster/peers"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) DeleteClusterPeer(uuid string) (err error) {
	uri := "/api/cluster/peers/" + uuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}
