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
	Http             App              = "HTTP"
	Ontapi           App              = "ONTAPI"
	Ssh              App              = "SSH"
	Vsadmin          UserRole         = "vsadmin"
	Admin            UserRole         = "admin"
)

type Owner struct {
	Uuid string `json:"uuid,omitempty"`
	Name string `json:"name,omitempty"`
}

type Application struct {
	AppType          App         `json:"application,omitempty"`
	AuthMethods      AuthMethods `json:"authentication_methods,omitempty"`
	SecondAuthMethod string      `json:"second_authentication_methods,omitempty"`
}

type AuthMethods struct {
	Method []AuthMethodOption
}

type SecurityAccountPayload struct {
	Owner        Owner         `json:"owner,omitempty"`
	Name         string        `json:"name,omitempty"`
	Applications []Application `json:"applications,omitempty"`
	Role         UserRole      `json:"role,omitempty"`
	Password     string        `json:"password,omitempty"`
}

// Create SVM
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
