/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is now set in `main.go`
var Version string

// versionCmd represents the version command, which displays the CLI version
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Prints the current binary version",
	Long: `Displays the current version of the edge-cli tool.

This command is useful for verifying which version of edge-cli is installed.
The version is set dynamically during build time using GitVersion & GoReleaser.

Example usage:
  edgectl version
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("\n🔧 Client Version: %s\n", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// versionCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// versionCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
