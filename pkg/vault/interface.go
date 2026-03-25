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

	// Cluster token management
	StoreJoinToken(distro, clusterID, token string) error
	RetrieveJoinToken(distro, clusterID string) (string, error)

	// Cluster master/server management
	StoreMasterInfo(distro, clusterID, hostname string, hosts []string, vip string) error
	RetrieveMasterInfo(distro, clusterID string) (hosts []string, vip string, hostIPs map[string]string, err error)
	RetrieveFirstMasterIP(distro, clusterID string) (string, error)

	// Cluster kubeconfig management
	StoreKubeConfig(distro, clusterID, kubeconfigPath, vip string) error
	RetrieveKubeConfig(distro, clusterID, destinationPath string) error

	// Cluster load balancer management
	StoreLBInfo(distro, clusterID, hostname, vip string, isMain bool) error
	RetrieveLBInfo(distro, clusterID string) (nodes []map[string]interface{}, vip string, err error)
	RemoveLBNode(distro, clusterID, hostname string) error

	// Cluster management
	DeleteClusterData(distro, clusterID string) error
}

// Compile-time check: *Client must satisfy SecretStore.
var _ SecretStore = (*Client)(nil)
