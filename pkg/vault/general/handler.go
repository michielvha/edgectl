package vault

// import (
// 	"fmt"
// 	"os"

// 	vault "github.com/hashicorp/vault/api"
// )

// // Client wraps the Vault client for reuse
// type Client struct {
// 	VaultClient *vault.Client
// }

// //TODO: generalize this for any kind of secret & create seperate RKE2 logic in a different package.
// // so currently the rke2 logic is in the vault package, check if we want to have some more generalized logic for storing & retrieving secrets

// // NewClient initializes a Vault client using environment variables (VAULT_ADDR, VAULT_TOKEN)
// func NewClient() (*Client, error) {
// 	config := vault.DefaultConfig() // uses VAULT_ADDR or default http://127.0.0.1:8200

// 	client, err := vault.NewClient(config)
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to create Vault client: %w", err)
// 	}

// 	token := os.Getenv("VAULT_TOKEN")
// 	if token == "" {
// 		return nil, fmt.Errorf("VAULT_TOKEN not set")
// 	}
// 	client.SetToken(token)

// 	return &Client{VaultClient: client}, nil
// }

// // StoreJoinToken saves the RKE2 join token to a Vault path
// func (c *Client) StoreToken(tokenPath, tokenKey, tokenValue string) error {
// 	path := fmt.Sprintf("kv/data/%s", tokenPath)

// 	_, err := c.VaultClient.Logical().Write(path, map[string]interface{}{
// 		"data": map[string]interface{}{
// 			tokenKey: tokenValue,
// 		},
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to write token to Vault: %w", err)
// 	}
// 	return nil
// }

// // Fetches token from Vault for a given cluster ID
// func (c *Client) RetrieveToken(tokenPath string) (string, error) {
// 	path := fmt.Sprintf("kv/data/%s", tokenPath)

// 	secret, err := c.VaultClient.Logical().Read(path)
// 	if err != nil {
// 		return "", fmt.Errorf("failed to read from Vault: %w", err)
// 	}
// 	if secret == nil || secret.Data == nil {
// 		return "", fmt.Errorf("no data found at path: %s", path)
// 	}

// 	data, ok := secret.Data["data"].(map[string]interface{})
// 	if !ok {
// 		return "", fmt.Errorf("invalid data format at path: %s", path)
// 	}

// 	token, ok := data["join_token"].(string)
// 	if !ok {
// 		return "", fmt.Errorf("join_token not found in Vault at path: %s", path)
// 	}

// 	return token, nil
// }
