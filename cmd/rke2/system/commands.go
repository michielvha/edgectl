/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package system

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
	"github.com/spf13/cobra"
)

// Get user home directory for storing kubeconfig
var (
	userHomeDir, _ = os.UserHomeDir()
)

var Cmd = &cobra.Command{
	Use:   "system",
	Short: "Manage RKE2 system operations",
	Long: `The "system" command provides operations for RKE2 system management.
	
Examples:
  edgectl rke2 system status      # Check status of RKE2
  edgectl rke2 system purge       # Uninstall RKE2 from the host
  edgectl rke2 system kubeconfig  # Fetch kubeconfig from Vault
  edgectl rke2 system bash        # Configure bash environment for RKE2
`,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of RKE2",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("system status command executed")
		common.RunBashFunction("rke2-status.sh", "rke2_status")
	},
}

// TODO: Enhance function to remove all state from vault via `GoVaultClient`
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge RKE2 install from host",
	Long:  `Completely removes RKE2 installation from the host.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("system purge command executed")
		fmt.Println("🗑️  Purging RKE2 from the host...")
		common.RunBashFunction("rke2-purge.sh", "rke2_purge")
		fmt.Println("✅ RKE2 purged successfully")
	},
}

var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Fetch kubeconfig from Vault and store it on the host",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("system kubeconfig command executed")

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		outputPath, _ := cmd.Flags().GetString("output")

		vaultClient, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Failed to initialize Vault client: %v\n", err)
			os.Exit(1)
		}

		err = vaultClient.RetrieveKubeConfig(clusterID, outputPath)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve kubeconfig: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Kubeconfig successfully written to: %s\n", outputPath)

		// Configure bash shell to use the kubeconfig
		common.RunBashFunction("rke2-bash.sh", "setup_kubectl_bash_env")
	},
}

var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Configure the bash environment for RKE2",
	Long:  `Configures the bash environment to use RKE2 binaries and kubeconfig.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("system bash command executed")

		fmt.Println("🔧 Configuring bash environment for RKE2...")
		common.RunBashFunction("rke2-bash.sh", "setup_rke2_node_bash_env")
		fmt.Println("✅ Bash environment configured for RKE2")
	},
}

// Initialize and register subcommands
func init() {
	// Kubeconfig command flags
	kubeconfigCmd.Flags().String("cluster-id", "", "The ID of the cluster to fetch the kubeconfig for")
	// Set default output path for kubeconfig generated from userHomeDir
	homeBasedKubeconfig := filepath.Join(userHomeDir, ".kube/config")
	kubeconfigCmd.Flags().String("output", homeBasedKubeconfig, "Destination path to store the kubeconfig")

	_ = kubeconfigCmd.MarkFlagRequired("cluster-id")

	// Register subcommands
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(purgeCmd)
	Cmd.AddCommand(kubeconfigCmd)
	Cmd.AddCommand(bashCmd)
}
