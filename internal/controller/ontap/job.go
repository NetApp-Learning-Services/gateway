package ontap

import (
	"encoding/json"
	"fmt"
	"time"
)

type Job struct {
	UUID        string    `json:"uuid"`
	Description string    `json:"description"`
	State       string    `json:"state"`
	Message     string    `json:"message"`
	Code        int       `json:"code"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Links       struct {
		Self struct {
			Href string `json:"href"`
		} `json:"self"`
	} `json:"_links"`
}

func (c *Client) GetJob(url string) (job Job, err error) {

	data, err := c.clientGet(url)
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return job, err
	}

	err = json.Unmarshal(data, &job)
	if err != nil {
		return job, err
	}

	return job, nil

}
