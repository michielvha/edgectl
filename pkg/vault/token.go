/*
Copyright © 2025 VH & Co - contact@vhco.pro

Package vault provides specialized handlers for cluster secrets management.

This file handles the token management functionality for Kubernetes clusters:
- StoreJoinToken: Saves a cluster join token in the secret store under a specific cluster ID
- RetrieveJoinToken: Retrieves the join token for a given cluster ID

These functions are critical for the cluster bootstrapping process, allowing
servers and agents to securely join existing clusters without manual token handling.
*/
package vault

import (
	"fmt"
)

// StoreJoinToken saves a token under a specific cluster path
func (c *Client) StoreJoinToken(distro, clusterID, token string) error {
	return c.StoreSecret(fmt.Sprintf("kv/data/%s/%s/token", distro, clusterID), map[string]interface{}{
		"join_token": token,
		"cluster":    clusterID,
	})
}

// RetrieveJoinToken loads a join token using cluster ID
func (c *Client) RetrieveJoinToken(distro, clusterID string) (string, error) {
	data, err := c.RetrieveSecret(fmt.Sprintf("kv/data/%s/%s/token", distro, clusterID))
	if err != nil {
		return "", err
	}
	token, ok := data["join_token"].(string)
	if !ok {
		return "", fmt.Errorf("join_token not found for cluster %s", clusterID)
	}
	return token, nil
}
