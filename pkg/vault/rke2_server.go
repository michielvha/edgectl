/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Package vault provides specialized handlers for RKE2 cluster secrets management.

This file handles master node management for RKE2 clusters:
- StoreMasterInfo: Records metadata about master nodes, including their hostnames, IPs, and VIPs
- RetrieveMasterInfo: Gets the list of master nodes, their IPs, and associated VIP
- RetrieveFirstMasterIP: Gets the IP of the first (initial) master node for joining operations
- Helper functions: getFirstMasterIP, getHostIP

These functions enable multi-master high availability configurations by tracking
cluster node membership, determining network endpoints for control plane access,
and facilitating orderly cluster expansion and maintenance.
*/
package vault

import (
	"fmt"
	"net"
)

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
