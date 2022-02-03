//go:build windows
// +build windows

package term

import (
	"os"
	"os/exec"
)

// Clear terminal output
func Clear() error {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	return cmd.Run()
}
