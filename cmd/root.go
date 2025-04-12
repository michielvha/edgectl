/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "edgectl",
	Short: "A CLI tool for managing Edge Cloud infrastructure",
	Long: `
edgectl is a lightweight command-line interface for managing Edge Cloud resources.

It streamlines tasks such as provisioning clusters, joining nodes, managing configurations,
and interacting with secure secrets storage — all tailored for edge environments.

Whether you're deploying a new RKE2 cluster, automating node registration, or storing
kubeconfigs securely in Vault, edgectl helps you orchestrate your edge infrastructure with ease.
`,
	// This ensures the logger is set up before any command runs
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Initialize logger with verbose flag from viper (which combines cli flags, env vars, config file)
		logger.Init(viper.GetBool("verbose"))

		// Always log these messages at debug level to verify verbose mode
		logger.Debug("CLI execution started")

		// Log config file path if one was found
		if viper.ConfigFileUsed() != "" {
			logger.Debug("Config file found: %s", viper.ConfigFileUsed())
		} else {
			logger.Debug("No config file found, using defaults and environment variables")
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Initialize cobra
	cobra.OnInitialize(initConfig)

	// Define persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.edgectl.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging")

	// Bind flags to viper for config file and env var support
	if err := viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose")); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding verbose flag: %v\n", err)
	}

	// Also bind to environment variables
	if err := viper.BindEnv("verbose", "VERBOSE"); err != nil {
		fmt.Fprintf(os.Stderr, "Error binding verbose environment variable: %v\n", err)
	}

	// Cobra also supports local flags, which will only run when this action is called directly
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set (viper)
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			// Can't use logger yet as it's not initialized
			fmt.Fprintf(os.Stderr, "Error: could not find home directory: %v\n", err)
			os.Exit(1)
		}

		// Search config in home directory with name ".edgectl" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName(".edgectl")
		viper.SetConfigType("yaml")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in (silently fail if not found)
	if err := viper.ReadInConfig(); err == nil {
		// Log that we found and are using a config file
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
