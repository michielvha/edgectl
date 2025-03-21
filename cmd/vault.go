package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/michielvha/edgectl/pkg/vault"
)

var (
	ClusterID string
	Token     string
)

// TODO: Rework to more objective functionality of storing a token in Vault, current use is for testing rke2 functionality
var vaultUploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a join token to Vault",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔐 Uploading token to Vault...")

		client, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Vault client error: %v\n", err)
			return
		}

		err = client.StoreJoinToken(ClusterID, Token)
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
		Short: "Interact with Edge Vault",
	}
    // upload flags
	vaultUploadCmd.Flags().StringVar(&ClusterID, "cluster-id", "test-cluster", "Cluster ID to store the token under")
	vaultUploadCmd.Flags().StringVar(&Token, "token", "dummy-token", "The join token to upload")

	vaultCmd.AddCommand(vaultUploadCmd)
	rootCmd.AddCommand(vaultCmd)
}
