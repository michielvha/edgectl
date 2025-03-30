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

// StoreSecret stores any secret (key-value map) under a Vault path
func (c *Client) StoreSecret(fullVaultPath string, data map[string]interface{}) error {
	_, err := c.VaultClient.Logical().Write(fullVaultPath, map[string]interface{}{
		"data": data,
	})
	if err != nil {
		return fmt.Errorf("failed to store secret at path '%s': %w", fullVaultPath, err)
	}
	return nil
}

// RetrieveSecret retrieves a key-value map from a Vault path
func (c *Client) RetrieveSecret(fullVaultPath string) (map[string]interface{}, error) {
	secret, err := c.VaultClient.Logical().Read(fullVaultPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read secret at path '%s': %w", fullVaultPath, err)
	}
	if secret == nil || secret.Data == nil {
		return nil, fmt.Errorf("no data found at path: %s", fullVaultPath)
	}
	data, ok := secret.Data["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid data format at path: %s", fullVaultPath)
	}
	return data, nil
}
