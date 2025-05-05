/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package server

import (
	"fmt"
	"os"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/rke2/server"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "server",
	Short: "Manage RKE2 server installation",
	Long: `The "server" command allows you to install and manage RKE2 servers.
	
Examples:
  edgectl rke2 server install                            # Install new RKE2 Server
  edgectl rke2 server install --cluster-id my-cluster    # Join existing RKE2 cluster as server
`,
}

// installCmd represents the "server install" command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install RKE2 Server",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("server install command executed")

		// Check if user is root
		if common.CheckRoot() != nil {
			os.Exit(1)
		}

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		isExisting := cmd.Flags().Changed("cluster-id")
		vip, _ := cmd.Flags().GetString("vip")

		err := server.Install(clusterID, isExisting, vip)
		if err != nil {
			fmt.Printf("❌ RKE2 server install failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ RKE2 server installed successfully")
	},
}

// Initialize command flags and register subcommands
func init() {
	// Install command flags
	installCmd.Flags().String("cluster-id", "", "The clusterID required to join an existing cluster")
	installCmd.Flags().String("vip", "", "Virtual IP to use for the load balancer (used for TLS SANs)")

	// Register subcommands
	Cmd.AddCommand(installCmd)
}
