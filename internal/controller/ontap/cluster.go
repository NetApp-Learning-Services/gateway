package ontap

import (
	"encoding/json"
)

type Cluster struct {
	Name    string `json:"name"`
	UUID    string `json:"uuid"`
	Version struct {
		Full       string `json:"full"`
		Generation int    `json:"generation"`
		Major      int    `json:"major"`
		Minor      int    `json:"minor"`
	} `json:"version"`
	Links struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

func (c *Client) GetCluster() (cluster Cluster, err error) {
	uri := "/api/cluster"

	data, err := c.clientGet(uri)
	if err != nil {
		return cluster, &apiError{1, err.Error()}
	}

	var resp Cluster
	err = json.Unmarshal(data, &resp)
	if err != nil {
		return resp, &apiError{2, err.Error()}
	}

	return resp, nil
}
