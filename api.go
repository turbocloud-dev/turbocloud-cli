package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
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
	CPUUsage       string
	MEMUsage       string
	DiskUsage      string
}

type MachineStats struct {
	Id              string
	MachineId       string
	CPUUsage        int64
	AvailableMemory int64
	TotalMemory     int64
	AvailableDisk   int64
	TotalDisk       int64
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

	machinesStats := getMachineStates()

	for _, machineStats := range machinesStats {
		for index := range machineMsg {
			if machineStats.MachineId == machineMsg[index].Id {
				machineMsg[index].CPUUsage = fmt.Sprintf("%d", machineStats.CPUUsage)
				machineMsg[index].MEMUsage = fmt.Sprintf("%d", machineStats.AvailableMemory)
				machineMsg[index].DiskUsage = fmt.Sprintf("%d", machineStats.AvailableDisk/(1024*1024))
			}
		}
	}

	return machineMsg
}

func getMachineStates() []MachineStats {

	var machineStats []MachineStats
	// Create an HTTP client and make a GET request.
	c := &http.Client{Timeout: 10 * time.Second}
	res, err := c.Get(baseUrl + "machine/stats")

	if err != nil {
		// There was an error making our request. Wrap the error we received
		// in a message and return it.
		return machineStats
	}

	defer res.Body.Close()

	// We received a response from the server. Return the HTTP status code
	// as a message.
	dec := json.NewDecoder(res.Body)

	if err := dec.Decode(&machineStats); err == io.EOF {
		return machineStats
	} else if err != nil {
		return machineStats
	}

	return machineStats
}

func getMachineOptions() ([]huh.Option[string], []Machine) {
	machines := getMachines().(MachineMsg) //type asssertion
	opts := []string{}
	for _, machine := range machines {
		opts = append(opts, machine.Name)
	}
	return huh.NewOptions(opts...), machines

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

func deleteMachine(machineId string) bool {

	req, err := http.NewRequest(http.MethodDelete, baseUrl+"machine/"+machineId, nil)
	if err != nil {
		return false
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	return true

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
	Id                   string
	Name                 string
	Branch               string
	GitTag               string
	Domains              []string
	MachineIds           []string
	Port                 string
	ServiceId            string
	LastDeploymentStatus string
}

type EnvironmentsMsg []Environment

func getEnvironmentsCmd(serviceId string) tea.Cmd {

	return func() tea.Msg {
		return getEnvironments(serviceId)
	}

}

func getEnvironments(serviceId string) tea.Msg {
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

type NewEnvironmentAddedMsg Environment

func postEnvironment(newEnvironment Environment) tea.Cmd {

	return func() tea.Msg {
		var environment NewEnvironmentAddedMsg

		// Create an HTTP client and make a GET request.
		c := &http.Client{Timeout: 10 * time.Second}

		// JSON body
		scriptTemplate := createTemplate("environment", `{
			"ServiceId":"{{.SERVICE_ID}}",
			"Name":"{{.ENV_NAME}}",
			"Branch":"{{.ENV_BRANCH}}",
			"GitTag":"{{.ENV_TAG}}",
			"Domains":["{{.ENV_DOMAIN}}"],
			"Port":"{{.ENV_PORT}}",
			"MachineIds":["{{.MACHINE_IDS}}"]
	}`)

		var bodyBytes bytes.Buffer
		templateData := map[string]string{
			"SERVICE_ID":  newEnvironment.ServiceId,
			"ENV_NAME":    newEnvironment.Name,
			"ENV_BRANCH":  newEnvironment.Branch,
			"ENV_TAG":     newEnvironment.GitTag,
			"ENV_DOMAIN":  newEnvironment.Domains[0], //currently we can add only one domain during environment creation
			"ENV_PORT":    newEnvironment.Port,
			"MACHINE_IDS": strings.Join(newEnvironment.MachineIds, `","`),
		}

		if err := scriptTemplate.Execute(&bodyBytes, templateData); err != nil {
			fmt.Println("Cannot execute template for Caddyfile:", err)
		}

		res, err := c.Post(baseUrl+"environment", "application/json", bytes.NewBuffer(bodyBytes.Bytes()))

		if err != nil {
			// There was an error making our request. Wrap the error we received
			// in a message and return it.
			return environment
		}

		defer res.Body.Close()

		// We received a response from the server. Return the HTTP status code
		// as a message.
		dec := json.NewDecoder(res.Body)

		if err := dec.Decode(&environment); err == io.EOF {
			return environment
		} else if err != nil {
			return environment
		}

		return environment
	}

}

type EnvironmentEditedMsg Environment

func updateEnvironment(newEnvironment Environment) EnvironmentEditedMsg {

	var environment EnvironmentEditedMsg

	// JSON body
	scriptTemplate := createTemplate("environment", `{
			"Id":"{{.ENV_ID}}",
			"Name":"{{.ENV_NAME}}",
			"Branch":"{{.ENV_BRANCH}}",
			"GitTag":"{{.ENV_TAG}}",
			"Domains":["{{.ENV_DOMAIN}}"],
			"Port":"{{.ENV_PORT}}",
			"MachineIds":["{{.MACHINE_IDS}}"]
	}`)

	var bodyBytes bytes.Buffer
	templateData := map[string]string{
		"ENV_ID":      newEnvironment.Id,
		"ENV_NAME":    newEnvironment.Name,
		"ENV_BRANCH":  newEnvironment.Branch,
		"ENV_TAG":     newEnvironment.GitTag,
		"ENV_DOMAIN":  newEnvironment.Domains[0], //currently we can add only one domain during environment creation
		"ENV_PORT":    newEnvironment.Port,
		"MACHINE_IDS": strings.Join(newEnvironment.MachineIds, `","`),
	}

	if err := scriptTemplate.Execute(&bodyBytes, templateData); err != nil {
		fmt.Println("Cannot execute template for Caddyfile:", err)
	}

	req, err := http.NewRequest(http.MethodPut, baseUrl+"environment", bytes.NewBuffer(bodyBytes.Bytes()))
	if err != nil {
		// handle error
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return environment
	}
	defer res.Body.Close()

	// We received a response from the server. Return the HTTP status code
	// as a message.
	dec := json.NewDecoder(res.Body)

	if err := dec.Decode(&environment); err == io.EOF {
		return environment
	} else if err != nil {
		return environment
	}

	return environment

}

func deleteEnvironment(environmentId string) bool {

	req, err := http.NewRequest(http.MethodDelete, baseUrl+"environment/"+environmentId, nil)
	if err != nil {
		return false
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	return true

}

func deployEnvironment(environmentId string) bool {

	req, err := http.NewRequest(http.MethodGet, baseUrl+"deploy/environment/"+environmentId, nil)
	if err != nil {
		return false
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false
	}
	defer res.Body.Close()

	return true

}
