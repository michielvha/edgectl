/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package agent

import (
	"fmt"
	"os"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// Install sets up the RKE2 agent on the host.
// It fetches the join token from Vault using the supplied clusterID.
func Install(clusterID string) error {
	vaultClient, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	if _, err := FetchToken(clusterID); err != nil {
		return err
	}

	// fetch the VIP from Master Info in Vault
	_, storedVIP, _, err := vaultClient.RetrieveMasterInfo(clusterID)
	if err == nil && storedVIP != "" {
		fmt.Printf("🔍 VIP fetched from Vault: %s\n", storedVIP)
	}

	installOptions := ""
	if storedVIP != "" {
		installOptions = fmt.Sprintf("-l %s", storedVIP)
		fmt.Printf("🌐 Using VIP %s for load balancer TLS SANs\n", storedVIP)
	} else {
		// TODO: add fallback ?
		logger.Debug("No VIP found in Vault, using default settings")
	}
	// Run the installation script with options
	common.RunBashFunction("rke2.sh", fmt.Sprintf("install_rke2_agent %s", installOptions))

	return nil
}

// Fetch token from Vault & set as env var
// Also retrieves the first master's IP if joining an existing cluster
func FetchToken(clusterID string) (string, error) {
	vaultClient, err := vault.NewClient()
	if err != nil {
		return "", fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	token, err := vaultClient.RetrieveJoinToken(clusterID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve join token: %w", err)
	}

	// ensure edgectl main directory exists
	_ = os.MkdirAll("/etc/edgectl", 0o755)

	if err := os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o644); err != nil {
		return "", fmt.Errorf("failed to write cluster-id: %w", err)
	}

	// Set token as environment variable for the bash script to use
	_ = os.Setenv("RKE2_TOKEN", token)
	fmt.Println("✅ Set RKE2_TOKEN environment variable")

	return token, nil
}