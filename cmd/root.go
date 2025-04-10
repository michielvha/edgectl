/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize logger
		initLogger()
		return nil
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

// initLogger configures zerolog based on verbose flag
func initLogger() {
	// Set up console writer with color and time formatting
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Initialize logger with timestamp
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// By default, only show info level and above (info, warn, error, fatal)
	level := zerolog.InfoLevel

	// Check verbose flag (from flag or config file or env var)
	if viper.GetBool("verbose") {
		level = zerolog.DebugLevel
		log.Debug().Msg("Verbose logging enabled")
	}

	// Set the global log level
	zerolog.SetGlobalLevel(level)
}

func init() {
	// Initialize cobra
	cobra.OnInitialize(initConfig)

	// Define persistent flags (available to all commands)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.edgectl.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output for debugging")

	// Bind flags to viper for config file and env var support
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Also bind to environment variables
	viper.BindEnv("verbose", "VERBOSE")

	// Cobra also supports local flags, which will only run when this action is called directly
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set
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
		// Only show in verbose mode
		log.Debug().Msgf("Using config file: %s", viper.ConfigFileUsed())
	}
}
