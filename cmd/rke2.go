/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>

Only supported on linux because bash dependencies and containers on windows.. yeah, nope.
*/
package cmd

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	vault "github.com/michielvha/edgectl/pkg/vault/rke2"
	"github.com/spf13/cobra"
)

// TODO: Move functions to a separate package. Only keep the cobra command logic here.

//go:embed scripts/*.sh
var embeddedScripts embed.FS

// Extracts an embedded script to /tmp
func extractEmbeddedScript(scriptName string) string {
	scriptPath := filepath.Join("/tmp", scriptName)

	// Read script from embedded FS
	data, err := embeddedScripts.ReadFile("scripts/" + scriptName)
	if err != nil {
		fmt.Printf("❌ Failed to read embedded script: %v\n", err)
		os.Exit(1)
	}

	// Write to a temp file
	if err := os.WriteFile(scriptPath, data, 0755); err != nil {
		fmt.Printf("❌ Failed to write script: %v\n", err)
		os.Exit(1)
	}

	return scriptPath
}

// Runs a function from the sourced script
func runBashFunction(scriptName, functionName string) {
	scriptPath := extractEmbeddedScript(scriptName)

	// Run the function from the sourced script
	cmd := exec.Command("bash", "-c", fmt.Sprintf("source %s && %s", scriptPath, functionName))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error executing function %s from %s: %v\n", functionName, scriptPath, err)
		os.Exit(1)
	}
}

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

// Install RKE2 Server
// var installLoadBalancerCmd = &cobra.Command{
// 	Use:   "lb",
// 	Short: "Install RKE2 load balancer",
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Println("🚀 Install a load balancer for RKE2...")
// 		runBashFunction("rke2.sh", "install_rke2_lb")
// 	},
// }

// Install RKE2 Server
var installServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Install RKE2 Server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Installing RKE2 Server...")

		// Reuse our vault abstraction in ``pkg/vault/rke2``
		vaultClient, err := vault.NewClient()
		if err != nil {
			fmt.Printf("❌ Failed to initialize Vault client: %v\n", err)
			os.Exit(1)
		}

		clusterID, _ := cmd.Flags().GetString("cluster-id")

		if clusterID != "" {
			fmt.Println("🔐 Cluster ID supplied, retrieving join token from Vault...")

			token, err := vaultClient.RetrieveJoinToken(clusterID)
			if err != nil {
				fmt.Printf("❌ Failed to retrieve join token from Vault: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("✅ Retrieved token: %s\n", token)
			_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0644)
			_ = os.Setenv("RKE2_TOKEN", token)
		} else {
			clusterID = fmt.Sprintf("rke2-%s", uuid.New().String()[:8])
			_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0644)
			fmt.Printf("🆔 Generated cluster ID: %s\n", clusterID)
		}

		runBashFunction("rke2.sh", "install_rke2_server")

		// Store the token in vault if cluster-id wasn't supplied
		if !cmd.Flags().Changed("cluster-id") {
			tokenBytes, err := os.ReadFile("/var/lib/rancher/rke2/server/node-token")
			if err != nil {
				fmt.Printf("❌ Failed to read generated node token: %v\n", err)
				os.Exit(1)
			}

			token := strings.TrimSpace(string(tokenBytes))
			if err := vaultClient.StoreJoinToken(clusterID, token); err != nil {
				fmt.Printf("❌ Failed to store token in Vault: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("🔐 Token successfully stored in Vault for cluster %s\n", clusterID)
		}
	},
}

// Install RKE2 Agent
var installAgentCmd = &cobra.Command{
	Use:   "agent",
	Short: "Install RKE2 Agent",
	Run: func(cmd *cobra.Command, args []string) {
		clusterID, _ := cmd.Flags().GetString("cluster-id")
		if clusterID == "" {
			fmt.Println("❌ cluster ID is required to join an existing cluster.")
			os.Exit(1)
		}
		runBashFunction("rke2.sh", "install_rke2_agent")
	},
}

// Check RKE2 status
// TODO: Add more status checks
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of RKE2",
	Run: func(cmd *cobra.Command, args []string) {
		runBashFunction("rke2.sh", "rke2_status")
	},
}

// Uninstall RKE2
var uninstallCmd = &cobra.Command{
	Use:   "purge",
	Short: "purge RKE2 install from host",
	Run: func(cmd *cobra.Command, args []string) {
		runBashFunction("rke2.sh", "purge_rke2")
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
	if err := installAgentCmd.MarkFlagRequired("cluster-id"); err != nil {
		fmt.Printf("❌ Failed to mark cluster-id flag as required: %v\n", err)
		os.Exit(1)
	}

	// Attach subcommands under rke2
	rke2Cmd.AddCommand(installServerCmd)
	rke2Cmd.AddCommand(installAgentCmd)
	rke2Cmd.AddCommand(statusCmd)
	rke2Cmd.AddCommand(uninstallCmd)
}
