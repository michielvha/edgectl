/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package cmd

import (
	"fmt"

	"github.com/michielvha/edgectl/pkg/vault"
	"github.com/spf13/cobra"
)

// --- Generic commands ---

// Get a specific key from a Vault KV v2 path
var vaultGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a secret value from Vault",
	Long: `Retrieve a specific key from a Vault KV v2 path.

Example:
  edgectl vault get --path kv/data/rke2/my-cluster/token --key token`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultClient := vault.InitVaultClient()
		if vaultClient == nil {
			return
		}

		path, _ := cmd.Flags().GetString("path")
		key, _ := cmd.Flags().GetString("key")

		data, err := vaultClient.RetrieveSecret(path)
		if err != nil {
			fmt.Printf("❌ Failed to read secret: %v\n", err)
			return
		}

		if key != "" {
			val, ok := data[key]
			if !ok {
				fmt.Printf("❌ Key '%s' not found at path '%s'\n", key, path)
				return
			}
			fmt.Printf("%v\n", val)
		} else {
			// Print all keys
			for k, v := range data {
				fmt.Printf("%s: %v\n", k, v)
			}
		}
	},
}

// Set a key-value pair at a Vault KV v2 path
var vaultSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a secret value in Vault",
	Long: `Store a key-value pair at a Vault KV v2 path.

Example:
  edgectl vault set --path kv/data/myapp/config --key api_url --value https://example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		vaultClient := vault.InitVaultClient()
		if vaultClient == nil {
			return
		}

		path, _ := cmd.Flags().GetString("path")
		key, _ := cmd.Flags().GetString("key")
		value, _ := cmd.Flags().GetString("value")

		err := vaultClient.StoreSecret(path, map[string]interface{}{
			key: value,
		})
		if err != nil {
			fmt.Printf("❌ Failed to store secret: %v\n", err)
			return
		}

		fmt.Printf("✅ Stored '%s' at '%s'\n", key, path)
	},
}

// --- RKE2-specific convenience commands ---

var vaultUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an RKE2 join token to Vault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 Uploading token to Vault...")

		vaultClient := vault.InitVaultClient()
		if vaultClient == nil {
			return
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		token, _ := cmd.Flags().GetString("token")

		err := vaultClient.StoreJoinToken(clusterID, token)
		if err != nil {
			fmt.Printf("❌ Failed to store token: %v\n", err)
			return
		}

		fmt.Println("✅ Token successfully stored in Vault.")
	},
}

var vaultFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch an RKE2 join token from Vault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔎 Fetching token from Vault...")

		vaultClient := vault.InitVaultClient()
		if vaultClient == nil {
			return
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		token, err := vaultClient.RetrieveJoinToken(clusterID)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve token: %v\n", err)
			return
		}

		fmt.Printf("✅ Retrieved token: %s\n", token)
	},
}

func init() {
	// Parent command: edgectl vault
	vaultCmd := &cobra.Command{
		Use:   "vault",
		Short: "Interact with Edge Vault",
	}

	// get flags
	vaultGetCmd.Flags().String("path", "", "Vault KV v2 path (e.g. kv/data/myapp/config)")
	vaultGetCmd.Flags().String("key", "", "Specific key to retrieve (omit to list all keys)")
	_ = vaultGetCmd.MarkFlagRequired("path")

	// set flags
	vaultSetCmd.Flags().String("path", "", "Vault KV v2 path (e.g. kv/data/myapp/config)")
	vaultSetCmd.Flags().String("key", "", "Key to store")
	vaultSetCmd.Flags().String("value", "", "Value to store")
	_ = vaultSetCmd.MarkFlagRequired("path")
	_ = vaultSetCmd.MarkFlagRequired("key")
	_ = vaultSetCmd.MarkFlagRequired("value")

	// upload flags (RKE2 convenience)
	vaultUploadCmd.Flags().String("cluster-id", "test-cluster", "Cluster ID to store the token under")
	vaultUploadCmd.Flags().String("token", "dummy-token", "The token to upload")

	// fetch flags (RKE2 convenience)
	vaultFetchCmd.Flags().String("cluster-id", "test-cluster", "Cluster ID to fetch the token from")

	vaultCmd.AddCommand(vaultGetCmd)
	vaultCmd.AddCommand(vaultSetCmd)
	vaultCmd.AddCommand(vaultUploadCmd)
	vaultCmd.AddCommand(vaultFetchCmd)
	rootCmd.AddCommand(vaultCmd)
}
