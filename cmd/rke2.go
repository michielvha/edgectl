/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Only supported on linux because bash dependencies and containers on windows.. yeah, nope.
*/
package cmd

import (
	"fmt"
	"os"

	common "github.com/michielvha/edgectl/pkg/common"
	lb "github.com/michielvha/edgectl/pkg/lb"
	"github.com/michielvha/edgectl/pkg/logger"
	server "github.com/michielvha/edgectl/pkg/rke2/server" // Import the new package
	vault "github.com/michielvha/edgectl/pkg/vault"
	"github.com/spf13/cobra"
)

// rke2Cmd represents the "rke2" command
var rke2Cmd = &cobra.Command{
	Use:   "rke2",
	Short: "Manage RKE2 cluster",
	Long: `The "rke2" command allows you to install, manage, and uninstall RKE2.

Examples:
  edgectl rke2 server      # Install RKE2 Server
  edgectl rke2 agent       # Install RKE2 Agent
  edgectl rke2 uninstall   # Uninstall RKE2
`,
}

var installServerCmd = &cobra.Command{
	Use:   "server install",
	Short: "Install RKE2 Server",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("server install command executed")
		// ✅ Check if user is root
		if os.Geteuid() != 0 {
			fmt.Println("❌ This command must be run as root. Try using `sudo`.")
			os.Exit(1)
		}

		// ✅ Extract values and call logic
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		isExisting := cmd.Flags().Changed("cluster-id")

		err := server.Install(clusterID, isExisting)
		if err != nil {
			fmt.Printf("❌ RKE2 server install failed: %v\n", err)
			os.Exit(1)
		}
	},
}

// Install RKE2 Agent
var installAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Install RKE2 Agent",
	Run: func(cmd *cobra.Command, args []string) {
		clusterID, _ := cmd.Flags().GetString("cluster-id")

		// Create Vault client
		vaultClient, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Failed to create Vault client: %v\n", err)
			os.Exit(1)
		}

		// Retrieve token from Vault
		if _, err := server.FetchTokenFromVault(clusterID); err != nil {
			fmt.Printf("❌ Failed to fetch token from Vault: %v\n", err)
			os.Exit(1)
		}

		// Get LB info (VIP) from Vault
		_, vip, err := vaultClient.RetrieveLBInfo(clusterID)
		if err != nil {
			fmt.Printf("⚠️  No load balancer found for cluster %s. Using hostname from arguments.\n", clusterID)
			// Try to use loadbalancer hostname if provided
			lbHostname, _ := cmd.Flags().GetString("lb-hostname")
			if lbHostname == "" {
				fmt.Printf("❌ No load balancer VIP found and no lb-hostname provided.\n")
				os.Exit(1)
			}
			common.RunBashFunction("rke2.sh", fmt.Sprintf("install_rke2_agent -l %s", lbHostname))
			return
		}

		fmt.Printf("ℹ️ Using load balancer VIP: %s\n", vip)
		common.RunBashFunction("rke2.sh", fmt.Sprintf("install_rke2_agent -l %s", vip))
	},
}

// Check RKE2 status
// TODO: Add more status checks
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of RKE2",
	Run: func(cmd *cobra.Command, args []string) {
		common.RunBashFunction("rke2.sh", "rke2_status")
	},
}

// Uninstall RKE2
var uninstallCmd = &cobra.Command{
	Use:   "purge",
	Short: "purge RKE2 install from host",
	Run: func(cmd *cobra.Command, args []string) {
		common.RunBashFunction("rke2.sh", "purge_rke2")
	},
}

// Configure kubeconfig
var SetKubeConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Fetch kubeconfig from Vault and store it on the host",
	Run: func(cmd *cobra.Command, args []string) {
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

// LB commands
var lbCmd = &cobra.Command{
	Use:   "lb",
	Short: "Manage RKE2 load balancer",
	Long: `The "lb" command allows you to set up and manage HAProxy load balancers for RKE2.
	
Examples:
  edgectl rke2 lb create --cluster-id my-cluster --vip 192.168.10.100  # Create a new load balancer
  edgectl rke2 lb status --cluster-id my-cluster                       # Check load balancer status
`,
}

// Create LB command
var lbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a load balancer for RKE2",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("lb create command executed")

		// Check if user is root
		if os.Geteuid() != 0 {
			fmt.Println("❌ This command must be run as root. Try using `sudo`.")
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

// LB Status command
var lbStatusCmd = &cobra.Command{
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

// Register subcommands
func init() {
	// Attach rke2 as rootCmd
	rootCmd.AddCommand(rke2Cmd)

	// installServerCmd Flags
	installServerCmd.Flags().String("cluster-id", "", "The clusterID required to join an existing cluster")
	// installAgentCmd Flags
	installAgentCmd.Flags().String("cluster-id", "", "The ID of the cluster you want to join")
	installAgentCmd.Flags().String("lb-hostname", "", "The hostname of the load balancer to use if VIP is not found")
	_ = installAgentCmd.MarkFlagRequired("cluster-id")

	SetKubeConfigCmd.Flags().String("cluster-id", "", "The ID of the cluster to fetch the kubeconfig for")
	SetKubeConfigCmd.Flags().String("output", "/etc/rancher/rke2/rke2.yaml", "Destination path to store the kubeconfig")
	_ = SetKubeConfigCmd.MarkFlagRequired("cluster-id")

	// LB command flags
	lbCreateCmd.Flags().String("cluster-id", "", "The ID of the cluster to create a load balancer for")
	lbCreateCmd.Flags().String("vip", "", "Virtual IP address for the load balancer")
	_ = lbCreateCmd.MarkFlagRequired("cluster-id")

	lbStatusCmd.Flags().String("cluster-id", "", "The ID of the cluster to check load balancer status for")
	_ = lbStatusCmd.MarkFlagRequired("cluster-id")

	// Add LB commands
	lbCmd.AddCommand(lbCreateCmd)
	lbCmd.AddCommand(lbStatusCmd)
	rke2Cmd.AddCommand(lbCmd)

	// Attach subcommands under rke2
	rke2Cmd.AddCommand(installServerCmd)
	rke2Cmd.AddCommand(installAgentCmd)
	rke2Cmd.AddCommand(statusCmd)
	rke2Cmd.AddCommand(uninstallCmd)
	rke2Cmd.AddCommand(SetKubeConfigCmd)
}
