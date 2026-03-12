/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package common

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/michielvha/edgectl/pkg/logger"
)

// CheckRoot checks if the current process is running as root.
// If not, it re-executes the current command under sudo, preserving environment variables.
func CheckRoot() error {
	if os.Geteuid() == 0 {
		return nil
	}

	logger.Debug("not running as root, re-executing with sudo")
	fmt.Println("🔒 Root privileges required, escalating with sudo...")

	sudoPath, err := exec.LookPath("sudo")
	if err != nil {
		return fmt.Errorf("sudo not found: %w", err)
	}

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}

	// Build args: sudo -E <executable> <original args...>
	// -E preserves environment variables (VAULT_ADDR, BAO_TOKEN, etc.)
	args := append([]string{"sudo", "-E", execPath}, os.Args[1:]...)

	return syscall.Exec(sudoPath, args, os.Environ())
}
