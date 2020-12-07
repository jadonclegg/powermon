package power

import (
	"os/exec"
)

// Shutdown shuts down the computer...
func Shutdown() error {
	cmd := exec.Command("shutdown", "now")
	return cmd.Run()
}
