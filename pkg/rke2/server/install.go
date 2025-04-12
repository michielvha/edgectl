/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// Install sets up the RKE2 server on the host.
// If `isExisting` is true, it pulls the token from Vault using the supplied clusterID.
// Otherwise, it generates a new clusterID and saves token + kubeconfig to Vault.
// If `vip` is provided, it will be used in the TLS SANs for the server. if a cluster id is provided, it will fetch VIP from the vault.
func Install(clusterID string, isExisting bool, vip string) error {
	vaultClient, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	// Get current hostname
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// If the cluster ID was provided (existing cluster), fetch the join token
	if isExisting {
		if _, err := FetchTokenFromVault(clusterID); err != nil {
			return err
		}

		// For existing clusters, try to fetch the VIP from Vault if none was provided
		if vip == "" {
			_, storedVIP, err := vaultClient.RetrieveMasterInfo(clusterID)
			if err == nil && storedVIP != "" {
				fmt.Printf("🔍 VIP fetched from Vault: %s\n", storedVIP)
				vip = storedVIP
			}
		}
	} else {
		// Generate a new cluster ID
		clusterID = fmt.Sprintf("rke2-%s", uuid.New().String()[:8])
		_ = os.MkdirAll("/etc/edgectl", 0o755)
		_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o644)
		fmt.Printf("🆔 Generated cluster ID: %s\n", clusterID)
	}

	// If a VIP was provided, use that in the TLS SANs
	installOptions := ""
	if vip != "" {
		installOptions = fmt.Sprintf("-l %s", vip)
		fmt.Printf("🌐 Using VIP %s for load balancer TLS SANs\n", vip)
	}

	// Run the installation script with options
	common.RunBashFunction("rke2.sh", fmt.Sprintf("install_rke2_server %s", installOptions))

	// If this is a new cluster, store token and kubeconfig in Vault
	if !isExisting {
		tokenBytes, err := os.ReadFile("/var/lib/rancher/rke2/server/node-token")
		if err != nil {
			return fmt.Errorf("failed to read generated node token: %w", err)
		}

		token := strings.TrimSpace(string(tokenBytes))
		if err := vaultClient.StoreJoinToken(clusterID, token); err != nil {
			return fmt.Errorf("failed to store token in Vault: %w", err)
		}
		fmt.Printf("🔐 Token successfully stored in Vault for cluster %s\n", clusterID)

		kubeconfigPath := "/etc/rancher/rke2/rke2.yaml"
		if _, statErr := os.Stat(kubeconfigPath); os.IsNotExist(statErr) {
			return fmt.Errorf("kubeconfig file not found at path: %s", kubeconfigPath)
		}

		err = vaultClient.StoreKubeConfig(clusterID, kubeconfigPath, vip)
		if err != nil {
			return fmt.Errorf("failed to store kubeconfig in Vault: %w", err)
		}
		fmt.Printf("🔐 Kubeconfig successfully stored in Vault for cluster %s\n", clusterID)
	}

	// Track master nodes in Vault (for both new and existing clusters)
	logger.Debug("Updating master node information in Vault")

	// Try to get existing master nodes if any
	var hosts []string
	existingVIP := vip // Use provided VIP as default

	existingHosts, storedVIP, err := vaultClient.RetrieveMasterInfo(clusterID)
	if err == nil {
		// Successfully retrieved existing master info
		hosts = existingHosts
		if storedVIP != "" {
			existingVIP = storedVIP // Use the stored VIP if it exists
		}
		logger.Debug("%s", fmt.Sprintf("Found existing master nodes: %v", hosts))
	} else {
		// First master node in this cluster
		hosts = []string{}
		logger.Debug("No existing master nodes found, initializing new master list")
	}

	// Add this host to the list if not already present
	found := false
	for _, h := range hosts {
		if h == hostname {
			found = true
			break
		}
	}

	if !found {
		hosts = append(hosts, hostname)
		logger.Debug("Added this host (%s) to master nodes list", hostname)
	} else {
		logger.Debug("This host (%s) is already in master nodes list", hostname)
	}

	// Store updated master info with the VIP
	err = vaultClient.StoreMasterInfo(clusterID, hostname, hosts, existingVIP)
	if err != nil {
		return fmt.Errorf("failed to store master node info in Vault: %w", err)
	}

	fmt.Printf("🔄 Master nodes updated in Vault: %d node(s) registered\n", len(hosts))
	if existingVIP != "" {
		fmt.Printf("ℹ️ Load balancer VIP stored in Vault: %s\n", existingVIP)
	}

	return nil
}

// Fetch token from Vault & set as env var
// Also retrieves the first master's IP if joining an existing cluster
func FetchTokenFromVault(clusterID string) (string, error) {
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

	// For additional master nodes, get the first master's IP
	firstMasterIP, ipErr := vaultClient.RetrieveFirstMasterIP(clusterID)
	if ipErr == nil && firstMasterIP != "" {
		// Set server IP as environment variable if available
		_ = os.Setenv("RKE2_SERVER_IP", firstMasterIP)
		fmt.Printf("✅ Set RKE2_SERVER_IP environment variable: %s\n", firstMasterIP)
	} else if ipErr != nil {
		// Log the error but continue since it's not critical (could be first server)
		logger.Debug("Could not find first master IP: %v", ipErr)
	}

	return token, nil
}
