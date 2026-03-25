/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/michielvha/edgectl/pkg/vault"
)

// --- Generic commands ---

// Get a specific key from a KV v2 path
var secretsGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a secret value from the secret store",
	Long: `Retrieve a specific key from a KV v2 path.

Example:
  edgectl secrets get --path kv/data/<distro>/my-cluster/token --key token`,
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

// Set a key-value pair at a KV v2 path
var secretsSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a secret value in the secret store",
	Long: `Store a key-value pair at a KV v2 path.

Example:
  edgectl secrets set --path kv/data/myapp/config --key api_url --value https://example.com`,
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

var secretsUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload an RKE2 join token to the secret store",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 Uploading token to secret store...")

		vaultClient := vault.InitVaultClient()
		if vaultClient == nil {
			return
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		token, _ := cmd.Flags().GetString("token")
		distro, _ := cmd.Flags().GetString("distro")

		err := vaultClient.StoreJoinToken(distro, clusterID, token)
		if err != nil {
			fmt.Printf("❌ Failed to store token: %v\n", err)
			return
		}

		fmt.Println("✅ Token successfully stored in secret store.")
	},
}

var secretsFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch an RKE2 join token from the secret store",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔎 Fetching token from secret store...")

		vaultClient := vault.InitVaultClient()
		if vaultClient == nil {
			return
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		distro, _ := cmd.Flags().GetString("distro")
		token, err := vaultClient.RetrieveJoinToken(distro, clusterID)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve token: %v\n", err)
			return
		}

		fmt.Printf("✅ Retrieved token: %s\n", token)
	},
}

func init() {
	// Parent command: edgectl secrets
	secretsCmd := &cobra.Command{
		Use:   "secrets",
		Short: "Interact with the secret store",
	}

	// get flags
	secretsGetCmd.Flags().String("path", "", "KV v2 path (e.g. kv/data/myapp/config)")
	secretsGetCmd.Flags().String("key", "", "Specific key to retrieve (omit to list all keys)")
	_ = secretsGetCmd.MarkFlagRequired("path")

	// set flags
	secretsSetCmd.Flags().String("path", "", "KV v2 path (e.g. kv/data/myapp/config)")
	secretsSetCmd.Flags().String("key", "", "Key to store")
	secretsSetCmd.Flags().String("value", "", "Value to store")
	_ = secretsSetCmd.MarkFlagRequired("path")
	_ = secretsSetCmd.MarkFlagRequired("key")
	_ = secretsSetCmd.MarkFlagRequired("value")

	// upload flags
	secretsUploadCmd.Flags().String("cluster-id", "test-cluster", "Cluster ID to store the token under")
	secretsUploadCmd.Flags().String("token", "dummy-token", "The token to upload")
	secretsUploadCmd.Flags().String("distro", "rke2", "Cluster distribution (rke2 or k3s)")

	// fetch flags
	secretsFetchCmd.Flags().String("cluster-id", "test-cluster", "Cluster ID to fetch the token from")
	secretsFetchCmd.Flags().String("distro", "rke2", "Cluster distribution (rke2 or k3s)")

	secretsCmd.AddCommand(secretsGetCmd)
	secretsCmd.AddCommand(secretsSetCmd)
	secretsCmd.AddCommand(secretsUploadCmd)
	secretsCmd.AddCommand(secretsFetchCmd)
	rootCmd.AddCommand(secretsCmd)
}
