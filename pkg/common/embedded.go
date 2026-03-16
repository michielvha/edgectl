/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package common

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strings"
)

//go:embed scripts/*.sh
var embeddedScripts embed.FS

// ExtractEmbeddedScript extracts an embedded script to a user-namespaced temp directory.
// Each user gets their own /tmp/edgectl-{uid}/ directory to avoid permission conflicts.
func ExtractEmbeddedScript(scriptName string) string {
	uid := "unknown"
	if u, err := user.Current(); err == nil {
		uid = u.Uid
	}

	dir := filepath.Join("/tmp", "edgectl-"+uid)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		fmt.Printf("❌ Failed to create script dir: %v\n", err)
		os.Exit(1)
	}

	scriptPath := filepath.Join(dir, scriptName)

	// Read script from embedded FS
	data, err := embeddedScripts.ReadFile("scripts/" + scriptName)
	if err != nil {
		fmt.Printf("❌ Failed to read embedded script: %v\n", err)
		os.Exit(1)
	}

	// Write to a temp file
	if err := os.WriteFile(scriptPath, data, 0o700); err != nil {
		fmt.Printf("❌ Failed to write script: %v\n", err)
		os.Exit(1)
	}

	return scriptPath
}

// Runs a function from the sourced script
func RunBashFunction(scriptName, commandString string) {
	scriptPath := ExtractEmbeddedScript(scriptName)

	// Split the command into function name and arguments
	parts := strings.Fields(commandString)
	functionName := parts[0]
	args := []string{scriptPath, functionName}

	// Add any additional arguments if present
	if len(parts) > 1 {
		args = append(args, parts[1:]...)
	}

	// Run the full script and pass the function name and arguments
	cmd := exec.Command("bash", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Important to inherit input in case sudo or interactive steps exist

	// Pass the current environment, including updated vars from os.Setenv
	cmd.Env = os.Environ() // 👈 Important! Ensures it inherits updated env

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error executing %s from %s: %v\n", commandString, scriptPath, err)
		os.Exit(1)
	}
}
