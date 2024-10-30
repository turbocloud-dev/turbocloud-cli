package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const baseUrl = "http://localhost:5445/"

type Machine struct {
	Id             string
	VPNIp          string //IP inside VPN
	PublicIp       string //Public Ip
	CloudPrivateIp string //Private Ip inside data center
	Name           string
	Types          []string
	Status         string
	Domains        []string
	JoinURL        string
	PublicSSHKey   string
}

/*Machines*/
func getMachines() tea.Msg {

	// Create an HTTP client and make a GET request.
	c := &http.Client{Timeout: 10 * time.Second}
	res, err := c.Get(baseUrl + "machine")

	if err != nil {
		// There was an error making our request. Wrap the error we received
		// in a message and return it.
		return errMsg{err}
	}

	defer res.Body.Close()

	// We received a response from the server. Return the HTTP status code
	// as a message.
	dec := json.NewDecoder(res.Body)

	var machineMsg MachineMsg
	if err := dec.Decode(&machineMsg); err == io.EOF {
		return errMsg{err}
	} else if err != nil {
		return errMsg{err}
	}

	return machineMsg
}

func postMachine(newMachineName string, newMachineTypes string) Machine {
	var machine Machine

	// Create an HTTP client and make a GET request.
	c := &http.Client{Timeout: 10 * time.Second}

	// JSON body
	scriptTemplate := createTemplate("caddyfile", `{
		"Name":"{{.MACHINE_NAME}}",
		"Types":["{{.MACHINE_TYPES}}"]
	}`)

	var bodyBytes bytes.Buffer
	templateData := map[string]string{
		"MACHINE_NAME":  newMachineName,
		"MACHINE_TYPES": newMachineTypes,
	}

	if err := scriptTemplate.Execute(&bodyBytes, templateData); err != nil {
		fmt.Println("Cannot execute template for Caddyfile:", err)
	}

	res, err := c.Post(baseUrl+"machine", "application/json", bytes.NewBuffer(bodyBytes.Bytes()))

	if err != nil {
		// There was an error making our request. Wrap the error we received
		// in a message and return it.
		return machine
	}

	defer res.Body.Close()

	// We received a response from the server. Return the HTTP status code
	// as a message.
	dec := json.NewDecoder(res.Body)

	if err := dec.Decode(&machine); err == io.EOF {
		return machine
	} else if err != nil {
		return machine
	}

	return machine
}

type MachineMsg []Machine
type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }

/*Services*/
type Service struct {
	Id        string
	Name      string
	GitURL    string
	ProjectId string
}
type ServicesMsg []Service

func getServices() tea.Msg {

	// Create an HTTP client and make a GET request.
	c := &http.Client{Timeout: 10 * time.Second}
	res, err := c.Get(baseUrl + "service")

	if err != nil {
		// There was an error making our request. Wrap the error we received
		// in a message and return it.
		return errMsg{err}
	}

	defer res.Body.Close()

	// We received a response from the server. Return the HTTP status code
	// as a message.
	dec := json.NewDecoder(res.Body)

	var servicesMsg ServicesMsg
	if err := dec.Decode(&servicesMsg); err == io.EOF {
		return errMsg{err}
	} else if err != nil {
		return errMsg{err}
	}

	return servicesMsg
}

func postService(newServiceName string, newServiceGitURL string) Service {
	var service Service

	// Create an HTTP client and make a GET request.
	c := &http.Client{Timeout: 10 * time.Second}

	// JSON body
	scriptTemplate := createTemplate("caddyfile", `{
		"Name":"{{.SERVICE_NAME}}",
		"GitURL":"{{.GIT_URL}}"
	}`)

	var bodyBytes bytes.Buffer
	templateData := map[string]string{
		"SERVICE_NAME": newServiceName,
		"GIT_URL":      newServiceGitURL,
	}

	if err := scriptTemplate.Execute(&bodyBytes, templateData); err != nil {
		fmt.Println("Cannot execute template for Caddyfile:", err)
	}

	res, err := c.Post(baseUrl+"service", "application/json", bytes.NewBuffer(bodyBytes.Bytes()))

	if err != nil {
		// There was an error making our request. Wrap the error we received
		// in a message and return it.
		return service
	}

	defer res.Body.Close()

	// We received a response from the server. Return the HTTP status code
	// as a message.
	dec := json.NewDecoder(res.Body)

	if err := dec.Decode(&service); err == io.EOF {
		return service
	} else if err != nil {
		return service
	}

	return service
}

type Environment struct {
	Id         string
	Name       string
	Branch     string
	GitTag     string
	Domains    []string
	MachineIds []string
	Port       string
	ServiceId  string
}

type EnvironmentsMsg []Environment

func getEnvironments(serviceId string) tea.Cmd {

	return func() tea.Msg {
		// Create an HTTP client and make a GET request.
		c := &http.Client{Timeout: 10 * time.Second}
		res, err := c.Get(baseUrl + "service/" + serviceId + "/environment")

		if err != nil {
			// There was an error making our request. Wrap the error we received
			// in a message and return it.
			return errMsg{err}
		}

		defer res.Body.Close()

		// We received a response from the server. Return the HTTP status code
		// as a message.
		dec := json.NewDecoder(res.Body)

		var environmentsMsg EnvironmentsMsg
		if err := dec.Decode(&environmentsMsg); err == io.EOF {
			return errMsg{err}
		} else if err != nil {
			return errMsg{err}
		}
		return environmentsMsg
	}

}
