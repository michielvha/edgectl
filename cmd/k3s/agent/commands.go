/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package agent

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/k3s/agent"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// Cmd is the top-level "k3s agent" command.
var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage K3s agent installation",
	Long: `The "agent" command allows you to install and manage K3s agents.

Examples:
  edgectl k3s agent install --cluster-id my-cluster  # Install K3s Agent
`,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install K3s Agent",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s agent install command executed")

		// Check if user is root
		if common.CheckRoot() != nil {
			os.Exit(1)
		}

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		vip, _ := cmd.Flags().GetString("vip")
		lbHostname, _ := cmd.Flags().GetString("lb-hostname")

		store := vault.InitVaultClient()
		if store == nil {
			os.Exit(1)
		}

		err := agent.Install(store, clusterID, vip, lbHostname)
		if err != nil {
			fmt.Printf("❌ K3s agent install failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ K3s agent installed successfully")
	},
}

// Initialize command flags and register subcommands
func init() {
	// Install command flags
	installCmd.Flags().String("cluster-id", "", "The ID of the cluster you want to join")
	installCmd.Flags().String("vip", "", "Virtual IP fallback if VIP is not found in secret store")
	installCmd.Flags().String("lb-hostname", "", "Load balancer hostname to resolve as VIP fallback (last resort)")
	_ = installCmd.MarkFlagRequired("cluster-id")

	// Register subcommands
	Cmd.AddCommand(installCmd)
}
