/*
Copyright © 2025 VH & Co - contact@vhco.pro

Package vault provides specialized handlers for cluster secrets management.

This file handles cluster-level operations:
- DeleteClusterData: Removes all secret store data for a given cluster (token, kubeconfig, masters, LB entries)
*/
package vault

import (
	"fmt"

	"github.com/michielvha/edgectl/pkg/logger"
)

// DeleteClusterData permanently removes all secret store data for a cluster.
// Uses kv/metadata/ prefix for permanent deletion of all KV v2 versions.
// Errors are logged as warnings and do not stop the cleanup — best-effort deletion.
func (c *Client) DeleteClusterData(distro, clusterID string) error {
	basePath := fmt.Sprintf("kv/metadata/%s/%s", distro, clusterID)
	var lastErr error

	// Delete known fixed paths
	for _, subpath := range []string{"token", "kubeconfig", "masters"} {
		path := fmt.Sprintf("%s/%s", basePath, subpath)
		if err := c.DeleteSecret(path); err != nil {
			logger.Warn("Failed to delete %s: %v", path, err)
			lastErr = err
		} else {
			logger.Debug("Deleted %s", path)
		}
	}

	// Delete all LB entries (list then delete each)
	lbPath := fmt.Sprintf("%s/lb", basePath)
	keys, err := c.ListKeys(lbPath)
	if err == nil {
		for _, key := range keys {
			path := fmt.Sprintf("%s/%s", lbPath, key)
			if err := c.DeleteSecret(path); err != nil {
				logger.Warn("Failed to delete LB entry %s: %v", path, err)
				lastErr = err
			} else {
				logger.Debug("Deleted LB entry %s", path)
			}
		}
		// Delete the LB directory itself
		if err := c.DeleteSecret(lbPath); err != nil {
			logger.Warn("Failed to delete LB path %s: %v", lbPath, err)
		}
	}

	// Delete the cluster root
	if err := c.DeleteSecret(basePath); err != nil {
		logger.Warn("Failed to delete cluster root %s: %v", basePath, err)
	}

	if lastErr != nil {
		return fmt.Errorf("some cluster data could not be deleted (see warnings above)")
	}
	return nil
}
