/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package common

import (
	"errors"
	"fmt"
	"os"

	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// CheckRoot checks if the current process is running as root.
// It prints an error message and returns an error if not running as root.
func CheckRoot() error {
	if os.Geteuid() != 0 {
		logger.Debug("verifying if user is root, program will exit if not")
		err := errors.New("this command must be run as root, try using `sudo`")
		fmt.Printf("❌ %v\n", err)
		return err
	}
	return nil
}

// InitVaultClient centralizes Vault client creation and error handling
// Returns nil if the client initialization failed
// TODO: implement this everywhere we create vaultclient, example call in cmd/vault.go on line 48
func InitVaultClient() *vault.Client {
	logger.Debug("initializing Vault client")
	vaultClient, err := vault.NewClient()
	if err != nil {
		fmt.Printf("❌ failed to initialize Vault client: %v\n", err)
		return nil
	}
	return vaultClient
}
