/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Only supported on linux because bash dependencies and containers on windows.. yeah, nope.
*/
package cmd

import (
	"fmt"
	"os"

	common "github.com/michielvha/edgectl/pkg/common"
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

		common.FetchTokenFromVault(clusterID) // this will fetch the token and safe as env var to be used in bash function.
		// TODO: figure how to dynamically set lb hostname/ip as env var...
		common.RunBashFunction("rke2.sh", "install_rke2_agent -l 192.168.10.125")
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
	Use:   "config kube",
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

// Register subcommands
func init() {
	// Attach rke2 as rootCmd
	rootCmd.AddCommand(rke2Cmd)

	// installServerCmd Flags
	installServerCmd.Flags().String("cluster-id", "", "The clusterID required to join an existing cluster")
	// installAgentCmd Flags
	installAgentCmd.Flags().String("cluster-id", "", "The ID of the cluster you want to join")
	_ = installAgentCmd.MarkFlagRequired("cluster-id")

	SetKubeConfigCmd.Flags().String("cluster-id", "", "The ID of the cluster to fetch the kubeconfig for")
	SetKubeConfigCmd.Flags().String("output", "/etc/rancher/rke2/rke2.yaml", "Destination path to store the kubeconfig")
	_ = SetKubeConfigCmd.MarkFlagRequired("cluster-id")

	// Attach subcommands under rke2
	rke2Cmd.AddCommand(installServerCmd)
	rke2Cmd.AddCommand(installAgentCmd)
	rke2Cmd.AddCommand(statusCmd)
	rke2Cmd.AddCommand(uninstallCmd)
	rke2Cmd.AddCommand(SetKubeConfigCmd)
}
