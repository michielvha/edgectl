/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package vault

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// StoreJoinToken saves a token under a specific cluster path
func (c *Client) StoreJoinToken(clusterID, token string) error {
	return c.StoreSecret(fmt.Sprintf("kv/data/rke2/%s/token", clusterID), map[string]interface{}{
		"join_token": token,
		"cluster":    clusterID,
	})
}

// RetrieveJoinToken loads a join token using cluster ID
func (c *Client) RetrieveJoinToken(clusterID string) (string, error) {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/token", clusterID))
	if err != nil {
		return "", err
	}
	token, ok := data["join_token"].(string)
	if !ok {
		return "", fmt.Errorf("join_token not found for cluster %s", clusterID)
	}
	return token, nil
}

// StoreKubeConfig reads the kubeconfig from the host, modifies it to use VIP if provided, and uploads it to Vault
func (c *Client) StoreKubeConfig(clusterID, kubeconfigPath string, vip string) error {
	kubeconfig, err := os.ReadFile(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to read kubeconfig from path '%s': %w", kubeconfigPath, err)
	}

	// Replace the localhost URL with the VIP in the kubeconfig if VIP is provided
	kubeconfigStr := string(kubeconfig)
	if vip != "" {
		kubeconfigStr = strings.ReplaceAll(
			kubeconfigStr,
			"server: https://127.0.0.1:6443",
			fmt.Sprintf("server: https://%s:6443", vip),
		)
		fmt.Printf("🔄 Updated kubeconfig to use VIP: %s\n", vip)
	}

	return c.StoreSecret(fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID), map[string]interface{}{
		"kubeconfig": kubeconfigStr,
	})
}

// RetrieveKubeConfig fetches the kubeconfig from Vault and saves it to the host
// TODO: save to a path that is users home directory instead of the default directory which is only accessible by root & this won't work if the host is not part of the rke2 setup
func (c *Client) RetrieveKubeConfig(clusterID, destinationPath string) error {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/kubeconfig", clusterID))
	if err != nil {
		return fmt.Errorf("failed to retrieve kubeconfig for cluster %s: %w", clusterID, err)
	}

	kubeconfig, ok := data["kubeconfig"].(string)
	if !ok {
		return fmt.Errorf("kubeconfig not found or invalid type for cluster %s", clusterID)
	}

	// Create directory structure if it doesn't exist
	dir := filepath.Dir(destinationPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	err = os.WriteFile(destinationPath, []byte(kubeconfig), 0o600)
	if err != nil {
		return fmt.Errorf("failed to write kubeconfig to path '%s': %w", destinationPath, err)
	}

	return nil
}

// StoreLBInfo stores information about a load balancer node
func (c *Client) StoreLBInfo(clusterID, hostname, vip string, isMain bool) error {
	path := fmt.Sprintf("kv/data/rke2/%s/lb/%s", clusterID, hostname)
	return c.StoreSecret(path, map[string]interface{}{
		"hostname": hostname,
		"vip":      vip,
		"is_main":  isMain,
	})
}

// RetrieveLBInfo retrieves information about load balancer nodes
func (c *Client) RetrieveLBInfo(clusterID string) ([]map[string]interface{}, string, error) {
	// List all LB entries for this cluster
	path := fmt.Sprintf("kv/metadata/rke2/%s/lb", clusterID)
	keys, err := c.ListKeys(path)
	if err != nil {
		return nil, "", fmt.Errorf("failed to list load balancers for cluster %s: %w", clusterID, err)
	}

	lbNodes := []map[string]interface{}{}
	var vip string

	// Retrieve details for each LB
	for _, key := range keys {
		data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/lb/%s", clusterID, key))
		if err != nil {
			continue
		}

		lbNodes = append(lbNodes, data)

		// Get VIP from any LB (they should all have the same VIP)
		if vip == "" {
			vip, _ = data["vip"].(string)
		}

		// If this is the main LB, use its VIP as the definitive one
		isMain, ok := data["is_main"].(bool)
		if ok && isMain && data["vip"] != nil {
			vip = data["vip"].(string)
		}
	}

	if len(lbNodes) == 0 {
		return nil, "", fmt.Errorf("no load balancers found for cluster %s", clusterID)
	}

	return lbNodes, vip, nil
}

// StoreMasterInfo stores information about RKE2 master nodes and their configuration
func (c *Client) StoreMasterInfo(clusterID, hostname string, hosts []string, vip string) error {
	// Get the IP address of this host
	ipAddr, err := getHostIP(hostname)
	if err != nil {
		// If we can't get the IP, just use hostname as fallback
		ipAddr = hostname
	}

	// Check if we already have master IPs stored
	var hostIPs map[string]string
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/masters", clusterID))
	if err == nil && data["host_ips"] != nil {
		// Try to retrieve existing host_ips map
		if ipsData, ok := data["host_ips"].(map[string]interface{}); ok {
			hostIPs = make(map[string]string)
			for k, v := range ipsData {
				if strVal, ok := v.(string); ok {
					hostIPs[k] = strVal
				}
			}
		}
	}

	// Initialize map if it doesn't exist yet
	if hostIPs == nil {
		hostIPs = make(map[string]string)
	}

	// Add/update this host's IP
	hostIPs[hostname] = ipAddr

	path := fmt.Sprintf("kv/data/rke2/%s/masters", clusterID)
	return c.StoreSecret(path, map[string]interface{}{
		"hosts":      hosts,
		"vip":        vip,
		"last_added": hostname,
		"host_ips":   hostIPs,
		"first_ip":   getFirstMasterIP(hosts, hostIPs, ipAddr),
	})
}

// Helper function to get the IP of the first master in the list
func getFirstMasterIP(hosts []string, hostIPs map[string]string, currentIP string) string {
	// If we have no hosts, return the current IP
	if len(hosts) == 0 {
		return currentIP
	}

	// Get the first host in the list
	firstHost := hosts[0]

	// If we have an IP for this host, return it
	if ip, ok := hostIPs[firstHost]; ok {
		return ip
	}

	// Otherwise return hostname as fallback
	return firstHost
}

// Helper function to get the IP address of a hostname
func getHostIP(hostname string) (string, error) {
	addrs, err := net.LookupHost(hostname)
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", fmt.Errorf("no IP addresses found for host: %s", hostname)
	}
	return addrs[0], nil
}

