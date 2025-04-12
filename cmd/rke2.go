/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Only supported on linux because bash dependencies and containers on windows.. yeah, nope.
*/
package cmd

import (
	"fmt"
	"os"

	lbcmd "github.com/michielvha/edgectl/cmd/rke2/lb"
	common "github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	server "github.com/michielvha/edgectl/pkg/rke2/server"
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

// TODO: rework to global flag ? create instead of install, or something that works better for maintainability & scalability.
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
		vip, _ := cmd.Flags().GetString("vip")

		err := server.Install(clusterID, isExisting, vip)
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
// TODO: When purging a master cluster we should also remove the state from the vault.
var uninstallCmd = &cobra.Command{
	Use:   "purge",
	Short: "purge RKE2 install from host",
	Run: func(cmd *cobra.Command, args []string) {
		common.RunBashFunction("rke2.sh", "purge_rke2")
	},
}

// Configure kubeconfig
var setKubeConfigCmd = &cobra.Command{
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

// New command to just configure the bash environment
// TODO: create global set package for `set config` & `set bash`
var bashCmd = &cobra.Command{
	Use:   "bash",
	Short: "Configure the bash environment for RKE2",
	Long:  `Configures the bash environment to use RKE2 binaries and kubeconfig.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🔧 Configuring bash environment for RKE2...")
		common.RunBashFunction("rke2.sh", "configure_rke2_bash")
		fmt.Println("✅ Bash environment configured for RKE2")
	},
}

// Register subcommands
func init() {
	// Attach rke2 as rootCmd
	rootCmd.AddCommand(rke2Cmd)

	// installServerCmd Flags
	installServerCmd.Flags().String("cluster-id", "", "The clusterID required to join an existing cluster")
	installServerCmd.Flags().String("vip", "", "Virtual IP to use for the load balancer (used for TLS SANs)")
	// installAgentCmd Flags
	installAgentCmd.Flags().String("cluster-id", "", "The ID of the cluster you want to join")
	installAgentCmd.Flags().String("lb-hostname", "", "The hostname of the load balancer to use if VIP is not found")
	_ = installAgentCmd.MarkFlagRequired("cluster-id")

	setKubeConfigCmd.Flags().String("cluster-id", "", "The ID of the cluster to fetch the kubeconfig for")
	setKubeConfigCmd.Flags().String("output", "/etc/rancher/rke2/rke2.yaml", "Destination path to store the kubeconfig")
	_ = setKubeConfigCmd.MarkFlagRequired("cluster-id")

	// Attach subcommands under rke2
	rke2Cmd.AddCommand(installServerCmd)
	rke2Cmd.AddCommand(installAgentCmd)
	rke2Cmd.AddCommand(statusCmd)
	rke2Cmd.AddCommand(uninstallCmd)
	rke2Cmd.AddCommand(setKubeConfigCmd)
	rke2Cmd.AddCommand(bashCmd)

	// Add loadbalancer command from the new package
	rke2Cmd.AddCommand(lbcmd.Cmd)
}
