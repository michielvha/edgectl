/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package lb

import (
	"fmt"
	"os"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/lb"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
	"github.com/spf13/cobra"
)

// Cmd represents the "lb" command
var Cmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage RKE2 load balancer",
	Long: `The "lb" command allows you to set up and manage HAProxy load balancers for RKE2.
	
Examples:
  edgectl rke2 lb create --cluster-id my-cluster --vip 192.168.10.100  # Create a new load balancer
  edgectl rke2 lb status --cluster-id my-cluster                       # Check load balancer status
`,
}

// createCmd represents the "lb create" command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a load balancer for RKE2",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("lb create command executed")

		// Check if user is root
		if common.CheckRoot() != nil {
			os.Exit(1)
		}

		// Extract values
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		vip, _ := cmd.Flags().GetString("vip")

		// Create load balancer
		err := lb.CreateLoadBalancer(clusterID, vip)
		if err != nil {
			fmt.Printf("❌ Failed to create load balancer: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("✅ RKE2 load balancer created successfully")
	},
}

// statusCmd represents the "lb status" command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of RKE2 load balancer",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("lb status command executed")

		clusterID, _ := cmd.Flags().GetString("cluster-id")
		if clusterID == "" {
			fmt.Println("❌ Cluster ID is required.")
			_ = cmd.Help()
			os.Exit(1)
		}

		// Connect to Vault
		client, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Failed to create Vault client: %v\n", err)
			os.Exit(1)
		}

		// Get LB info
		lbNodes, vip, err := client.RetrieveLBInfo(clusterID)
		if err != nil {
			fmt.Printf("❌ Failed to retrieve load balancer info: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("ℹ️ RKE2 Load balancer VIP: %s\n", vip)
		fmt.Println("ℹ️ Load balancer nodes:")

		for _, node := range lbNodes {
			hostname := node["hostname"].(string)
			isMain := node["is_main"].(bool)

			role := "BACKUP"
			if isMain {
				role = "MASTER"
			}

			fmt.Printf("  - %s (%s)\n", hostname, role)
		}
	},
}

// cleanupCmd represents the "lb cleanup" command
var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Clean up a load balancer for RKE2",
	Long: `The "cleanup" command removes the load balancer configuration for an RKE2 cluster.
This includes disabling services (which also stops them) and removing configuration files.
The HAProxy and Keepalived packages will remain installed.

Example:
  edgectl rke2 lb cleanup --cluster-id my-cluster  # Clean up LB and remove from Vault`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("lb cleanup command executed")

		// Check if user is root
		if os.Geteuid() != 0 {
			fmt.Println("❌ This command must be run as root. Try using `sudo`.")
			os.Exit(1)
		}

		// Extract values
		clusterID, _ := cmd.Flags().GetString("cluster-id")

		// Cleanup load balancer
		err := lb.CleanupLoadBalancer(clusterID)
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
