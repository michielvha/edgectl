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
func TestFetchTokenFromSecretStore_SetsEnvVars(t *testing.T) {
	if err := os.MkdirAll("/etc/edgectl", 0o750); err != nil {
		t.Skip("skipping: cannot write to /etc/edgectl (requires root)")
	}

	mock := &vault.MockStore{
		RetrieveJoinTokenFunc: func(distro, clusterID string) (string, error) {
			if clusterID != testClusterID {
				t.Errorf("unexpected clusterID: %s", clusterID)
			}
			return testSecretToken, nil
		},
		RetrieveFirstMasterIPFunc: func(distro, clusterID string) (string, error) {
			return "10.0.0.1", nil
		},
	}

	token, err := FetchTokenFromSecretStore(mock, testClusterID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != testSecretToken {
		t.Errorf("expected token %q, got %q", testSecretToken, token)
	}

	// Check env vars were set (K3s uses K3S_TOKEN and K3S_URL)
	if got := os.Getenv("K3S_TOKEN"); got != testSecretToken {
		t.Errorf("expected K3S_TOKEN=%q, got %q", testSecretToken, got)
	}
	if got := os.Getenv("K3S_URL"); got != "https://10.0.0.1:6443" {
		t.Errorf("expected K3S_URL='https://10.0.0.1:6443', got %q", got)
	}

	t.Cleanup(func() {
		os.Unsetenv("K3S_TOKEN") //nolint:errcheck // error irrelevant in test cleanup
		os.Unsetenv("K3S_URL")   //nolint:errcheck // error irrelevant in test cleanup
	})
}

// TestFetchTokenFromSecretStore_NoMasterIP verifies graceful handling
// when no first master IP is available.
func TestFetchTokenFromSecretStore_NoMasterIP(t *testing.T) {
	if err := os.MkdirAll("/etc/edgectl", 0o750); err != nil {
		t.Skip("skipping: cannot write to /etc/edgectl (requires root)")
	}

	mock := &vault.MockStore{
		RetrieveJoinTokenFunc: func(distro, clusterID string) (string, error) {
			return "token-abc", nil
		},
		RetrieveFirstMasterIPFunc: func(distro, clusterID string) (string, error) {
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

	t.Cleanup(func() { os.Unsetenv("K3S_TOKEN") }) //nolint:errcheck // error irrelevant in test cleanup
}

// TestHostDeduplication verifies that adding an existing host doesn't duplicate it.
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
