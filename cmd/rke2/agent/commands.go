/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package agent

import (
	"fmt"
	"os"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/rke2/agent"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "agent",
	Short: "Manage RKE2 agent installation",
	Long: `The "agent" command allows you to install and manage RKE2 agents.
	
Examples:
  edgectl rke2 agent install --cluster-id my-cluster  # Install RKE2 Agent
`,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install RKE2 Agent",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("agent install command executed")

		// Check if user is root
		if common.CheckRoot() != nil {
			os.Exit(1)
		}

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		vip, _ := cmd.Flags().GetString("vip")
		lbHostname, _ := cmd.Flags().GetString("lb-hostname")

		err := agent.Install(clusterID, vip, lbHostname)
		if err != nil {
			fmt.Printf("❌ RKE2 agent install failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ RKE2 agent installed successfully")
	},
}

// Initialize command flags and register subcommands
func init() {
	// Install command flags
	installCmd.Flags().String("cluster-id", "", "The ID of the cluster you want to join")
	installCmd.Flags().String("vip", "", "Virtual IP fallback if VIP is not found in Vault")
	installCmd.Flags().String("lb-hostname", "", "Load balancer hostname to resolve as VIP fallback (last resort)")
	_ = installCmd.MarkFlagRequired("cluster-id")

	// Register subcommands
	Cmd.AddCommand(installCmd)
}
