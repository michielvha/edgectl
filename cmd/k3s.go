/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package cmd

import (
	"github.com/spf13/cobra"

	k3sagentcmd "github.com/michielvha/edgectl/cmd/k3s/agent"
	k3slbcmd "github.com/michielvha/edgectl/cmd/k3s/lb"
	k3sservercmd "github.com/michielvha/edgectl/cmd/k3s/server"
	k3ssystemcmd "github.com/michielvha/edgectl/cmd/k3s/system"
)

// k3sCmd represents the "k3s" command
var k3sCmd = &cobra.Command{
	Use:   "k3s",
	Short: "Manage K3s cluster",
	Long: `The "k3s" command allows you to install, manage, and uninstall K3s components.

Examples:
  edgectl k3s server install        # Install K3s Server
  edgectl k3s agent install         # Install K3s Agent
  edgectl k3s lb create             # Create a load balancer for K3s
  edgectl k3s system purge          # Uninstall K3s
  edgectl k3s system kubeconfig     # Fetch kubeconfig from secret store
  edgectl k3s system bash           # Configure bash environment
`,
}

// Register subcommands
func init() {
	rootCmd.AddCommand(k3sCmd)

	k3sCmd.AddCommand(k3sservercmd.Cmd)
	k3sCmd.AddCommand(k3sagentcmd.Cmd)
	k3sCmd.AddCommand(k3ssystemcmd.Cmd)
	k3sCmd.AddCommand(k3slbcmd.Cmd)
}
