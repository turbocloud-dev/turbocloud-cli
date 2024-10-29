package main

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const url = "http://localhost:5445/machine"

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

func getMachines() tea.Msg {

	// Create an HTTP client and make a GET request.
	c := &http.Client{Timeout: 10 * time.Second}
	res, err := c.Get(url)

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

type MachineMsg []Machine

type errMsg struct{ err error }

// For messages that contain errors it's often handy to also implement the
// error interface on the message.
func (e errMsg) Error() string { return e.err.Error() }
