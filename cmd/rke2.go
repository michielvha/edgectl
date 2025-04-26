/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Only supported on linux because bash dependencies and containers on windows.. yeah, nope.
*/
package cmd

import (
	agentcmd "github.com/michielvha/edgectl/cmd/rke2/agent"
	configcmd "github.com/michielvha/edgectl/cmd/rke2/config"
	lbcmd "github.com/michielvha/edgectl/cmd/rke2/lb"
	servercmd "github.com/michielvha/edgectl/cmd/rke2/server"
	systemcmd "github.com/michielvha/edgectl/cmd/rke2/system"
	"github.com/spf13/cobra"
)

// rke2Cmd represents the "rke2" command
var rke2Cmd = &cobra.Command{
	Use:   "rke2",
	Short: "Manage RKE2 cluster",
	Long: `The "rke2" command allows you to install, manage, and uninstall RKE2.

Examples:
  edgectl rke2 server install       # Install RKE2 Server
  edgectl rke2 agent install        # Install RKE2 Agent
  edgectl rke2 system purge         # Uninstall RKE2
`,
}

// Register subcommands
func init() {
	// Attach rke2 as rootCmd
	rootCmd.AddCommand(rke2Cmd)

	// Add all the modularized command packages
	rke2Cmd.AddCommand(servercmd.Cmd)
	rke2Cmd.AddCommand(agentcmd.Cmd)
	rke2Cmd.AddCommand(configcmd.Cmd)
	rke2Cmd.AddCommand(systemcmd.Cmd)
	rke2Cmd.AddCommand(lbcmd.Cmd)
}
