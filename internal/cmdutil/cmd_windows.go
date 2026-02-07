//go:build windows

package cmdutil

import (
	"os/exec"
	"syscall"
)

// HideWindow sets the command to run without showing a console window on Windows
func HideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: 0x08000000, // CREATE_NO_WINDOW
	}
}
