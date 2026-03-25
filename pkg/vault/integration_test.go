//go:build integration

package vault

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	openbaoImage    = "openbao/openbao:2.3.1"
	openbaoDevToken = "root"
)

// Shared container address for all integration tests.
var integrationAddr string

// TestMain starts a single OpenBao container for all integration tests,
// runs the tests, then cleans up.
func TestMain(m *testing.M) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        openbaoImage,
		ExposedPorts: []string{"8200/tcp"},
		Cmd:          []string{"server", "-dev", "-dev-root-token-id=" + openbaoDevToken, "-dev-listen-address=0.0.0.0:8200"},
		Env: map[string]string{
			"BAO_ADDR":              "http://0.0.0.0:8200",
			"SKIP_SETCAP":           "true",
			"BAO_DEV_ROOT_TOKEN_ID": openbaoDevToken,
		},
		WaitingFor: wait.ForHTTP("/v1/sys/seal-status").
			WithPort("8200/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to start OpenBao container: %v\n", err)
		os.Exit(1)
	}

	host, err := container.Host(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get container host: %v\n", err)
		os.Exit(1)
	}

	mappedPort, err := container.MappedPort(ctx, "8200")
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get mapped port: %v\n", err)
		os.Exit(1)
	}

	integrationAddr = fmt.Sprintf("http://%s:%s", host, mappedPort.Port())

	// Run all tests
	code := m.Run()

	// Cleanup
	if err := container.Terminate(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to terminate container: %v\n", err)
	}

	os.Exit(code)
}

// newTestClient creates a vault Client connected to the shared dev OpenBao instance.
// Each test gets its own KV mount via a unique prefix to avoid cross-test contamination.
func newTestClient(t *testing.T) *Client {
	t.Helper()

	// Set environment for the vault SDK
	t.Setenv("VAULT_ADDR", integrationAddr)
	t.Setenv("BAO_TOKEN", openbaoDevToken)

	client, err := NewClient()
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}

	// Enable KV v2 secrets engine at kv/ (dev mode may only have secret/)
	_, err = client.VaultClient.Logical().Write("sys/mounts/kv", map[string]interface{}{
		"type": "kv",
		"options": map[string]interface{}{
			"version": "2",
		},
	})
	if err != nil {
		// May already exist, that's fine — suppress
		_ = err
	}

	return client
}

// --- Generic CRUD ---

func TestIntegration_GenericCRUD(t *testing.T) {
	client := newTestClient(t)

	path := "kv/data/test/crud"

	// Store
	err := client.StoreSecret(path, map[string]interface{}{
		"username": "admin",
		"password": "s3cret",
	})
	if err != nil {
		t.Fatalf("StoreSecret failed: %v", err)
	}

	// Retrieve
	data, err := client.RetrieveSecret(path)
	if err != nil {
		t.Fatalf("RetrieveSecret failed: %v", err)
	}
	if data["username"] != "admin" {
		t.Errorf("expected username=admin, got %v", data["username"])
	}
	if data["password"] != "s3cret" {
		t.Errorf("expected password=s3cret, got %v", data["password"])
	}

	// List keys
	keys, err := client.ListKeys("kv/metadata/test")
	if err != nil {
		t.Fatalf("ListKeys failed: %v", err)
	}
	if len(keys) == 0 {
		t.Error("expected at least 1 key under kv/metadata/test")
	}

	// Delete
	err = client.DeleteSecret("kv/metadata/test/crud")
	if err != nil {
		t.Fatalf("DeleteSecret failed: %v", err)
	}

	// Verify deleted
	_, err = client.RetrieveSecret(path)
	if err == nil {
		t.Error("expected error after deletion, got nil")
	}
}

// --- Token round-trip ---

func TestIntegration_TokenStoreRetrieve(t *testing.T) {
	client := newTestClient(t)

	clusterID := "integration-test-cluster"
	token := "K10abc123def456::server:xyz789"

	err := client.StoreJoinToken("rke2", clusterID, token)
	if err != nil {
		t.Fatalf("StoreJoinToken failed: %v", err)
	}

	got, err := client.RetrieveJoinToken("rke2", clusterID)
	if err != nil {
		t.Fatalf("RetrieveJoinToken failed: %v", err)
	}
	if got != token {
		t.Errorf("expected token %q, got %q", token, got)
	}
}

// --- Master info accumulation ---

func TestIntegration_MasterInfoAccumulation(t *testing.T) {
	client := newTestClient(t)

	clusterID := "master-test-cluster"

	// First master
	err := client.StoreMasterInfo("rke2", clusterID, "master1", []string{"master1"}, "10.0.0.100")
	if err != nil {
		t.Fatalf("StoreMasterInfo (1st) failed: %v", err)
	}

	hosts, vip, hostIPs, err := client.RetrieveMasterInfo("rke2", clusterID)
	if err != nil {
		t.Fatalf("RetrieveMasterInfo (1st) failed: %v", err)
	}
	if len(hosts) != 1 || hosts[0] != "master1" {
		t.Errorf("expected [master1], got %v", hosts)
	}
	if vip != "10.0.0.100" {
		t.Errorf("expected VIP 10.0.0.100, got %q", vip)
	}

	// Second master
	err = client.StoreMasterInfo("rke2", clusterID, "master2", []string{"master1", "master2"}, "10.0.0.100")
	if err != nil {
		t.Fatalf("StoreMasterInfo (2nd) failed: %v", err)
	}

	hosts, _, hostIPs, err = client.RetrieveMasterInfo("rke2", clusterID)
	if err != nil {
		t.Fatalf("RetrieveMasterInfo (2nd) failed: %v", err)
	}
	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts, got %d: %v", len(hosts), hosts)
	}
	// Both masters should have IPs in the map
	if len(hostIPs) < 1 {
		t.Errorf("expected at least 1 entry in hostIPs, got %d", len(hostIPs))
	}

	// First master IP should be retrievable
	firstIP, err := client.RetrieveFirstMasterIP("rke2", clusterID)
	if err != nil {
		t.Fatalf("RetrieveFirstMasterIP failed: %v", err)
	}
	if firstIP == "" {
		t.Error("expected non-empty first master IP")
	}
}

