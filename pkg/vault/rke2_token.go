/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu

Package vault provides specialized handlers for RKE2 cluster secrets management.

This file handles the token management functionality for RKE2 clusters:
- StoreJoinToken: Saves a cluster join token in Vault under a specific cluster ID
- RetrieveJoinToken: Retrieves the join token for a given cluster ID

These functions are critical for the RKE2 cluster bootstrapping process, allowing
servers and agents to securely join existing clusters without manual token handling.
*/
package vault

import (
	"fmt"
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
