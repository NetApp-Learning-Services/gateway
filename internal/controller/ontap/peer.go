package ontap

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type ClusterPeer struct {
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
	Records []ClusterPeer `json:"records,omitempty"`
}

type SvmPeer struct {
	Name         string           `json:"name"`
	Uuid         string           `json:"uuid,omitempty"`
	LocalSvm     SvmRef           `json:"svm,omitempty"`
	State        string           `json:"state,omitempty"`
	Applications []string         `json:"applications,omitempty"`
	Peer         PeerRelationship `json:"peer,omitempty"`
}

type PeerRelationship struct {
	Svm     SvmRef     `json:"svm,omitempty"`
	Cluster ClusterRef `json:"cluster,omitempty"`
}

type ClusterRef struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

type SvmPeersResponse struct {
	BaseResponse
	Records []SvmPeer `json:"records,omitempty"`
}

type SvmPeerPatch struct {
	State string `json:"state"`
}

//const returnPeerRecords string = "?return_records=true"

func (c *Client) GetClusterPeers() (clusterPeers ClusterPeersResponse, err error) {
	uri := "/api/cluster/peers?order_by=name&fields=remote,status,uuid,authentication,encryption"

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
		return clusterPeers, errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "no cluster peers")
	}

	return resp, nil
}

func (c *Client) CreateClusterPeer(jsonPayload []byte) (err error) {
	uri := "/api/cluster/peers"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
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

func (c *Client) GetSvmPeer(localSvm string) (svmPeers SvmPeersResponse, err error) {
	uri := "/api/svm/peers?svm.name=" + localSvm

	data, err := c.clientGet(uri)
	if err != nil {
		return svmPeers, &apiError{1, err.Error()}
	}

	var resp SvmPeersResponse
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	if resp.NumRecords == 0 {
		return svmPeers, errors.NewNotFound(schema.GroupResource{Group: "gateway.netapp.com", Resource: "StorageVirtualMachine"}, "no svm peers")
	}

	return resp, nil
}

func (c *Client) CreateSvmPeer(jsonPayload []byte) (err error) {
	uri := "/api/svm/peers"
	_, err = c.clientPost(uri, jsonPayload)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) DeleteSvmPeer(uuid string) (err error) {
	uri := "/api/svm/peers/" + uuid

	_, err = c.clientDelete(uri)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}

func (c *Client) PatchSvmPeer(jsonPayload []byte, uuid string) (err error) {
	uri := "/api/svm/peers/" + uuid

	_, err = c.clientPatch(uri, jsonPayload)
	if err != nil {
		return &apiError{1, err.Error()}
	}

	return nil
}