// RetrieveMasterInfo retrieves RKE2 master nodes information
func (c *Client) RetrieveMasterInfo(clusterID string) ([]string, string, map[string]string, error) {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/masters", clusterID))
	if err != nil {
		return nil, "", nil, err
	}

	hostsRaw, ok := data["hosts"]
	if !ok {
		return nil, "", nil, fmt.Errorf("hosts information not found for cluster %s", clusterID)
	}

	// Convert interface{} to string slice
	hosts := []string{}
	if hostsArray, ok := hostsRaw.([]interface{}); ok {
		for _, h := range hostsArray {
			if hostStr, ok := h.(string); ok {
				hosts = append(hosts, hostStr)
			}
		}
	}

	vip, _ := data["vip"].(string)

	// Extract host_ips map if available
	hostIPs := make(map[string]string)
	if hostIPsRaw, ok := data["host_ips"].(map[string]interface{}); ok {
		for hostname, ipRaw := range hostIPsRaw {
			if ip, ok := ipRaw.(string); ok {
				hostIPs[hostname] = ip
			}
		}
	}

	return hosts, vip, hostIPs, nil
}

// RetrieveFirstMasterIP retrieves the IP address of the first master node in the cluster
func (c *Client) RetrieveFirstMasterIP(clusterID string) (string, error) {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/rke2/%s/masters", clusterID))
	if err != nil {
		return "", fmt.Errorf("failed to retrieve master info: %w", err)
	}

	// Try to get the explicitly stored first_ip
	if firstIP, ok := data["first_ip"].(string); ok && firstIP != "" {
		return firstIP, nil
	}

	// Fallback: Try to get the first host's IP from host_ips map
	if hosts, ok := data["hosts"].([]interface{}); ok && len(hosts) > 0 {
		if firstHost, ok := hosts[0].(string); ok {
			if hostIPs, ok := data["host_ips"].(map[string]interface{}); ok {
				if ip, ok := hostIPs[firstHost].(string); ok {
					return ip, nil
				}
			}
			// If we have a hostname but no IP, return the hostname as fallback
			return firstHost, nil
		}
	}

	return "", fmt.Errorf("no master IP information found for cluster %s", clusterID)
}
