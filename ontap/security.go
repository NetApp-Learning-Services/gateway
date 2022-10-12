package ontap

import (
	"encoding/json"
	"fmt"
)

type AuthMethodOption string
type App string
type UserRole string

const (
	Password         AuthMethodOption = "password"
	Domain           AuthMethodOption = "domain"
	Nsswitch         AuthMethodOption = "nsswitch"
	Certificate      AuthMethodOption = "certificate"
	Amqp             App              = "amqp"
	Console          App              = "console"
	ServiceProcessor App              = "service_processor"
	Http             App              = "http"
	Ontapi           App              = "ontapi"
	Ssh              App              = "ssh"
	Vsadmin          UserRole         = "vsadmin"
	Admin            UserRole         = "admin"
)

type Owner struct {
	Uuid string `json:"uuid,omitempty"`
	Name string `json:"name,omitempty"`
}

type Application struct {
	AppType          App                `json:"application,omitempty"`
	AuthMethods      []AuthMethodOption `json:"authentication_methods,omitempty"`
	SecondAuthMethod string             `json:"second_authentication_method,omitempty"`
}

type SecurityAccountPayload struct {
	Owner        Owner         `json:"owner,omitempty"`
	Name         string        `json:"name,omitempty"`
	Applications []Application `json:"applications,omitempty"`
	Role         UserRole      `json:"role,omitempty"`
	Password     string        `json:"password,omitempty"`
	Comment      string        `json:"comment,omitempty"`
	Locked       bool          `json:"locked,omitempty"`
}

// Create securtiy account
func (c *Client) CreateSecurityAccount(jsonPayload []byte) (err error) {
	uri := "/api/security/accounts"
	data, err := c.clientPost(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return &apiError{2, err.Error()}
	} else {
		fmt.Println(fmt.Sprintf("result: %v", result))
	}

	return nil
}

// Patch securtiy account
func (c *Client) PatchSecurityAccount(jsonPayload []byte, uuid string, name string) (err error) {
	uri := "/api/security/accounts/" + uuid + "/" + name
	data, err := c.clientPatch(uri, jsonPayload)
	if err != nil {
		//fmt.Println("Error: " + err.Error())
		return &apiError{1, err.Error()}
	}

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return &apiError{2, err.Error()}
	} else {
		fmt.Println(fmt.Sprintf("result: %v", result))
	}

	return nil
}
