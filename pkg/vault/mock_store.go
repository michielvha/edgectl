/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package vault

// MockStore is a hand-written mock implementing SecretStore.
// Each field is a function that, when set, overrides the default (zero-value) behavior.
// Tests set only the methods they care about; unset methods panic with a clear message.
type MockStore struct {
	StoreSecretFunc           func(fullVaultPath string, data map[string]interface{}) error
	RetrieveSecretFunc        func(fullVaultPath string) (map[string]interface{}, error)
	ListKeysFunc              func(fullVaultPath string) ([]string, error)
	DeleteSecretFunc          func(fullVaultPath string) error
	StoreJoinTokenFunc        func(clusterID, token string) error
	RetrieveJoinTokenFunc     func(clusterID string) (string, error)
	StoreMasterInfoFunc       func(clusterID, hostname string, hosts []string, vip string) error
	RetrieveMasterInfoFunc    func(clusterID string) ([]string, string, map[string]string, error)
	RetrieveFirstMasterIPFunc func(clusterID string) (string, error)
	StoreKubeConfigFunc       func(clusterID, kubeconfigPath, vip string) error
	RetrieveKubeConfigFunc    func(clusterID, destinationPath string) error
	StoreLBInfoFunc           func(clusterID, hostname, vip string, isMain bool) error
	RetrieveLBInfoFunc        func(clusterID string) ([]map[string]interface{}, string, error)
	RemoveLBNodeFunc          func(clusterID, hostname string) error
	DeleteClusterDataFunc     func(clusterID string) error
}

// Compile-time check: *MockStore must satisfy SecretStore.
var _ SecretStore = (*MockStore)(nil)

func (m *MockStore) StoreSecret(fullVaultPath string, data map[string]interface{}) error {
	if m.StoreSecretFunc != nil {
		return m.StoreSecretFunc(fullVaultPath, data)
	}
	panic("MockStore.StoreSecret not set")
}

func (m *MockStore) RetrieveSecret(fullVaultPath string) (map[string]interface{}, error) {
	if m.RetrieveSecretFunc != nil {
		return m.RetrieveSecretFunc(fullVaultPath)
	}
	panic("MockStore.RetrieveSecret not set")
}

func (m *MockStore) ListKeys(fullVaultPath string) ([]string, error) {
	if m.ListKeysFunc != nil {
		return m.ListKeysFunc(fullVaultPath)
	}
	panic("MockStore.ListKeys not set")
}

func (m *MockStore) DeleteSecret(fullVaultPath string) error {
	if m.DeleteSecretFunc != nil {
		return m.DeleteSecretFunc(fullVaultPath)
	}
	panic("MockStore.DeleteSecret not set")
}

func (m *MockStore) StoreJoinToken(clusterID, token string) error {
	if m.StoreJoinTokenFunc != nil {
		return m.StoreJoinTokenFunc(clusterID, token)
	}
	panic("MockStore.StoreJoinToken not set")
}

func (m *MockStore) RetrieveJoinToken(clusterID string) (string, error) {
	if m.RetrieveJoinTokenFunc != nil {
		return m.RetrieveJoinTokenFunc(clusterID)
	}
	panic("MockStore.RetrieveJoinToken not set")
}

func (m *MockStore) StoreMasterInfo(clusterID, hostname string, hosts []string, vip string) error {
	if m.StoreMasterInfoFunc != nil {
		return m.StoreMasterInfoFunc(clusterID, hostname, hosts, vip)
	}
	panic("MockStore.StoreMasterInfo not set")
}

func (m *MockStore) RetrieveMasterInfo(clusterID string) (hosts []string, vip string, hostIPs map[string]string, err error) {
	if m.RetrieveMasterInfoFunc != nil {
		return m.RetrieveMasterInfoFunc(clusterID)
	}
	panic("MockStore.RetrieveMasterInfo not set")
}

func (m *MockStore) RetrieveFirstMasterIP(clusterID string) (string, error) {
	if m.RetrieveFirstMasterIPFunc != nil {
		return m.RetrieveFirstMasterIPFunc(clusterID)
	}
	panic("MockStore.RetrieveFirstMasterIP not set")
}

func (m *MockStore) StoreKubeConfig(clusterID, kubeconfigPath, vip string) error {
	if m.StoreKubeConfigFunc != nil {
		return m.StoreKubeConfigFunc(clusterID, kubeconfigPath, vip)
	}
	panic("MockStore.StoreKubeConfig not set")
}

func (m *MockStore) RetrieveKubeConfig(clusterID, destinationPath string) error {
	if m.RetrieveKubeConfigFunc != nil {
		return m.RetrieveKubeConfigFunc(clusterID, destinationPath)
	}
	panic("MockStore.RetrieveKubeConfig not set")
}

func (m *MockStore) StoreLBInfo(clusterID, hostname, vip string, isMain bool) error {
	if m.StoreLBInfoFunc != nil {
		return m.StoreLBInfoFunc(clusterID, hostname, vip, isMain)
	}
	panic("MockStore.StoreLBInfo not set")
}

func (m *MockStore) RetrieveLBInfo(clusterID string) (nodes []map[string]interface{}, vip string, err error) {
	if m.RetrieveLBInfoFunc != nil {
		return m.RetrieveLBInfoFunc(clusterID)
	}
	panic("MockStore.RetrieveLBInfo not set")
}

func (m *MockStore) RemoveLBNode(clusterID, hostname string) error {
	if m.RemoveLBNodeFunc != nil {
		return m.RemoveLBNodeFunc(clusterID, hostname)
	}
	panic("MockStore.RemoveLBNode not set")
}

func (m *MockStore) DeleteClusterData(clusterID string) error {
	if m.DeleteClusterDataFunc != nil {
		return m.DeleteClusterDataFunc(clusterID)
	}
	panic("MockStore.DeleteClusterData not set")
}
