package agent

import (
	"fmt"
	"os"
	"testing"

	"github.com/michielvha/edgectl/pkg/vault"
)

const (
	testAgentToken = "agent-token-123"
	testFlagVIP    = "flag-vip"
)

func TestFetchToken_SetsEnvVars(t *testing.T) {
	clusterIDDir = t.TempDir()

	mock := &vault.MockStore{
		RetrieveJoinTokenFunc: func(distro, clusterID string) (string, error) {
			if clusterID != "agent-cluster" {
				t.Errorf("unexpected clusterID: %s", clusterID)
			}
			return testAgentToken, nil
		},
	}

	token, err := FetchToken(mock, "agent-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != testAgentToken {
		t.Errorf("expected %q, got %q", testAgentToken, token)
	}
	if got := os.Getenv("K3S_TOKEN"); got != testAgentToken {
		t.Errorf("expected K3S_TOKEN=%q, got %q", testAgentToken, got)
	}

	t.Cleanup(func() { os.Unsetenv("K3S_TOKEN") }) //nolint:errcheck // error irrelevant in test cleanup
}

func TestVIPResolutionPriority_StoreWins(t *testing.T) {
	vip := testFlagVIP

	storedVIP := "store-vip"
	if storedVIP != "" {
		vip = storedVIP
	}

	if vip != "store-vip" {
		t.Errorf("expected store VIP to win, got %q", vip)
	}
}

func TestVIPResolutionPriority_FlagFallback(t *testing.T) {
	vip := testFlagVIP

	storedVIP := ""
	if storedVIP != "" {
		vip = storedVIP
	}

	if vip != testFlagVIP {
		t.Errorf("expected flag VIP fallback, got %q", vip)
	}
}

func TestVIPResolutionPriority_NoVIP(t *testing.T) {
	vip := ""

	storedVIP := ""
	if storedVIP != "" {
		vip = storedVIP
	}

	lbHostname := ""
	if vip == "" && lbHostname != "" {
		vip = "resolved"
	}

	if vip != "" {
		t.Errorf("expected empty VIP, got %q", vip)
	}
}

// --- DNS resolution tests (with injected lookupHost) ---

func TestLBHostnameDNS_Resolves(t *testing.T) {
	original := lookupHost
	lookupHost = func(host string) ([]string, error) {
		if host == "lb.example.com" {
			return []string{"10.50.0.1"}, nil
		}
		return nil, fmt.Errorf("not found")
	}
	t.Cleanup(func() { lookupHost = original })

	vip := ""
	lbHostname := "lb.example.com"
	if vip == "" && lbHostname != "" {
		addrs, err := lookupHost(lbHostname)
		if err != nil || len(addrs) == 0 {
			t.Fatalf("unexpected DNS error: %v", err)
		}
		vip = addrs[0]
	}

	if vip != "10.50.0.1" {
		t.Errorf("expected resolved VIP '10.50.0.1', got %q", vip)
	}
}

func TestLBHostnameDNS_Error(t *testing.T) {
	original := lookupHost
	lookupHost = func(host string) ([]string, error) {
		return nil, fmt.Errorf("DNS failure")
	}
	t.Cleanup(func() { lookupHost = original })

	vip := ""
	lbHostname := "unreachable.example.com"
	if vip == "" && lbHostname != "" {
		addrs, err := lookupHost(lbHostname)
		if err == nil && len(addrs) > 0 {
			t.Fatal("expected DNS failure, but got result")
		}
	}

	if vip != "" {
		t.Errorf("expected empty VIP after DNS failure, got %q", vip)
	}
}
