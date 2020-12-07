package commands

import (
	"fmt"
	"net"
)

// MACCommand is the struct for the options for the MAC command.
// No options for now, not needed.
type MACCommand struct {
}

// Execute gets run when the mac command is passed on the command line.
// Basically this just lists all the mac addresses and network interfaces on the computer, so you can use them to send WOL packets
// There are many ways to do this already built into the OS, but this is just convenient
func (command *MACCommand) Execute(args []string) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	for _, i := range interfaces {
		if i.HardwareAddr != nil {
			fmt.Printf("Interface: %s\nMAC: %s\n\n", i.Name, i.HardwareAddr)
		}
	}

	return nil
}
