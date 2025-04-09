/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package cmd

import (
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
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Configure zerolog based on verbose flag
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

		// Set up console writer with color and time formatting
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}

		// Create multi-writer if you want to also write to a file
		// file, _ := os.OpenFile("edgectl.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		// multi := zerolog.MultiLevelWriter(output, file)
		// log.Logger = zerolog.New(multi).With().Timestamp().Caller().Logger()

		// Just use console output for now
		log.Logger = zerolog.New(output).With().Timestamp().Logger()

		// Set global log level based on verbose flag
		if verbose {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log.Debug().Msg("Debug logging enabled")
		} else {
			zerolog.SetGlobalLevel(zerolog.InfoLevel)
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
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.edgectl.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output (debug logging)")

	// Bind the verbose flag to viper for global access
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// TODO: Move to seperate viper package ?
// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal().Err(err).Msg("Could not find home directory")
		}

		// Search config in home directory with name ".edgectl" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigName(".edgectl")
		viper.SetConfigType("yaml")
	}

	// Read in environment variables that match
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		log.Debug().Msgf("Using config file: %s", viper.ConfigFileUsed())
	}
}
