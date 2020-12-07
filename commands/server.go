package commands

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
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

var verified chan string

// Execute is run by the command line args parser.
func (command *ServerCommand) Execute(args []string) error {
	err := Powermon.Init()
	if err != nil {
		return err
	}

	if command.Port < 1 || command.Port > 65535 {
		return errors.New("error: invalid port")
	}

	err = command.runWakeups()
	if err != nil {
		return err
	}

	r := mux.NewRouter()

	if command.Verify {
		verified = make(chan string)
		r.HandleFunc("/verify", verifyHandler).Methods("POST")
	}

	r.HandleFunc("/status", statusHandler).Methods("GET")

	http.Handle("/", r)

	Logger.Info("Listening on port ", command.Port)

	SendNotification(fmt.Sprintf("Server started, listening on port %d", command.Port))

	err = http.ListenAndServe(fmt.Sprintf(":%d", command.Port), nil)

	return err
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Status okay.\n")
	Logger.Info("Received ping from client at ip ", r.RemoteAddr)
}

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
	if command.Wake != nil && len(command.Wake) > 0 {
		for _, macStr := range command.Wake {
			mac, err := net.ParseMAC(macStr)
			if err != nil {
				return err
			}

			macs[mac.String()] = false
		}
	}

	if command.WakeFile != "" && len(command.WakeFile) > 0 {
		file, err := os.Open(command.WakeFile)
		if err != nil {
			return err
		}

		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineCount := 0
		for scanner.Scan() {
			lineCount++
			line := scanner.Text()
			if line != "" && len(line) > 0 && line[0] != '#' {
				mac, err := net.ParseMAC(line)
				if err != nil {
					return errors.New("invalid mac address on line " + string(line) + " in wakelist file")
				}

				macs[mac.String()] = false
			}
		}
	}

	go sendWolpackets(macs, command.Verify)
	return nil
}

func sendWolpackets(macs map[string]bool, verify bool) {
	verifiedCount := 0

	t := time.NewTicker(time.Second * 15)
	sentCount := 0

	for {
		select {
		case mac := <-verified:
			_, ok := macs[mac]
			if ok {
				macs[mac] = true
				verifiedCount++

				if verifiedCount == len(macs) {
					t.Stop()
					Logger.Info("Received verification from all clients. Stopped sending WOL packets.")
					SendNotification("All clients are back online.")
				}
			}
		case <-t.C:
			sentCount++
			for mac, online := range macs {
				if online {
					continue
				}

				wake(mac)
			}

			if !verify && sentCount > 10 {
				Logger.Info("Sent 10 WOL packets to each client, stopping WOL sender.")
				return
			}
		}
	}
}

func verifyHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	macs := r.Form["mac"]
	SendNotification(fmt.Sprintf("Client at %s with mac addresses %v is back online.", r.RemoteAddr, macs))
	for _, mac := range macs {
		verified <- mac
		Logger.Info(fmt.Sprintf("Received verification for mac %s from %s", mac, r.RemoteAddr))
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Received.\n")
}
