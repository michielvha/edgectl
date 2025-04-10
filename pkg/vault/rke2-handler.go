/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package vault

import (
	"fmt"
	"os"
	"path/filepath"
)

// StoreJoinToken saves a token under a specific cluster path
func (c *Client) StoreJoinToken(clusterID, token string) error {
	return c.StoreSecret(fmt.Sprintf("kv/data/rke2/%s/token", clusterID), map[string]interface{}{
		"join_token": token,
		"cluster":    clusterID,
	})
}

// RetrieveJoinToken loads a join token using cluster ID
func (c *Client) RetrieveJoinToken(clusterID string) (string, error) {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/token", clusterID))
	if err != nil {
		return "", err
	}
	token, ok := data["join_token"].(string)
	if !ok {
		return "", fmt.Errorf("join_token not found for cluster %s", clusterID)
	}
	return token, nil
}

// StoreKubeConfig reads the kubeconfig from the host and uploads it to Vault
func (c *Client) StoreKubeConfig(clusterID, kubeconfigPath string) error {
	kubeconfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig from path '%s': %w", kubeconfigPath, err)
	}

	return c.StoreSecret(fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID), map[string]interface{}{
		"kubeconfig": string(kubeconfig),
	})
}

// RetrieveKubeConfig fetches the kubeconfig from Vault and saves it to the host
func (c *Client) RetrieveKubeConfig(clusterID, destinationPath string) error {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID))
	if err != nil {
		return fmt.Errorf("failed to retrieve kubeconfig for cluster %s: %w", clusterID, err)
	}

	kubeconfig, ok := data["kubeconfig"].(string)
	if !ok {
		return fmt.Errorf("kubeconfig not found or invalid type for cluster %s", clusterID)
	}

	// Create directory structure if it doesn't exist
	dir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	err = os.WriteFile(destinationPath, []byte(kubeconfig), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to path '%s': %w", destinationPath, err)
	}

	return nil
}

// StoreLBInfo stores information about a load balancer node
func (c *Client) StoreLBInfo(clusterID, hostname, vip string, isMain bool) error {
	path := fmt.Sprintf("kv/data/rke2/%s/lb/%s", clusterID, hostname)
	return c.StoreSecret(path, map[string]interface{}{
		"hostname": hostname,
		"vip":      vip,
		"is_main":  isMain,
	})
}

// RetrieveLBInfo retrieves information about load balancer nodes
func (c *Client) RetrieveLBInfo(clusterID string) ([]map[string]interface{}, string, error) {
	// List all LB entries for this cluster
	path := fmt.Sprintf("kv/metadata/rke2/%s/lb", clusterID)
	keys, err := c.ListKeys(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list load balancers for cluster %s: %w", clusterID, err)
	}

	lbNodes := []map[string]interface{}{}
	var vip string

	// Retrieve details for each LB
	for _, key := range keys {
		data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/lb/%s", clusterID, key))
		if err != nil {
			continue
		}

		lbNodes = append(lbNodes, data)

		// Get VIP from any LB (they should all have the same VIP)
		if vip == "" {
			vip, _ = data["vip"].(string)
		}

		// If this is the main LB, use its VIP as the definitive one
		isMain, ok := data["is_main"].(bool)
		if ok && isMain && data["vip"] != nil {
			vip = data["vip"].(string)
		}
	}

	if len(lbNodes) == 0 {
		return nil, "", fmt.Errorf("no load balancers found for cluster %s", clusterID)
	}

	return lbNodes, vip, nil
}

// StoreMasterInfo stores information about RKE2 master nodes and their configuration
func (c *Client) StoreMasterInfo(clusterID, hostname string, hosts []string, vip string) error {
	path := fmt.Sprintf("kv/data/rke2/%s/masters", clusterID)
	return c.StoreSecret(path, map[string]interface{}{
		"hosts":      hosts,
		"vip":        vip,
		"last_added": hostname,
	})
}

// RetrieveMasterInfo retrieves RKE2 master nodes information
func (c *Client) RetrieveMasterInfo(clusterID string) ([]string, string, error) {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/masters", clusterID))
	if err != nil {
		return nil, "", err
	}

	hostsRaw, ok := data["hosts"]
	if !ok {
		return nil, "", fmt.Errorf("hosts information not found for cluster %s", clusterID)
	}

	// Convert interface{} to string slice
	hosts := []string{}
	if hostsArray, ok := hostsRaw.([]interface{}); ok {
		for _, h := range hostsArray {
			if hostStr, ok := h.(string); ok {
				hosts = append(hosts, hostStr)
			}
		}
	}

	vip, _ := data["vip"].(string)

	return hosts, vip, nil
}
