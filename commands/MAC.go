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
