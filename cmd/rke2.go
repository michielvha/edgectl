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
	"github.com/google/uuid"
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
var installServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Install RKE2 Server",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("🚀 Installing RKE2 Server...")
		runBashFunction("rke2.sh", "install_rke2_server")

		token, _ := cmd.Flags().GetString("token") // get the token from the flag
		if token == "" {
			id := fmt.Sprintf("rke2-%s", uuid.New().String()[:8])
			_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(id), 0644)
			fmt.Printf("🆔 Generated new cluster ID: %s\n", id)
		} else {
			fmt.Println("🔐 Token supplied, skipping cluster ID generation.")
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
			fmt.Println("❌ cluster-id is required to join an existing cluster.")
			os.Exit(1)
		}
		runBashFunction("rke2.sh", "install_rke2_agent")
	},
}

// Check RKE2 status
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
	// Attach rke2 directly under rootCmd
	rootCmd.AddCommand(rke2Cmd)

	// Add flags to the installServerCmd and installAgentCmd
	installServerCmd.Flags().String("token", "", "The token required to join an existing cluster")
	installAgentCmd.Flags().String("cluster-id", "", "The ID of the cluster you want to join")
	installAgentCmd.MarkFlagRequired("cluster-id")

	// Attach subcommands under rke2
	rke2Cmd.AddCommand(installServerCmd)
	rke2Cmd.AddCommand(installAgentCmd)
	rke2Cmd.AddCommand(statusCmd)
	rke2Cmd.AddCommand(uninstallCmd)
}
