/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/vault"
)

// Install sets up the RKE2 server on the host.
// If `isExisting` is true, it pulls the token from Vault using the supplied clusterID.
// Otherwise, it generates a new clusterID and saves token + kubeconfig to Vault.
func Install(clusterID string, isExisting bool) error {
	vaultClient, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	// If the cluster ID was provided (existing cluster), fetch the join token
	if isExisting {
		if _, err := FetchTokenFromVault(clusterID); err != nil {
			return err
		}
	} else {
		// Generate a new cluster ID
		clusterID = fmt.Sprintf("rke2-%s", uuid.New().String()[:8])
		_ = os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o644)
		fmt.Printf("🆔 Generated cluster ID: %s\n", clusterID)
	}

	// Run the installation script
	common.RunBashFunction("rke2.sh", "install_rke2_server")

	// If this is a new cluster, store token and kubeconfig in Vault
	if !isExisting {
		tokenBytes, err := os.ReadFile("/var/lib/rancher/rke2/server/node-token")
		if err != nil {
			return fmt.Errorf("failed to read generated node token: %w", err)
		}

		token := strings.TrimSpace(string(tokenBytes))
		if err := vaultClient.StoreJoinToken(clusterID, token); err != nil {
			return fmt.Errorf("failed to store token in Vault: %w", err)
		}
		fmt.Printf("🔐 Token successfully stored in Vault for cluster %s\n", clusterID)

		kubeconfigPath := "/etc/rancher/rke2/rke2.yaml"
		if _, statErr := os.Stat(kubeconfigPath); os.IsNotExist(statErr) {
			return fmt.Errorf("kubeconfig file not found at path: %s", kubeconfigPath)
		}

		err = vaultClient.StoreKubeConfig(clusterID, kubeconfigPath)
		if err != nil {
			return fmt.Errorf("failed to store kubeconfig in Vault: %w", err)
		}
		fmt.Printf("🔐 Kubeconfig successfully stored in Vault for cluster %s\n", clusterID)
	}

	return nil
}

// Fetch token from Vault & set as env var / file
// TODO: check if we can rewrite this with viper package.
func FetchTokenFromVault(clusterID string) (string, error) {
	vaultClient, err := vault.NewClient()
	if err != nil {
		return "", fmt.Errorf("failed to initialize Vault client: %w", err)
	}

	token, err := vaultClient.RetrieveJoinToken(clusterID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve join token: %w", err)
	}

	// ensure edgectl main directory exists
	_ = os.MkdirAll("/etc/edgectl", 0o755)

	if err := os.WriteFile("/etc/edgectl/cluster-id", []byte(clusterID), 0o644); err != nil {
		return "", fmt.Errorf("failed to write cluster-id: %w", err)
	}

	// This sets it only for the current Go process
	_ = os.Setenv("RKE2_TOKEN", token)

	// ✅ Write token to config.yaml
	rke2ConfigPath := "/etc/rancher/rke2/config.yaml"
	appendLine := fmt.Sprintf("token: \"%s\"\n", token)

	// Ensure the parent directory exists
	if err := os.MkdirAll("/etc/rancher/rke2", 0o755); err != nil {
		return "", fmt.Errorf("failed to create RKE2 config directory: %w", err)
	}
	fmt.Println("📁 Ensured /etc/rancher/rke2 exists")

	fmt.Printf("📄 Attempting to write token to config at %s\n", rke2ConfigPath)
	f, err := os.OpenFile(rke2ConfigPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error writing token to config: %v\n", err)
		return "", fmt.Errorf("failed to open rke2 config for writing token: %w", err)
	}
	fmt.Println("✅ Opened config file successfully")
	defer func() {
		if cerr := f.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "failed to close rke2 config file: %v\n", cerr)
		}
	}()

	if _, err := f.WriteString(appendLine); err != nil {
		fmt.Fprintf(os.Stderr, "❌ Error writing token to config: %v\n", err)
		return "", fmt.Errorf("failed to append token to rke2 config: %w", err)
	}
	fmt.Println("✅ Appended token to rke2 config")
	return token, nil
}
