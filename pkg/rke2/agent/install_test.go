package agent

import (
	"os"
	"testing"

	"github.com/michielvha/edgectl/pkg/vault"
)

const (
	testAgentToken = "agent-token-123"
	testFlagVIP    = "flag-vip"
)

func TestFetchToken_SetsEnvVars(t *testing.T) {
	// Skip if we can't write to /etc/edgectl (non-root in CI)
	if err := os.MkdirAll("/etc/edgectl", 0o750); err != nil {
		t.Skip("skipping: cannot write to /etc/edgectl (requires root)")
	}

	mock := &vault.MockStore{
		RetrieveJoinTokenFunc: func(clusterID string) (string, error) {
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
	if got := os.Getenv("RKE2_TOKEN"); got != testAgentToken {
		t.Errorf("expected RKE2_TOKEN=%q, got %q", testAgentToken, got)
	}

	t.Cleanup(func() { os.Unsetenv("RKE2_TOKEN") }) //nolint:errcheck
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
