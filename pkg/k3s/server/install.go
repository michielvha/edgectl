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

// Install sets up the K3s server on the host.
// If `isExisting` is true, it pulls the token from the secret store using the supplied clusterID.
// Otherwise, it generates a new clusterID and saves token + kubeconfig to the secret store.
// If `vip` is provided, it will be used in the TLS SANs for the server.
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
			_, storedVIP, _, err := store.RetrieveMasterInfo("k3s", clusterID)
			if err == nil && storedVIP != "" {
				fmt.Printf("🔍 VIP fetched from secret store: %s\n", storedVIP)
				vip = storedVIP
			}
		}
	} else {
		// Generate a new cluster ID
		clusterID = fmt.Sprintf("k3s-%s", uuid.New().String()[:8])
		_ = os.MkdirAll("/etc/edgectl", 0o750)
		_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o600)
		fmt.Printf("🆔 Generated cluster ID: %s\n", clusterID)
	}

	// If a VIP was provided, use that in the TLS SANs
	installOptions := ""
	if vip != "" {
		installOptions = fmt.Sprintf("-l %s", vip)
		fmt.Printf("🌐 Using VIP %s for load balancer TLS SANs\n", vip)
	}

	// Run the installation script with options
	common.RunBashFunction("k3s.sh", fmt.Sprintf("install_k3s_server %s", installOptions))

	// If this is a new cluster, store token and kubeconfig in the secret store
	if !isExisting {
		tokenBytes, err := os.ReadFile("/var/lib/rancher/k3s/server/node-token")
		if err != nil {
			return fmt.Errorf("failed to read generated node token: %w", err)
		}

		token := strings.TrimSpace(string(tokenBytes))
		if err := store.StoreJoinToken("k3s", clusterID, token); err != nil {
			return fmt.Errorf("failed to store token in secret store: %w", err)
		}
		fmt.Printf("🔐 Token successfully stored in secret store for cluster %s\n", clusterID)

		kubeconfigPath := "/etc/rancher/k3s/k3s.yaml"
		if _, statErr := os.Stat(kubeconfigPath); os.IsNotExist(statErr) {
			return fmt.Errorf("kubeconfig file not found at path: %s", kubeconfigPath)
		}

		err = store.StoreKubeConfig("k3s", clusterID, kubeconfigPath, vip)
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

	existingHosts, storedVIP, _, err := store.RetrieveMasterInfo("k3s", clusterID)
	if err == nil {
		hosts = existingHosts
		if storedVIP != "" {
			existingVIP = storedVIP
		}
		logger.Debug("%s", fmt.Sprintf("Found existing master nodes: %v", hosts))
	} else {
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
	err = store.StoreMasterInfo("k3s", clusterID, hostname, hosts, existingVIP)
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
	token, err := store.RetrieveJoinToken("k3s", clusterID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve join token: %w", err)
	}

	// ensure edgectl main directory exists
	_ = os.MkdirAll("/etc/edgectl", 0o750)

	if err := os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o600); err != nil {
		return "", fmt.Errorf("failed to write cluster-id: %w", err)
	}

	// Set token as environment variable for the bash script to use
	_ = os.Setenv("K3S_TOKEN", token)
	fmt.Println("✅ Set K3S_TOKEN environment variable")

	// For additional master nodes, get the first master's IP
	firstMasterIP, ipErr := store.RetrieveFirstMasterIP("k3s", clusterID)
	if ipErr == nil && firstMasterIP != "" {
		_ = os.Setenv("K3S_URL", fmt.Sprintf("https://%s:6443", firstMasterIP))
		fmt.Printf("✅ Set K3S_URL environment variable: https://%s:6443\n", firstMasterIP)
	} else if ipErr != nil {
		logger.Debug("Could not find first master IP: %v", ipErr)
	}

	return token, nil
}
