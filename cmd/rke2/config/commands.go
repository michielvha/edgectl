/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package config

import (
	"fmt"
	"os"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
	"github.com/spf13/cobra"
)

// Cmd represents the "config" command
var Cmd = &cobra.Command{
	Use:   "config",
	Short: "Manage RKE2 configuration",
	Long: `The "config" command allows you to manage RKE2 configuration.
	
Examples:
  edgectl rke2 config kubeconfig --cluster-id my-cluster  # Fetch kubeconfig from Vault
`,
}

// kubeconfigCmd represents the "config kubeconfig" command
var kubeconfigCmd = &cobra.Command{
	Use:   "kubeconfig",
	Short: "Fetch kubeconfig from Vault and store it on the host",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("config kubeconfig command executed")

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
		common.RunBashFunction("rke2.sh", "configure_rke2_bash")
	},
}

// bashCmd represents the "config bash" command
var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Configure the bash environment for RKE2",
	Long:  `Configures the bash environment to use RKE2 binaries and kubeconfig.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("config bash command executed")

		fmt.Println("🔧 Configuring bash environment for RKE2...")
		common.RunBashFunction("rke2.sh", "configure_rke2_bash")
		fmt.Println("✅ Bash environment configured for RKE2")
	},
}

// Initialize command flags and register subcommands
func init() {
	// Kubeconfig command flags
	kubeconfigCmd.Flags().String("cluster-id", "", "The ID of the cluster to fetch the kubeconfig for")
	kubeconfigCmd.Flags().String("output", "/etc/rancher/rke2/rke2.yaml", "Destination path to store the kubeconfig")
	_ = kubeconfigCmd.MarkFlagRequired("cluster-id")

	// Register subcommands
	Cmd.AddCommand(kubeconfigCmd)
	Cmd.AddCommand(bashCmd)
}
