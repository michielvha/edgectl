package vault

import (
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"
)

type Client struct {
	VaultClient *vault.Client
}

func NewClient() (*Client, error) {
	config := vault.DefaultConfig()
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Vault client: %w", err)
	}

	token := os.Getenv("VAULT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("VAULT_TOKEN not set")
	}
	client.SetToken(token)

	return &Client{VaultClient: client}, nil
}

// GetSecret reads a Vault KV v2 secret from a fully qualified path and returns its "data" map
func (c *Client) GetSecret(path string) (map[string]interface{}, error) {
	secret, err := c.VaultClient.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read from Vault at path '%s': %w", path, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path: %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format at path: %s", path)
	}
	return data, nil
}

// WriteSecret writes a map of values to a Vault KV v2 path
func (c *Client) WriteSecret(path string, data map[string]interface{}) error {
	_, err := c.VaultClient.Logical().Write(path, map[string]interface{}{"data": data})
	if err != nil {
		return fmt.Errorf("failed to write data to Vault at path '%s': %w", path, err)
	}
	return nil
}

// StoreJoinToken saves a token under a specific cluster path
func (c *Client) StoreJoinToken(clusterID, token string) error {
	return c.WriteSecret(fmt.Sprintf("kv/data/rke2/%s", clusterID), map[string]interface{}{
		"join_token": token,
		"cluster":    clusterID,
	})
}

// RetrieveJoinToken loads a join token using cluster ID
func (c *Client) RetrieveJoinToken(clusterID string) (string, error) {
	data, err := c.GetSecret(fmt.Sprintf("kv/data/rke2/%s", clusterID))
	if err != nil {
		return "", err
	}
	token, ok := data["join_token"].(string)
	if !ok {
		return "", fmt.Errorf("join_token not found for cluster %s", clusterID)
	}
	return token, nil
}

// RetrieveMasterInfo loads hostnames and VIP using cluster ID
func (c *Client) RetrieveMasterInfo(clusterID string) ([]string, string, error) {
	data, err := c.GetSecret(fmt.Sprintf("kv/data/rke2/%s/master-info", clusterID))
	if err != nil {
		return nil, "", err
	}

	rawHostnames, ok := data["hostnames"].([]interface{})
	if !ok {
		return nil, "", fmt.Errorf("hostnames not found or invalid type for cluster %s", clusterID)
	}

	var hostnames []string
	for _, h := range rawHostnames {
		if host, ok := h.(string); ok {
			hostnames = append(hostnames, host)
		}
	}

	vip, ok := data["vip"].(string)
	if !ok {
		return nil, "", fmt.Errorf("vip not found or invalid type for cluster %s", clusterID)
	}

	return hostnames, vip, nil
}

// StoreKubeConfig reads the kubeconfig from the host and uploads it to Vault
func (c *Client) StoreKubeConfig(clusterID, kubeconfigPath string) error {
	kubeconfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig from path '%s': %w", kubeconfigPath, err)
	}

	return c.WriteSecret(fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID), map[string]interface{}{
		"kubeconfig": string(kubeconfig),
	})
}

// RetrieveKubeConfig fetches the kubeconfig from Vault and saves it to the host
func (c *Client) RetrieveKubeConfig(clusterID, destinationPath string) error {
	data, err := c.GetSecret(fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID))
	if err != nil {
		return fmt.Errorf("failed to retrieve kubeconfig for cluster %s: %w", clusterID, err)
	}

	kubeconfig, ok := data["kubeconfig"].(string)
	if !ok {
		return fmt.Errorf("kubeconfig not found or invalid type for cluster %s", clusterID)
	}

	err = os.WriteFile(destinationPath, []byte(kubeconfig), 0600)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to path '%s': %w", destinationPath, err)
	}

	return nil
}