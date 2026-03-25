/*
Copyright © 2025 VH & Co - contact@vhco.pro

Package vault provides specialized handlers for cluster secrets management.

This file handles the kubeconfig management for Kubernetes clusters:
  - StoreKubeConfig: Reads the kubeconfig from a server, updates the API endpoint with VIP if provided,
    and stores it in the secret store
  - RetrieveKubeConfig: Fetches a kubeconfig from the secret store and writes it to a specified path on the host

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

// StoreKubeConfig reads the kubeconfig from the host, modifies it to use VIP if provided, and uploads it to the secret store
func (c *Client) StoreKubeConfig(distro, clusterID, kubeconfigPath, vip string) error {
	kubeconfig, err := os.ReadFile(kubeconfigPath) //nolint:gosec // path comes from trusted CLI input
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

	return c.StoreSecret(fmt.Sprintf("kv/data/%s/%s/kubeconfig", distro, clusterID), map[string]interface{}{
		"kubeconfig": kubeconfigStr,
	})
}

// RetrieveKubeConfig fetches the kubeconfig from the secret store and saves it to the host
func (c *Client) RetrieveKubeConfig(distro, clusterID, destinationPath string) error {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/%s/%s/kubeconfig", distro, clusterID))
	if err != nil {
		return fmt.Errorf("failed to retrieve kubeconfig for cluster %s: %w", clusterID, err)
	}

	kubeconfig, ok := data["kubeconfig"].(string)
	if !ok {
		return fmt.Errorf("kubeconfig not found or invalid type for cluster %s", clusterID)
	}

	// Create directory structure if it doesn't exist
	dir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	err = os.WriteFile(destinationPath, []byte(kubeconfig), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to path '%s': %w", destinationPath, err)
	}

	return nil
}
