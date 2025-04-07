// cmd/common/embedded.go
package common

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	vault "github.com/michielvha/edgectl/pkg/vault"
)

//go:embed scripts/*.sh
var embeddedScripts embed.FS

// Extracts an embedded script to /tmp
func ExtractEmbeddedScript(scriptName string) string {
	scriptPath := filepath.Join("/tmp", scriptName)

	// Read script from embedded FS
	data, err := embeddedScripts.ReadFile("scripts/" + scriptName)
	if err != nil {
		fmt.Printf("❌ Failed to read embedded script: %v\n", err)
		os.Exit(1)
	}

	// Write to a temp file
	if err := os.WriteFile(scriptPath, data, 0o755); err != nil {
		fmt.Printf("❌ Failed to write script: %v\n", err)
		os.Exit(1)
	}

	return scriptPath
}

// Runs a function from the sourced script
func RunBashFunction(scriptName, functionName string) {
	scriptPath := ExtractEmbeddedScript(scriptName)

	// Run the full script and pass the function name to call inside the script
	cmd := exec.Command("bash", scriptPath, functionName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin // Important to inherit input in case sudo or interactive steps exist
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Error executing %s from %s: %v\n", functionName, scriptPath, err)
		os.Exit(1)
	}
}

// Fetch token from Vault & set as env var / file
// TODO: check if we can rewrite this with viper package.
func FetchTokenFromVault(clusterID string) string {
	fmt.Println("🔐 Cluster ID supplied, retrieving join token from Vault...")

	vaultClient, err := vault.NewClient()
	if err != nil {
		fmt.Printf("❌ Failed to initialize Vault client: %v\n", err)
		os.Exit(1)
	}

	token, err := vaultClient.RetrieveJoinToken(clusterID)
	if err != nil {
		fmt.Printf("❌ Failed to retrieve join token from Vault: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✅ Retrieved token: %s\n", token)
	_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o644)
	_ = os.Setenv("RKE2_TOKEN", token)

	return token
}