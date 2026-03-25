/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// Get user home directory for storing kubeconfig
var (
	userHomeDir, _ = os.UserHomeDir()
)

// Cmd is the top-level "k3s system" command.
var Cmd = &cobra.Command{
	Use:   "system",
	Short: "Manage K3s system operations",
	Long: `The "system" command provides operations for K3s system management.

Examples:
  edgectl k3s system status      # Check status of K3s
  edgectl k3s system purge       # Uninstall K3s from the host
  edgectl k3s system kubeconfig  # Fetch kubeconfig from secret store
  edgectl k3s system bash        # Configure bash environment for K3s
`,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of K3s",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s system status command executed")
		common.RunBashFunction("k3s-status.sh", "k3s_status")
	},
}

var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge K3s install from host",
	Long: `Completely removes K3s installation from the host.
If --cluster-id is provided, also removes all cluster data from the secret store.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s system purge command executed")
		fmt.Println("🗑️  Purging K3s from the host...")
		common.RunBashFunction("k3s-purge.sh", "k3s_purge")
		fmt.Println("✅ K3s purged successfully")

		// If cluster-id is provided, also clean up secret store data
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		if clusterID != "" {
			fmt.Printf("🔄 Removing cluster data from secret store for %s...\n", clusterID)
			vaultClient := vault.InitVaultClient()
			if vaultClient == nil {
				fmt.Println("⚠️  Could not connect to secret store — skipping remote cleanup")
				return
			}
			if err := vaultClient.DeleteClusterData("k3s", clusterID); err != nil {
				fmt.Printf("⚠️  Secret store cleanup completed with warnings: %v\n", err)
			} else {
				fmt.Println("✅ Cluster data removed from secret store")
			}
		}
	},
}

var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Fetch kubeconfig from the secret store and store it on the host",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s system kubeconfig command executed")

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		outputPath, _ := cmd.Flags().GetString("output")

		vaultClient := vault.InitVaultClient()
		if vaultClient == nil {
			os.Exit(1)
		}

		err := vaultClient.RetrieveKubeConfig("k3s", clusterID, outputPath)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve kubeconfig: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Kubeconfig successfully written to: %s\n", outputPath)

		// Configure bash shell to use the kubeconfig
		common.RunBashFunction("k3s-bash.sh", "setup_kubectl_bash_env")
	},
}

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Configure the bash environment for K3s",
	Long:  `Configures the bash environment to use K3s kubeconfig.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s system bash command executed")

		fmt.Println("🔧 Configuring bash environment for K3s...")
		common.RunBashFunction("k3s-bash.sh", "setup_k3s_node_bash_env")
		fmt.Println("✅ Bash environment configured for K3s")
	},
}

// Initialize and register subcommands
func init() {
	// Kubeconfig command flags
	kubeconfigCmd.Flags().String("cluster-id", "", "The ID of the cluster to fetch the kubeconfig for")
	homeBasedKubeconfig := filepath.Join(userHomeDir, ".kube", "config")
	kubeconfigCmd.Flags().String("output", homeBasedKubeconfig, "Destination path to store the kubeconfig")

	_ = kubeconfigCmd.MarkFlagRequired("cluster-id")

	// Purge command flags
	purgeCmd.Flags().String("cluster-id", "", "Cluster ID to also remove data from the secret store (optional)")

	// Register subcommands
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(purgeCmd)
	Cmd.AddCommand(kubeconfigCmd)
	Cmd.AddCommand(bashCmd)
}
