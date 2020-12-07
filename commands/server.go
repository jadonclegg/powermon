package commands

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/linde12/gowol"
)

// ServerCommand holds the command line argument definitions for the program.
type ServerCommand struct {
	Port     int      `short:"p" long:"port" default:"10101" description:"Port to listen on, or connect to. Default is 10101"`
	Wake     []string `short:"w" long:"wake" description:"MAC addresses to send WOL packet to when the server is started."`
	WakeFile string   `long:"wakelist" description:"File with a list of MAC addresses separated by newlines to wake when the server is started."`
	Verify   bool     `long:"verify" description:"Send WOL packets until the hosts verify that they are back online. Use with the -w [--wake] or --wakelist options."`
}

var verifyChan chan ClientInfo
var doneVerifying bool

// Execute is run by the command line args parser.
func (command *ServerCommand) Execute(args []string) error {
	// Initialize the app with the global options first
	err := Powermon.Init()
	if err != nil {
		return err
	}

	// Make sure the port is valid
	if command.Port < 1 || command.Port > 65535 {
		return errors.New("error: invalid port")
	}

	doneVerifying = !command.Verify
	verifyChan = make(chan ClientInfo)

	// Run the code to send WOL packets
	err = command.runWakeups()
	if err != nil {
		return err
	}

	r := mux.NewRouter()

	// Handles the 'pings' from the clients
	r.HandleFunc("/status", statusHandler).Methods("POST")

	http.Handle("/", r)

	Logger.Info("Listening on port ", command.Port)

	// Send a notification that the server is started. Useful for if the power goes out, when it comes back online
	SendNotification(fmt.Sprintf("Server started, listening on port %d", command.Port))

	err = http.ListenAndServe(fmt.Sprintf(":%d", command.Port), nil)
	if err != nil {
		SendNotification(fmt.Sprintf("Error starting powermon server: %s", err))
	}

	return err
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	dec := json.NewDecoder(r.Body)
	clientData := ClientInfo{}
	err := dec.Decode(&clientData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Should always send a 200 status
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Status okay.\n")
	Logger.Info(fmt.Sprintf("Received ping from [%s] at ip %s", clientData.NickName, r.RemoteAddr))

	if !doneVerifying {
		verifyChan <- clientData
	}
}

// Send a WOL packet to the specified mac address
func wake(mac string) {
	packet, err := gowol.NewMagicPacket(mac)
	if err != nil {
		Logger.Warn(fmt.Sprintf("Failed to send WOL to %s -- %s", mac, err))
	} else {
		packet.Send("255.255.255.255")
		Logger.Info("Sent WOL packet to ", mac)
	}
}

func (command *ServerCommand) runWakeups() error {
	macs := make(map[string]bool)

	// Validate the mac addresses passed in using the -w [--wake] option. Can be more than one.
	if command.Wake != nil && len(command.Wake) > 0 {
		for _, macStr := range command.Wake {
			mac, err := net.ParseMAC(macStr)
			if err != nil {
				return err
			}

			// Use mac.String() instead of macStr to make sure it's in a valid format for the wake() function
			// boolean value says whether or not the verification has been received for that mac address
			macs[mac.String()] = false
		}
	}

	// If the wakefile is specified [--wakelist] open the file, and read the mac addresses from it. Validate them.
	if command.WakeFile != "" && len(command.WakeFile) > 0 {
		file, err := os.Open(command.WakeFile)
		if err != nil {
			return err
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)

		// Keep track of the line so we can report errors nicely
		lineCount := 0

		// scan the file line by line
		for scanner.Scan() {
			lineCount++
			line := strings.Trim(scanner.Text(), " \t")

			// Ignore empty lines and commented lines (using #)
			if line != "" && len(line) > 0 && line[0] != '#' {
				mac, err := net.ParseMAC(line)
				if err != nil {
					return errors.New("invalid mac address on line " + string(line) + " in wakelist file")
				}

				macs[mac.String()] = false
			}
		}
	}

	// Start the goroutine to send WOL packets
	go sendWolpackets(macs, command.Verify)
	return nil
}

func sendWolpackets(macs map[string]bool, verify bool) {
	// Keep track of how many clients have been verified to be online
	verifiedCount := 0

	// Create a ticker to send WOL packets every 15 seconds
	t := time.NewTicker(time.Second * 15)

	// Keep track of how many times we've sent WOL packets
	sentCount := 0

	for {
		select {
		// Receive a mac from the /verify endpoint.
		case clientData := <-verifyChan:
			if verifiedCount < len(macs) {
				for _, mac := range clientData.MACS {
					online, ok := macs[mac]
					// If the mac address isn't in the dictionary, ignore it
					if ok && !online {
						macs[mac] = true
						verifiedCount++
						Logger.Info(fmt.Sprintf("Received verification from [%s] mac %s", clientData.NickName, mac))
						SendNotification(fmt.Sprintf("Client [%s] %s is verified back online.", clientData.NickName, mac))

						// If we've verified that ALL of them are back online, stop sending WOL packets
						if verifiedCount == len(macs) {
							t.Stop()
							Logger.Info("Received verification from all clients. Stopped sending WOL packets.")
							SendNotification("All clients are back online.")
							doneVerifying = true
						}
					}
				}
			}
		case <-t.C:
			sentCount++
			// Send WOL packet to each mac, but only if they're not verified
			for mac, online := range macs {
				if online {
					continue
				}

				wake(mac)
			}

			// Quit after sending 10 WOL packets (Spread across 150 seconds) if verification isn't enabled.
			if !verify && sentCount > 10 {
				Logger.Info("Sent 10 WOL packets to each client, stopping WOL sender.")
				t.Stop()
			}
		}
	}
}
