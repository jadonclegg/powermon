package power

import (
	"os/exec"
)

// Shutdown shuts down the computer...
func Shutdown(sudo bool) error {
	var cmd *exec.Cmd
	if sudo {
		cmd = exec.Command("sudo", "shutdown", "now")
	} else {
		cmd = exec.Command("shutdown", "now")
	}

	return cmd.Run()
}
