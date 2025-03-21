package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/michielvha/edgectl/pkg/vault"
)

var (
	testClusterID string
	testToken     string
)

var vaultUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a test join token to Vault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 Uploading test token to Vault...")

		client, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Vault client error: %v\n", err)
			return
		}

		err = client.StoreJoinToken(testClusterID, testToken)
		if err != nil {
			fmt.Printf("❌ Failed to store token: %v\n", err)
			return
		}

		fmt.Println("✅ Token successfully stored in Vault.")
	},
}

func init() {
	// Parent command: edgectl vault
	var vaultCmd = &cobra.Command{
		Use:   "vault",
		Short: "Interact with Vault for testing",
	}

	vaultUploadCmd.Flags().StringVar(&testClusterID, "cluster-id", "test-cluster", "Cluster ID to store the token under")
	vaultUploadCmd.Flags().StringVar(&testToken, "token", "dummy-token", "The join token to upload")

	vaultCmd.AddCommand(vaultUploadCmd)
	rootCmd.AddCommand(vaultCmd)
}