// --- Kubeconfig VIP replacement ---

func TestIntegration_KubeconfigVIPReplacement(t *testing.T) {
	client := newTestClient(t)

	clusterID := "kubeconfig-test"

	// Create a fake kubeconfig with localhost
	tmpFile, err := os.CreateTemp(t.TempDir(), "kubeconfig-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	fakeKubeconfig := `apiVersion: v1
clusters:
- cluster:
    server: https://127.0.0.1:6443
  name: default
`
	if _, err := tmpFile.WriteString(fakeKubeconfig); err != nil {
		t.Fatalf("failed to write temp kubeconfig: %v", err)
	}
	tmpFile.Close()

	// Store with VIP replacement
	err = client.StoreKubeConfig("rke2", clusterID, tmpFile.Name(), "10.0.0.100")
	if err != nil {
		t.Fatalf("StoreKubeConfig failed: %v", err)
	}

	// Retrieve to a new file
	outPath := fmt.Sprintf("%s/kubeconfig-out.yaml", t.TempDir())
	err = client.RetrieveKubeConfig("rke2", clusterID, outPath)
	if err != nil {
		t.Fatalf("RetrieveKubeConfig failed: %v", err)
	}

	content, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output kubeconfig: %v", err)
	}

	if got := string(content); got == "" {
		t.Error("retrieved kubeconfig is empty")
	}

	// Verify VIP was substituted
	contentStr := string(content)
	if !strings.Contains(contentStr, "10.0.0.100") {
		t.Error("expected kubeconfig to contain VIP 10.0.0.100")
	}
	if strings.Contains(contentStr, "127.0.0.1") {
		t.Error("expected kubeconfig NOT to contain 127.0.0.1 after VIP replacement")
	}
}

// --- LB info with main/backup ---

func TestIntegration_LBInfoMainBackup(t *testing.T) {
	client := newTestClient(t)

	clusterID := "lb-test-cluster"

	// Store main LB
	err := client.StoreLBInfo("rke2", clusterID, "lb-main", "10.0.0.200", true)
	if err != nil {
		t.Fatalf("StoreLBInfo (main) failed: %v", err)
	}

	// Store backup LB
	err = client.StoreLBInfo("rke2", clusterID, "lb-backup", "10.0.0.200", false)
	if err != nil {
		t.Fatalf("StoreLBInfo (backup) failed: %v", err)
	}

	nodes, vip, err := client.RetrieveLBInfo("rke2", clusterID)
	if err != nil {
		t.Fatalf("RetrieveLBInfo failed: %v", err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 LB nodes, got %d", len(nodes))
	}
	if vip != "10.0.0.200" {
		t.Errorf("expected VIP 10.0.0.200, got %q", vip)
	}

	// Verify main/backup roles
	var hasMain, hasBackup bool
	for _, node := range nodes {
		if isMain, ok := node["is_main"].(bool); ok && isMain {
			hasMain = true
		} else {
			hasBackup = true
		}
	}
	if !hasMain {
		t.Error("expected at least one main LB node")
	}
	if !hasBackup {
		t.Error("expected at least one backup LB node")
	}

	// Remove a node
	err = client.RemoveLBNode("rke2", clusterID, "lb-backup")
	if err != nil {
		t.Fatalf("RemoveLBNode failed: %v", err)
	}

	nodes, _, err = client.RetrieveLBInfo("rke2", clusterID)
	if err != nil {
		t.Fatalf("RetrieveLBInfo after removal failed: %v", err)
	}
	if len(nodes) != 1 {
		t.Errorf("expected 1 LB node after removal, got %d", len(nodes))
	}
}

// --- DeleteClusterData full cleanup ---

func TestIntegration_DeleteClusterData(t *testing.T) {
	client := newTestClient(t)

	clusterID := "cleanup-test-cluster"

	// Store all types of data
	_ = client.StoreJoinToken("rke2", clusterID, "test-token")
	_ = client.StoreMasterInfo("rke2", clusterID, "master1", []string{"master1"}, "10.0.0.1")
	_ = client.StoreLBInfo("rke2", clusterID, "lb1", "10.0.0.1", true)

	// Create temp kubeconfig
	tmpFile, err := os.CreateTemp(t.TempDir(), "kubeconfig-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.WriteString("apiVersion: v1\nclusters: []\n")
	tmpFile.Close()
	_ = client.StoreKubeConfig("rke2", clusterID, tmpFile.Name(), "")

	// Verify data exists
	_, err = client.RetrieveJoinToken("rke2", clusterID)
	if err != nil {
		t.Fatalf("expected token to exist before cleanup: %v", err)
	}

	// Delete all
	err = client.DeleteClusterData("rke2", clusterID)
	if err != nil {
		t.Fatalf("DeleteClusterData failed: %v", err)
	}

	// Verify everything is gone
	_, err = client.RetrieveJoinToken("rke2", clusterID)
	if err == nil {
		t.Error("expected token retrieval to fail after cleanup")
	}

	_, _, _, err = client.RetrieveMasterInfo("rke2", clusterID)
	if err == nil {
		t.Error("expected master info retrieval to fail after cleanup")
	}
}
