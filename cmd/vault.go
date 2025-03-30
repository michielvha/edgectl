package cmd

import (
	"fmt"

	vault "github.com/michielvha/edgectl/pkg/vault"
	"github.com/spf13/cobra"
)

//TODO: Rework this package to be less specific for rke2, more like in general for using edge vault.
// handler done in pkg/vault/handler.go todo implement here after rke2 testing.

// initVaultClient centralizes Vault client creation and error handling
func initVaultClient() *vault.Client {
	client, err := vault.NewClient()
	if err != nil {
		fmt.Printf("❌ Vault client error: %v\n", err)
		return nil
	}
	return client
}

// Upload command
var vaultUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a token to Vault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 Uploading token to Vault...")

		client := initVaultClient()
		if client == nil {
			return
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		token, _ := cmd.Flags().GetString("token")

		err := client.StoreJoinToken(clusterID, token)
		if err != nil {
			fmt.Printf("❌ Failed to store token: %v\n", err)
			return
		}

		fmt.Println("✅ Token successfully stored in Vault.")
	},
}

// Fetch command
var vaultFetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch a token from Vault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔎 Fetching token from Vault...")

		client := initVaultClient()
		if client == nil {
			return
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		token, err := client.RetrieveJoinToken(clusterID)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve token: %v\n", err)
			return
		}

		fmt.Printf("✅ Retrieved token: %s\n", token)
	},
}

func init() {
	// Parent command: edgectl vault
	var vaultCmd = &cobra.Command{
		Use:   "vault",
		Short: "Interact with Edge Vault",
	}

	// upload flags
	vaultUploadCmd.Flags().String("cluster-id", "test-cluster", "Cluster ID to store the token under")
	vaultUploadCmd.Flags().String("token", "dummy-token", "The token to upload")

	// fetch flags
	vaultFetchCmd.Flags().String("cluster-id", "test-cluster", "Cluster ID to fetch the token from")

	vaultCmd.AddCommand(vaultUploadCmd)
	vaultCmd.AddCommand(vaultFetchCmd)
	rootCmd.AddCommand(vaultCmd)
}
