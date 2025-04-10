/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init initializes the zerolog logger with proper configuration
// It should be called early in the application startup
func Init(verbose bool) {
	// Set up console writer with color and time formatting
	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	// Initialize global logger with timestamp
	log.Logger = zerolog.New(output).With().Timestamp().Logger()

	// By default, only show info level and above (info, warn, error, fatal)
	level := zerolog.InfoLevel

	// Check verbose flag
	if verbose {
		level = zerolog.DebugLevel
		log.Debug().Msg("Verbose logging enabled")
	}

	// Set the global log level
	zerolog.SetGlobalLevel(level)

	// Show a debug message that will only appear if verbose mode is on
	log.Debug().Msg("Logger initialized in verbose mode")
}

// Simple wrapper functions for zerolog
// These provide a more traditional logging API that's less verbose

// Debug logs a debug message
func Debug(msg string, args ...interface{}) {
	if len(args) == 0 {
		log.Debug().Msg(msg)
	} else {
		log.Debug().Msgf(msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...interface{}) {
	if len(args) == 0 {
		log.Info().Msg(msg)
	} else {
		log.Info().Msgf(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...interface{}) {
	if len(args) == 0 {
		log.Warn().Msg(msg)
	} else {
		log.Warn().Msgf(msg, args...)
	}
}

// Error logs an error with a message
func Error(msg string, err error, args ...interface{}) {
	if err != nil {
		if len(args) == 0 {
			log.Error().Err(err).Msg(msg)
		} else {
			log.Error().Err(err).Msgf(msg, args...)
		}
	} else {
		if len(args) == 0 {
			log.Error().Msg(msg)
		} else {
			log.Error().Msgf(msg, args...)
		}
	}
}

// Fatal logs a fatal error and exits
func Fatal(msg string, err error, args ...interface{}) {
	if err != nil {
		if len(args) == 0 {
			log.Fatal().Err(err).Msg(msg)
		} else {
			log.Fatal().Err(err).Msgf(msg, args...)
		}
	} else {
		if len(args) == 0 {
			log.Fatal().Msg(msg)
		} else {
			log.Fatal().Msgf(msg, args...)
		}
	}
}

// Get the underlying zerolog logger for advanced use cases
func GetLogger() zerolog.Logger {
	return log.Logger
}
