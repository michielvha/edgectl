package vault

import (
	"fmt"
	"os"
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

	err = os.WriteFile(destinationPath, []byte(kubeconfig), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to path '%s': %w", destinationPath, err)
	}

	return nil
}
