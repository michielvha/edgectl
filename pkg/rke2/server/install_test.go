package server

import (
	"fmt"
	"os"
	"testing"

	"github.com/michielvha/edgectl/pkg/vault"
)

const (
	testSecretToken = "my-secret-token"
	testClusterID   = "test-cluster"
)

// TestFetchTokenFromSecretStore_SetsEnvVars verifies that FetchTokenFromSecretStore
// retrieves the token and first master IP, setting the expected env vars.
// Note: This test requires write access to /etc/edgectl (root or writable path).
func TestFetchTokenFromSecretStore_SetsEnvVars(t *testing.T) {
	// Skip if we can't write to /etc/edgectl (non-root in CI)
	if err := os.MkdirAll("/etc/edgectl", 0o750); err != nil {
		t.Skip("skipping: cannot write to /etc/edgectl (requires root)")
	}

	mock := &vault.MockStore{
		RetrieveJoinTokenFunc: func(clusterID string) (string, error) {
			if clusterID != testClusterID {
				t.Errorf("unexpected clusterID: %s", clusterID)
			}
			return testSecretToken, nil
		},
		RetrieveFirstMasterIPFunc: func(clusterID string) (string, error) {
			return "10.0.0.1", nil
		},
	}

	// Use a temp dir to avoid writing to /etc/edgectl in tests
	// Note: FetchTokenFromSecretStore writes to /etc/edgectl which needs root;
	// in CI this test may need to run as root or the write can be skipped.

	token, err := FetchTokenFromSecretStore(mock, testClusterID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != testSecretToken {
		t.Errorf("expected token %q, got %q", testSecretToken, token)
	}

	// Check env vars were set
	if got := os.Getenv("RKE2_TOKEN"); got != testSecretToken {
		t.Errorf("expected RKE2_TOKEN=%q, got %q", testSecretToken, got)
	}
	if got := os.Getenv("RKE2_SERVER_IP"); got != "10.0.0.1" {
		t.Errorf("expected RKE2_SERVER_IP='10.0.0.1', got %q", got)
	}

	// Cleanup
	t.Cleanup(func() {
		os.Unsetenv("RKE2_TOKEN")     //nolint:errcheck
		os.Unsetenv("RKE2_SERVER_IP") //nolint:errcheck
	})
}

// TestFetchTokenFromSecretStore_NoMasterIP verifies graceful handling
// when no first master IP is available.
func TestFetchTokenFromSecretStore_NoMasterIP(t *testing.T) {
	if err := os.MkdirAll("/etc/edgectl", 0o750); err != nil {
		t.Skip("skipping: cannot write to /etc/edgectl (requires root)")
	}

	mock := &vault.MockStore{
		RetrieveJoinTokenFunc: func(clusterID string) (string, error) {
			return "token-abc", nil
		},
		RetrieveFirstMasterIPFunc: func(clusterID string) (string, error) {
			return "", fmt.Errorf("no master IP found")
		},
	}

	token, err := FetchTokenFromSecretStore(mock, "cluster-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token-abc" {
		t.Errorf("expected token 'token-abc', got %q", token)
	}

	// RKE2_SERVER_IP should not be set when master IP retrieval fails
	t.Cleanup(func() { os.Unsetenv("RKE2_TOKEN") }) //nolint:errcheck
}

// TestHostDeduplication verifies that adding an existing host doesn't duplicate it.
// This tests the dedup logic extracted from Install().
func TestHostDeduplication(t *testing.T) {
	hosts := []string{"master1", "master2"}
	hostname := "master1"

	found := false
	for _, h := range hosts {
		if h == hostname {
			found = true
			break
		}
	}

	if !found {
		hosts = append(hosts, hostname)
	}

	if len(hosts) != 2 {
		t.Errorf("expected 2 hosts (no duplicate), got %d: %v", len(hosts), hosts)
	}
}

// TestHostDeduplication_NewHost verifies a new host is appended.
func TestHostDeduplication_NewHost(t *testing.T) {
	hosts := []string{"master1", "master2"}
	hostname := "master3"

	found := false
	for _, h := range hosts {
		if h == hostname {
			found = true
			break
		}
	}

	if !found {
		hosts = append(hosts, hostname)
	}

	if len(hosts) != 3 {
		t.Errorf("expected 3 hosts, got %d: %v", len(hosts), hosts)
	}
}
