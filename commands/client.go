package commands

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"time"
)

// ClientCommand holds command line options for the powermon client command
type ClientCommand struct {
	IP       string `short:"a" long:"host" required:"true" description:"Address of server. Can be an IP or a hostname"`
	Port     int    `short:"p" long:"port" default:"10101" description:"Port to listen on, or connect to. Default is 10101"`
	Timeout  int    `short:"t" long:"timeout" default:"60" description:"Shuts down the computer or runs custom timeout script X seconds after failing to ping the server. Default is 60"`
	Interval int    `short:"i" long:"interval" default:"60" description:"Check if server is up every X seconds. Default is 60."`

	stopped bool
}

// ClientInfo contains basic information about the client, sent with every ping
type ClientInfo struct {
	MACS     []string
	NickName string
}

var clientData ClientInfo

// Execute runs the command
func (command *ClientCommand) Execute(args []string) error {
	err := Powermon.Init()
	if err != nil {
		return err
	}

	err = command.validateOptions()
	if err != nil {
		return err
	}

	command.loadClientInfo()

	// Make a channel to receive ping updates, and start the pinger
	pingChan := make(chan bool)
	go command.pinger(pingChan)

	// Make a timer with the timeout as the duration
	timeout := time.Duration(command.Timeout) * time.Second
	timer := time.NewTimer(timeout)

	timerRunning := true
	for {
		select {
		case <-timer.C:
			// If we hit the timeout, run the onTimeout command
			command.onTimeout()

			// Quit the program
			return nil
		case success := <-pingChan:
			// If the ping was successful, stop the timer if it's running
			if success {
				if timerRunning {
					if !timer.Stop() {
						<-timer.C
					}
					timerRunning = false
					Logger.Info("Ping succeeded, stopping timeout")
				}
			} else {
				// if the ping failed, start the timeout timer
				if !timerRunning {
					timer.Reset(timeout)
					timerRunning = true
				}
			}
		}
	}
}

func (command *ClientCommand) validateOptions() error {
	// Parse the IP, make sure it's valid.
	ip := net.ParseIP(command.IP)
	if ip == nil {
		// Check if it's a valid DNS address...
		ips, err := net.LookupIP(command.IP)
		if err != nil || len(ips) == 0 {
			return errors.New("error: specified host doesn't exist")
		}
	} else {
		// Store the valid IP back in the command struct.
		command.IP = ip.String()
	}

	// Make sure the port is valid
	if command.Port < 1 || command.Port > 65535 {
		return errors.New("error: invaid port")
	}

	return nil
}

func (command *ClientCommand) onTimeout() {
	command.stopped = true
	Logger.Warn("Timeout reached, shutting down.")

	SendNotification("Timeout reached, shutting down.")

	cmd := exec.Command("shutdown", "now")
	err := cmd.Run()
	if err != nil {
		Logger.Error("Error executing command: ", err)
	}
}

func (command *ClientCommand) pinger(pchan chan bool) {
	// Build the url using the IP and port
	url := fmt.Sprintf("http://%s:%d/status", command.IP, command.Port)
	httpClient := http.Client{
		Timeout: 2 * time.Second,
	}

	for {
		// quit if the process was stopped
		if command.stopped {
			return
		}

		// Run the the ping, catch errors
		err := ping(url, httpClient)
		if err != nil {
			// If there's an error, print it to the console / log, and sleep for 3 seconds. Then re-run the loop.
			// Basically, if there's an error, we shorten the timeout to 3 seconds and keep trying to get a succesful ping until we do,
			// or the timeout kicks in.
			pchan <- false
			Logger.Warn("Error pinging server: ", err)
			time.Sleep(3 * time.Second)
			continue
		}

		// If there was no error, send to the pinger channel that we 'succeeded' in getting a successful ping
		pchan <- true
		// Sleep for the specified interval
		time.Sleep(time.Duration(command.Interval) * time.Second)
	}
}

func ping(url string, httpClient http.Client) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	err := enc.Encode(clientData)
	if err != nil {
		return err
	}

	// Try to "ping" the url. If there's an error, or the status code isn't 200, return an error.
	resp, err := httpClient.Post(url, "application/json", buf)
	if err != nil {
		return err
	}

	if resp != nil && resp.StatusCode != http.StatusOK {
		return errors.New("error: non 200 status code received")
	}

	return nil
}

func (command *ClientCommand) loadClientInfo() {
	clientData.NickName = Powermon.NickName

	interfaces, err := net.Interfaces()
	if err != nil {
		Logger.Error("Failed to get mac addresses for interfaces: ", err)
		return
	}

	for _, i := range interfaces {
		if i.HardwareAddr != nil {
			clientData.MACS = append(clientData.MACS, i.HardwareAddr.String())
		}
	}
}
