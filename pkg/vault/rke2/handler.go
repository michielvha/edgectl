package vault

import (
	"fmt"
	"os"

	vault "github.com/hashicorp/vault/api"
)

// Client wraps the Vault client for reuse
type Client struct {
	VaultClient *vault.Client
}

// NewClient initializes a Vault client using environment variables (VAULT_ADDR, VAULT_TOKEN)
func NewClient() (*Client, error) {
	config := vault.DefaultConfig() // uses VAULT_ADDR or default http://127.0.0.1:8200

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

// StoreJoinToken saves the RKE2 join token to a Vault path
// TODO: Expand this to also save the hostname ( will be needed in the future for the HAProxy config)
func (c *Client) StoreJoinToken(clusterID, token string) error {
	path := fmt.Sprintf("kv/data/rke2/%s", clusterID)

	_, err := c.VaultClient.Logical().Write(path, map[string]interface{}{
		"data": map[string]interface{}{
			"join_token": token,
			"cluster": clusterID,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to write join token to Vault: %w", err)
	}
	return nil
}

// RetrieveJoinToken fetches the join token from Vault for a given cluster ID
func (c *Client) RetrieveJoinToken(clusterID string) (string, error) {
	path := fmt.Sprintf("kv/data/rke2/%s", clusterID)

	secret, err := c.VaultClient.Logical().Read(path)
	if err != nil {
		return "", fmt.Errorf("failed to read from Vault: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return "", fmt.Errorf("no data found at path: %s", path)
	}

	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid data format at path: %s", path)
	}

	token, ok := data["join_token"].(string)
	if !ok {
		return "", fmt.Errorf("join_token not found in Vault at path: %s", path)
	}

	return token, nil
}

// RetrieveMasterInfo fetches the list of server hostnames and the VIP for a cluster from Vault
func (c *Client) RetrieveMasterInfo(clusterID string) ([]string, string, error) {
	path := fmt.Sprintf("kv/data/rke2/%s", clusterID)

	secret, err := c.VaultClient.Logical().Read(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read LB info from Vault: %w", err)
	}
	if secret == nil || secret.Data == nil {
		return nil, "", fmt.Errorf("no data found at path: %s", path)
	}

	rawData, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, "", fmt.Errorf("invalid data format at path: %s", path)
	}

	// Parse hostnames
	rawHostnames, ok := rawData["hostnames"].([]interface{})
	if !ok {
		return nil, "", fmt.Errorf("hostnames not found or wrong type at path: %s", path)
	}

	var hostnames []string
	for _, h := range rawHostnames {
		if host, ok := h.(string); ok {
			hostnames = append(hostnames, host)
		}
	}

	// Parse VIP
	vip, ok := rawData["vip"].(string)
	if !ok {
		return nil, "", fmt.Errorf("vip not found or invalid in Vault at path: %s", path)
	}

	return hostnames, vip, nil
}

