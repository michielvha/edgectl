package vault

import (
	"fmt"
	"testing"
)

func TestGetFirstMasterIP_EmptyHosts(t *testing.T) {
	result := getFirstMasterIP([]string{}, map[string]string{}, "192.168.1.1")
	if result != "192.168.1.1" {
		t.Errorf("expected currentIP fallback '192.168.1.1', got %q", result)
	}
}

func TestGetFirstMasterIP_FirstHostInMap(t *testing.T) {
	hosts := []string{"master1", "master2"}
	hostIPs := map[string]string{
		"master1": "10.0.0.1",
		"master2": "10.0.0.2",
	}
	result := getFirstMasterIP(hosts, hostIPs, "192.168.1.1")
	if result != "10.0.0.1" {
		t.Errorf("expected '10.0.0.1' from hostIPs map, got %q", result)
	}
}

func TestGetFirstMasterIP_FirstHostNotInMap(t *testing.T) {
	hosts := []string{"master1", "master2"}
	hostIPs := map[string]string{
		"master2": "10.0.0.2",
	}
	result := getFirstMasterIP(hosts, hostIPs, "192.168.1.1")
	if result != "master1" {
		t.Errorf("expected hostname fallback 'master1', got %q", result)
	}
}

func TestGetFirstMasterIP_NilHostIPs(t *testing.T) {
	hosts := []string{"master1"}
	result := getFirstMasterIP(hosts, nil, "192.168.1.1")
	if result != "master1" {
		t.Errorf("expected hostname fallback 'master1', got %q", result)
	}
}

// --- getHostIP tests (with injected lookupHost) ---

func TestGetHostIP_Success(t *testing.T) {
	original := lookupHost
	lookupHost = func(host string) ([]string, error) {
		if host == "myhost" {
			return []string{"10.0.0.42"}, nil
		}
		return nil, fmt.Errorf("not found")
	}
	t.Cleanup(func() { lookupHost = original })

	ip, err := getHostIP("myhost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.0.0.42" {
		t.Errorf("expected '10.0.0.42', got %q", ip)
	}
}

func TestGetHostIP_LookupError(t *testing.T) {
	original := lookupHost
	lookupHost = func(host string) ([]string, error) {
		return nil, fmt.Errorf("DNS failure")
	}
	t.Cleanup(func() { lookupHost = original })

	_, err := getHostIP("badhost")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetHostIP_EmptyAddrs(t *testing.T) {
	original := lookupHost
	lookupHost = func(host string) ([]string, error) {
		return []string{}, nil
	}
	t.Cleanup(func() { lookupHost = original })

	_, err := getHostIP("emptyhost")
	if err == nil {
		t.Fatal("expected error for empty addrs, got nil")
	}
}

func TestGetHostIP_MultipleAddrs_ReturnsFirst(t *testing.T) {
	original := lookupHost
	lookupHost = func(host string) ([]string, error) {
		return []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}, nil
	}
	t.Cleanup(func() { lookupHost = original })

	ip, err := getHostIP("multihost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ip != "10.0.0.1" {
		t.Errorf("expected first addr '10.0.0.1', got %q", ip)
	}
}
