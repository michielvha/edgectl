/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package common

import (
	"errors"
	"fmt"
	"os"
)

// CheckRoot checks if the current process is running as root.
// It prints an error message and returns an error if not running as root.
func CheckRoot() error {
	if os.Geteuid() != 0 {
		err := errors.New("this command must be run as root, try using `sudo`")
		fmt.Printf("❌ %v\n", err)
		return err
	}
	return nil
}
