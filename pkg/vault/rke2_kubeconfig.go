/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Package vault provides specialized handlers for RKE2 cluster secrets management.

This file handles the kubeconfig management for RKE2 clusters:
  - StoreKubeConfig: Reads the kubeconfig from a server, updates the API endpoint with VIP if provided,
    and stores it in Vault
  - RetrieveKubeConfig: Fetches a kubeconfig from Vault and writes it to a specified path on the host

These functions enable secure kubeconfig sharing between cluster members and administrators
without requiring direct SSH access to the control plane nodes.
*/
package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StoreKubeConfig reads the kubeconfig from the host, modifies it to use VIP if provided, and uploads it to Vault
func (c *Client) StoreKubeConfig(clusterID, kubeconfigPath string, vip string) error {
	kubeconfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig from path '%s': %w", kubeconfigPath, err)
	}

	// Replace the localhost URL with the VIP in the kubeconfig if VIP is provided
	kubeconfigStr := string(kubeconfig)
	if vip != "" {
		kubeconfigStr = strings.ReplaceAll(
			kubeconfigStr,
			"server: https://127.0.0.1:6443",
			fmt.Sprintf("server: https://%s:6443", vip),
		)
		fmt.Printf("🔄 Updated kubeconfig to use VIP: %s\n", vip)
	}

	return c.StoreSecret(fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID), map[string]interface{}{
		"kubeconfig": kubeconfigStr,
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
