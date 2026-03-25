/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package agent

import (
	"fmt"
	"net"
	"os"

	"github.com/michielvha/edgectl/pkg/common"
	"github.com/michielvha/edgectl/pkg/logger"
	"github.com/michielvha/edgectl/pkg/vault"
)

// lookupHost is a package-level variable wrapping net.LookupHost so tests can inject a stub.
var lookupHost = net.LookupHost

// clusterIDDir is the directory where the cluster-id file is written.
// Tests can override this to use a temporary directory.
var clusterIDDir = "/etc/edgectl"

// Install sets up the K3s agent on the host.
// It fetches the join token from the secret store using the supplied clusterID.
// VIP resolution priority: secret store > --vip flag > --lb-hostname flag (DNS resolved).
func Install(store vault.SecretStore, clusterID, vip, lbHostname string) error {
	if _, err := FetchToken(store, clusterID); err != nil {
		return err
	}

	// Priority 1: fetch the VIP from Master Info in the secret store
	_, storedVIP, _, err := store.RetrieveMasterInfo("k3s", clusterID)
	if err == nil && storedVIP != "" {
		vip = storedVIP
		fmt.Printf("🔍 VIP fetched from secret store: %s\n", storedVIP)
	}

	// Priority 2: --vip flag is already set via the parameter

	// Priority 3: resolve --lb-hostname to an IP as fallback
	if vip == "" && lbHostname != "" {
		addrs, err := lookupHost(lbHostname)
		if err != nil || len(addrs) == 0 {
			return fmt.Errorf("failed to resolve load balancer hostname %s: %w", lbHostname, err)
		}
		vip = addrs[0]
		fmt.Printf("🔍 Resolved LB hostname %s to %s\n", lbHostname, vip)
	}

	installOptions := ""
	if vip != "" {
		installOptions = fmt.Sprintf("-l %s", vip)
		fmt.Printf("🌐 Using VIP %s for load balancer TLS SANs\n", vip)
	} else {
		logger.Debug("No VIP found via secret store, --vip, or --lb-hostname, using default settings")
	}
	// Run the installation script with options
	common.RunBashFunction("k3s.sh", fmt.Sprintf("install_k3s_agent %s", installOptions))

	return nil
}

// FetchToken fetches token from the secret store & sets as env variable
func FetchToken(store vault.SecretStore, clusterID string) (string, error) {
	token, err := store.RetrieveJoinToken("k3s", clusterID)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve join token: %w", err)
	}

	// ensure edgectl main directory exists
	_ = os.MkdirAll(clusterIDDir, 0o750)

	if err := os.WriteFile(clusterIDDir+"/cluster-id", []byte(clusterID), 0o600); err != nil {
		return "", fmt.Errorf("failed to write cluster-id: %w", err)
	}

	// Set token as environment variable for the bash script to use
	_ = os.Setenv("K3S_TOKEN", token)
	fmt.Println("✅ Set K3S_TOKEN environment variable")

	return token, nil
}
