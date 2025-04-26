/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package system

import (
	"fmt"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/spf13/cobra"
)

// Cmd represents the "system" command
var Cmd = &cobra.Command{
	Use:   "system",
	Short: "Manage RKE2 system operations",
	Long: `The "system" command provides operations for RKE2 system management.
	
Examples:
  edgectl rke2 system status   # Check status of RKE2
  edgectl rke2 system purge    # Uninstall RKE2 from the host
`,
}

// statusCmd represents the "system status" command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of RKE2",
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("system status command executed")
		common.RunBashFunction("rke2.sh", "rke2_status")
	},
}

// purgeCmd represents the "system purge" command
var purgeCmd = &cobra.Command{
	Use:   "purge",
	Short: "Purge RKE2 install from host",
	Long:  `Completely removes RKE2 installation from the host.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Debug("system purge command executed")
		fmt.Println("🗑️  Purging RKE2 from the host...")
		common.RunBashFunction("rke2.sh", "purge_rke2")
		fmt.Println("✅ RKE2 purged successfully")
	},
}

// Initialize and register subcommands
func init() {
	// Register subcommands
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(purgeCmd)
}
