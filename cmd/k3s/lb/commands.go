/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package lb

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/lb"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// Cmd is the top-level "k3s lb" command.
var Cmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage K3s load balancer",
	Long: `The "lb" command allows you to set up and manage HAProxy load balancers for K3s.

Examples:
  edgectl k3s lb create --cluster-id my-cluster --vip 192.168.10.100  # Create a new load balancer
  edgectl k3s lb status --cluster-id my-cluster                       # Check load balancer status
`,
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a load balancer for K3s",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s lb create command executed")

		if common.CheckRoot() != nil {
			os.Exit(1)
		}

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		vip, _ := cmd.Flags().GetString("vip")

		store := vault.InitVaultClient()
		if store == nil {
			os.Exit(1)
		}

		err := lb.CreateLoadBalancer(store, clusterID, vip, "k3s")
		if err != nil {
			fmt.Printf("❌ Failed to create load balancer: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ K3s load balancer created successfully")
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of K3s load balancer",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s lb status command executed")

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		store := vault.InitVaultClient()
		if store == nil {
			os.Exit(1)
		}

		vip, nodes, err := lb.GetStatus(store, clusterID)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve load balancer info: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("ℹ️ K3s Load balancer VIP: %s\n", vip)
		fmt.Println("ℹ️ Load balancer nodes:")

		for _, node := range nodes {
			role := "BACKUP"
			if node.IsMain {
				role = "MASTER"
			}
			fmt.Printf("  - %s (%s)\n", node.Hostname, role)
		}
	},
}

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up a load balancer for K3s",
	Long: `The "cleanup" command removes the load balancer configuration for a K3s cluster.
This includes disabling services (which also stops them) and removing configuration files.
The HAProxy and Keepalived packages will remain installed.

Example:
  edgectl k3s lb cleanup --cluster-id my-cluster  # Clean up LB and remove from secret store`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("k3s lb cleanup command executed")

		if common.CheckRoot() != nil {
			os.Exit(1)
		}

		logger.Debug("Extracting values from command line arguments")
		clusterID, _ := cmd.Flags().GetString("cluster-id")

		store := vault.InitVaultClient()
		if store == nil {
			os.Exit(1)
		}

		err := lb.CleanupLoadBalancer(store, clusterID)
		if err != nil {
			fmt.Printf("❌ Failed to clean up load balancer: %v\n", err)
			os.Exit(1)
		}
	},
}

// Initialize command flags and register subcommands
func init() {
	// Create command flags
	createCmd.Flags().String("cluster-id", "", "The ID of the cluster to create a load balancer for")
	createCmd.Flags().String("vip", "", "Virtual IP address for the load balancer")
	_ = createCmd.MarkFlagRequired("cluster-id")

	// Status command flags
	statusCmd.Flags().String("cluster-id", "", "The ID of the cluster to check load balancer status for")
	_ = statusCmd.MarkFlagRequired("cluster-id")

	// Cleanup command flags
	cleanupCmd.Flags().String("cluster-id", "", "The ID of the cluster to clean up the load balancer for")
	_ = cleanupCmd.MarkFlagRequired("cluster-id")

	// Register subcommands
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(cleanupCmd)
}
