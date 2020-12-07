package power

import (
	"os/exec"
)

// Shutdown shuts down the computer, maybe... haven't tested the Windows version. In theory, it should work.
// HOWEVER: It will NOT force applications to close, and will possibly run updates. I hate Windows.
func Shutdown() error {
	cmd := exec.Command("shutdown", "/s", "/t", "0")
	return cmd.Run()
}
