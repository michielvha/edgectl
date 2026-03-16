/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package vault

// SecretStore defines the interface for all secret store operations.
// The existing *Client struct satisfies this interface implicitly.
// Consumers accept SecretStore to allow dependency injection and testing.
type SecretStore interface {
	// Generic CRUD
	StoreSecret(fullVaultPath string, data map[string]interface{}) error
	RetrieveSecret(fullVaultPath string) (map[string]interface{}, error)
	ListKeys(fullVaultPath string) ([]string, error)
	DeleteSecret(fullVaultPath string) error

	// RKE2 token management
	StoreJoinToken(clusterID, token string) error
	RetrieveJoinToken(clusterID string) (string, error)

	// RKE2 master/server management
	StoreMasterInfo(clusterID, hostname string, hosts []string, vip string) error
	RetrieveMasterInfo(clusterID string) ([]string, string, map[string]string, error)
	RetrieveFirstMasterIP(clusterID string) (string, error)

	// RKE2 kubeconfig management
	StoreKubeConfig(clusterID, kubeconfigPath string, vip string) error
	RetrieveKubeConfig(clusterID, destinationPath string) error

	// RKE2 load balancer management
	StoreLBInfo(clusterID, hostname, vip string, isMain bool) error
	RetrieveLBInfo(clusterID string) ([]map[string]interface{}, string, error)
	RemoveLBNode(clusterID, hostname string) error

	// RKE2 cluster management
	DeleteClusterData(clusterID string) error
}

// Compile-time check: *Client must satisfy SecretStore.
var _ SecretStore = (*Client)(nil)
