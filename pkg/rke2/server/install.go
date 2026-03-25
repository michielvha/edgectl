/*
Copyright © 2025 VH & Co - contact@vhco.pro
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

// clusterIDDir is the directory where the cluster-id file is written.
// Tests can override this to use a temporary directory.
var clusterIDDir = "/etc/edgectl"

// Install sets up the RKE2 server on the host.
// If `isExisting` is true, it pulls the token from the secret store using the supplied clusterID.
// Otherwise, it generates a new clusterID and saves token + kubeconfig to the secret store.
// If `vip` is provided, it will be used in the TLS SANs for the server. if a cluster id is provided, it will fetch VIP from the secret store.
func Install(store vault.SecretStore, clusterID string, isExisting bool, vip string) error {
	// Get current hostname
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// If the cluster ID was provided (existing cluster), fetch the join token
	if isExisting {
		if _, err := FetchTokenFromSecretStore(store, clusterID); err != nil {
			return err
		}

		// For existing clusters, try to fetch the VIP from the secret store if none was provided
		if vip == "" {
			_, storedVIP, _, err := store.RetrieveMasterInfo("rke2", clusterID)
			if err == nil && storedVIP != "" {
				fmt.Printf("🔍 VIP fetched from secret store: %s\n", storedVIP)
				vip = storedVIP
			}
		}
	} else {
		// Generate a new cluster ID
		clusterID = fmt.Sprintf("rke2-%s", uuid.New().String()[:8])
		_ = os.MkdirAll(clusterIDDir, 0o750)
		_ = os.WriteFile(clusterIDDir+"/cluster-id", []byte(clusterID), 0o600)
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

	// If this is a new cluster, store token and kubeconfig in the secret store
	if !isExisting {
		tokenBytes, err := os.ReadFile("/var/lib/rancher/rke2/server/node-token")
		if err != nil {
			return fmt.Errorf("failed to read generated node token: %w", err)
		}

		token := strings.TrimSpace(string(tokenBytes))
		if err := store.StoreJoinToken("rke2", clusterID, token); err != nil {
			return fmt.Errorf("failed to store token in secret store: %w", err)
		}
		fmt.Printf("🔐 Token successfully stored in secret store for cluster %s\n", clusterID)

		kubeconfigPath := "/etc/rancher/rke2/rke2.yaml"
		if _, statErr := os.Stat(kubeconfigPath); os.IsNotExist(statErr) {
			return fmt.Errorf("kubeconfig file not found at path: %s", kubeconfigPath)
		}

		err = store.StoreKubeConfig("rke2", clusterID, kubeconfigPath, vip)
		if err != nil {
			return fmt.Errorf("failed to store kubeconfig in secret store: %w", err)
		}
		fmt.Printf("🔐 Kubeconfig successfully stored in secret store for cluster %s\n", clusterID)
	}

	// Track master nodes in the secret store (for both new and existing clusters)
	logger.Debug("Updating master node information in secret store")

	// Try to get existing master nodes if any
	var hosts []string
	existingVIP := vip // Use provided VIP as default

	existingHosts, storedVIP, _, err := store.RetrieveMasterInfo("rke2", clusterID)
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
	err = store.StoreMasterInfo("rke2", clusterID, hostname, hosts, existingVIP)
	if err != nil {
		return fmt.Errorf("failed to store master node info in secret store: %w", err)
	}

	fmt.Printf("🔄 Master nodes updated in secret store: %d node(s) registered\n", len(hosts))
	if existingVIP != "" {
		fmt.Printf("ℹ️ Load balancer VIP stored in secret store: %s\n", existingVIP)
	}

	return nil
}

// FetchTokenFromSecretStore fetches token from the secret store & sets as env var.
// Also retrieves the first master's IP if joining an existing cluster.
func FetchTokenFromSecretStore(store vault.SecretStore, clusterID string) (string, error) {
	token, err := store.RetrieveJoinToken("rke2", clusterID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve join token: %w", err)
	}

	// ensure edgectl main directory exists
	_ = os.MkdirAll(clusterIDDir, 0o750)

	if err := os.WriteFile(clusterIDDir+"/cluster-id", []byte(clusterID), 0o600); err != nil {
		return "", fmt.Errorf("failed to write cluster-id: %w", err)
	}

	// Set token as environment variable for the bash script to use
	_ = os.Setenv("RKE2_TOKEN", token)
	fmt.Println("✅ Set RKE2_TOKEN environment variable")

	// For additional master nodes, get the first master's IP
	firstMasterIP, ipErr := store.RetrieveFirstMasterIP("rke2", clusterID)
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
