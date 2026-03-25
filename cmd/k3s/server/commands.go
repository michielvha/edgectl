/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package server

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/k3s/server"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// Cmd is the top-level "k3s server" command.
var Cmd = &cobra.Command{
	Use:   "server",
	Short: "Manage K3s server installation",
	Long: `The "server" command allows you to install and manage K3s servers.

Examples:
  edgectl k3s server install                            # Install new K3s Server
  edgectl k3s server install --cluster-id my-cluster    # Join existing K3s cluster as server
`,
}

// installCmd represents the "server install" command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install K3s Server",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s server install command executed")

		// Check if user is root
		if common.CheckRoot() != nil {
			os.Exit(1)
		}

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		isExisting := cmd.Flags().Changed("cluster-id")
		vip, _ := cmd.Flags().GetString("vip")

		store := vault.InitVaultClient()
		if store == nil {
			os.Exit(1)
		}

		err := server.Install(store, clusterID, isExisting, vip)
		if err != nil {
			fmt.Printf("❌ K3s server install failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ K3s server installed successfully")
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
