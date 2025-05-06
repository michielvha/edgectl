/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Package vault provides specialized handlers for RKE2 cluster secrets management.

This file handles the load balancer configuration for RKE2 clusters:
- StoreLBInfo: Stores information about a load balancer node, including its hostname, VIP, and primary/backup status
- RetrieveLBInfo: Gets information about all load balancer nodes for a cluster
- RemoveLBNode: Removes a load balancer node from the Vault storage when it's decommissioned

These functions enable high-availability load balancer configuration with primary/backup
relationships between nodes, tracking of virtual IPs (VIPs), and automatic failover capability.
*/
package vault

import (
	"fmt"
	"strings"
)

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
		// Return an empty list instead of an error when no LBs exist yet
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return []map[string]interface{}{}, "", nil
		}
		return nil, "", fmt.Errorf("failed to list load balancers for cluster %s: %w", clusterID, err)
	}

	// If keys list is empty, we have no load balancers yet
	if len(keys) == 0 {
		return []map[string]interface{}{}, "", nil
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

	// Return empty list instead of error when no load balancer nodes are found
	if len(lbNodes) == 0 {
		return []map[string]interface{}{}, "", nil
	}

	return lbNodes, vip, nil
}

// RemoveLBNode removes a load balancer node from the Vault storage
func (c *Client) RemoveLBNode(clusterID, hostname string) error {
	// Delete the LB node entry
	path := fmt.Sprintf("kv/metadata/rke2/%s/lb/%s", clusterID, hostname)
	if err := c.DeleteSecret(path); err != nil {
		// If the entry doesn't exist, don't return an error
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "not found") {
			return nil
		}
		return fmt.Errorf("failed to delete load balancer node %s for cluster %s: %w", hostname, clusterID, err)
	}
	return nil
}
