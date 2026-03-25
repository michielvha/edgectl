/*
Copyright © 2025 VH & Co - contact@vhco.pro
*/
package lb

import (
	"fmt"
	"net"
	"strings"
	"testing"

	"github.com/michielvha/edgectl/pkg/vault"
)

// --- generateHAProxyConfig tests ---

func TestGenerateHAProxyConfig_RKE2_MultipleHosts(t *testing.T) {
	hostIPs := map[string]string{
		"master1": "10.0.0.1",
		"master2": "10.0.0.2",
		"master3": "10.0.0.3",
	}
	hostnames := []string{"master1", "master2", "master3"}

	config, err := generateHAProxyConfig(hostnames, hostIPs, "rke2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify 6443 backend entries
	for _, h := range hostnames {
		expected6443 := "server " + h + " " + hostIPs[h] + ":6443 check"
		if !strings.Contains(config, expected6443) {
			t.Errorf("missing 6443 backend entry for %s; want %q in config", h, expected6443)
		}
		expected9345 := "server " + h + " " + hostIPs[h] + ":9345 check"
		if !strings.Contains(config, expected9345) {
			t.Errorf("missing 9345 backend entry for %s; want %q in config", h, expected9345)
		}
	}

	// RKE2 should include supervisor frontend/backend
	if !strings.Contains(config, "rke2-supervisor-frontend") {
		t.Error("missing rke2-supervisor-frontend for rke2 distro")
	}
	if !strings.Contains(config, "rke2-supervisor-backend") {
		t.Error("missing rke2-supervisor-backend for rke2 distro")
	}
}

func TestGenerateHAProxyConfig_RKE2_EmptyHosts(t *testing.T) {
	config, err := generateHAProxyConfig(nil, nil, "rke2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still have the frontend/backend structure, just no server lines
	if !strings.Contains(config, "frontend k8s-api-frontend") {
		t.Error("missing k8s-api-frontend section")
	}
	if !strings.Contains(config, "backend k8s-api-backend") {
		t.Error("missing k8s-api-backend section")
	}
	if !strings.Contains(config, "backend rke2-supervisor-backend") {
		t.Error("missing rke2-supervisor-backend section")
	}
	// No "server " lines should be present
	lines := strings.Split(config, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "server ") {
			t.Errorf("unexpected server line in empty-host config: %q", trimmed)
		}
	}
}

func TestGenerateHAProxyConfig_RKE2_SingleHost(t *testing.T) {
	hostIPs := map[string]string{"node1": "192.168.1.10"}
	config, err := generateHAProxyConfig([]string{"node1"}, hostIPs, "rke2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	count := strings.Count(config, "server node1")
	if count != 2 {
		t.Errorf("expected 2 server lines (6443+9345), got %d", count)
	}
}

// --- K3s-specific HAProxy tests ---

func TestGenerateHAProxyConfig_K3s_NoSupervisorPort(t *testing.T) {
	hostIPs := map[string]string{
		"master1": "10.0.0.1",
		"master2": "10.0.0.2",
	}
	hostnames := []string{"master1", "master2"}

	config, err := generateHAProxyConfig(hostnames, hostIPs, "k3s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// K3s should NOT have supervisor frontend/backend
	if strings.Contains(config, "rke2-supervisor-frontend") {
		t.Error("k3s config should not contain rke2-supervisor-frontend")
	}
	if strings.Contains(config, "rke2-supervisor-backend") {
		t.Error("k3s config should not contain rke2-supervisor-backend")
	}
	if strings.Contains(config, ":9345") {
		t.Error("k3s config should not contain port 9345")
	}

	// Should have 6443 entries
	for _, h := range hostnames {
		expected6443 := "server " + h + " " + hostIPs[h] + ":6443 check"
		if !strings.Contains(config, expected6443) {
			t.Errorf("missing 6443 backend entry for %s; want %q in config", h, expected6443)
		}
	}
}

func TestGenerateHAProxyConfig_K3s_SingleHost(t *testing.T) {
	hostIPs := map[string]string{"node1": "192.168.1.10"}
	config, err := generateHAProxyConfig([]string{"node1"}, hostIPs, "k3s")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// K3s: only 1 server line (6443 only, no 9345)
	count := strings.Count(config, "server node1")
	if count != 1 {
		t.Errorf("expected 1 server line (6443 only for k3s), got %d", count)
	}
}

// --- addServersToBackend tests ---

func TestAddServersToBackend_UsesHostIPsMap(t *testing.T) {
	var b strings.Builder
	hostIPs := map[string]string{
		"host-a": "10.1.1.1",
		"host-b": "10.1.1.2",
	}
	addServersToBackend(&b, []string{"host-a", "host-b"}, hostIPs, 6443)

	result := b.String()
	if !strings.Contains(result, "server host-a 10.1.1.1:6443 check") {
		t.Errorf("missing host-a entry, got:\n%s", result)
	}
	if !strings.Contains(result, "server host-b 10.1.1.2:6443 check") {
		t.Errorf("missing host-b entry, got:\n%s", result)
	}
}

func TestAddServersToBackend_EmptyHostnames(t *testing.T) {
	var b strings.Builder
	addServersToBackend(&b, []string{}, nil, 6443)
	if b.Len() != 0 {
		t.Errorf("expected empty output for empty hostnames, got: %q", b.String())
	}
}

func TestAddServersToBackend_HostNotInMap_FallsDNS(t *testing.T) {
	// When a host is NOT in hostIPs, it falls through to DNS lookup.
	// With an unresolvable hostname, the host is skipped silently.
	var b strings.Builder
	addServersToBackend(&b, []string{"nonexistent.invalid.test"}, map[string]string{}, 9345)

	result := b.String()
	if strings.Contains(result, "nonexistent.invalid.test") {
		t.Errorf("expected unresolvable host to be skipped, got: %q", result)
	}
}

func TestAddServersToBackend_DNSFallback_Resolves(t *testing.T) {
	// Inject a fake lookupIP that returns a known IP.
	original := lookupIP
	lookupIP = func(host string) ([]net.IP, error) {
		if host == "dns-host" {
			return []net.IP{net.ParseIP("10.99.99.1")}, nil
		}
		return nil, fmt.Errorf("not found")
	}
	t.Cleanup(func() { lookupIP = original })

	var b strings.Builder
	addServersToBackend(&b, []string{"dns-host"}, map[string]string{}, 6443)

	result := b.String()
	if !strings.Contains(result, "server dns-host 10.99.99.1:6443 check") {
		t.Errorf("expected resolved DNS entry, got:\n%s", result)
	}
}

func TestAddServersToBackend_DNSFallback_Error(t *testing.T) {
	// Inject a failing lookupIP — host should be skipped.
	original := lookupIP
	lookupIP = func(host string) ([]net.IP, error) {
		return nil, fmt.Errorf("DNS resolution failed")
	}
	t.Cleanup(func() { lookupIP = original })

	var b strings.Builder
	addServersToBackend(&b, []string{"fail-host"}, map[string]string{}, 9345)

	if b.Len() != 0 {
		t.Errorf("expected no output for failed DNS, got: %q", b.String())
	}
}

func TestAddServersToBackend_MapTakesPrecedenceOverDNS(t *testing.T) {
	// Even if lookupIP would return something, the hostIPs map should win.
	original := lookupIP
	lookupIP = func(host string) ([]net.IP, error) {
		t.Error("lookupIP should not be called when host is in map")
		return []net.IP{net.ParseIP("99.99.99.99")}, nil
	}
	t.Cleanup(func() { lookupIP = original })

	var b strings.Builder
	addServersToBackend(&b, []string{"cached-host"}, map[string]string{"cached-host": "10.0.0.5"}, 6443)

	result := b.String()
	if !strings.Contains(result, "10.0.0.5:6443") {
		t.Errorf("expected cached IP, got:\n%s", result)
	}
}

// --- generateKeepalivedConfig tests ---

func TestGenerateKeepalivedConfig_Master(t *testing.T) {
	config := generateKeepalivedConfig("eth0", "10.0.0.100", "MASTER", "200")

	if !strings.Contains(config, "interface eth0") {
		t.Error("missing interface directive")
	}
	if !strings.Contains(config, "state MASTER") {
		t.Error("missing MASTER state")
	}
	if !strings.Contains(config, "priority 200") {
		t.Error("missing priority 200")
	}
	if !strings.Contains(config, "10.0.0.100/24") {
		t.Error("missing VIP in virtual_ipaddress block")
	}
}

func TestGenerateKeepalivedConfig_Backup(t *testing.T) {
	config := generateKeepalivedConfig("ens192", "172.16.0.50", "BACKUP", "100")

	if !strings.Contains(config, "interface ens192") {
		t.Error("missing interface directive")
	}
	if !strings.Contains(config, "state BACKUP") {
		t.Error("missing BACKUP state")
	}
	if !strings.Contains(config, "priority 100") {
		t.Error("missing priority 100")
	}
	if !strings.Contains(config, "172.16.0.50/24") {
		t.Error("missing VIP in virtual_ipaddress block")
	}
}

func TestGenerateKeepalivedConfig_ContainsHealthCheck(t *testing.T) {
	config := generateKeepalivedConfig("eth0", "10.0.0.1", "MASTER", "200")

	if !strings.Contains(config, "vrrp_script chk_haproxy") {
		t.Error("missing haproxy health check script block")
	}
	if !strings.Contains(config, "track_script") {
		t.Error("missing track_script block")
	}
}

// --- GetStatus tests (with mock store) ---

func TestGetStatus_ReturnsNodesAndVIP(t *testing.T) {
	mock := &vault.MockStore{
		RetrieveLBInfoFunc: func(distro, clusterID string) ([]map[string]interface{}, string, error) {
			return []map[string]interface{}{
				{"hostname": "lb1", "is_main": true},
				{"hostname": "lb2", "is_main": false},
			}, "10.0.0.100", nil
		},
	}

	vip, nodes, err := GetStatus(mock, "rke2", "test-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vip != "10.0.0.100" {
		t.Errorf("expected VIP '10.0.0.100', got %q", vip)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Hostname != "lb1" || !nodes[0].IsMain {
		t.Errorf("expected lb1 as MAIN, got %+v", nodes[0])
	}
	if nodes[1].Hostname != "lb2" || nodes[1].IsMain {
		t.Errorf("expected lb2 as BACKUP, got %+v", nodes[1])
	}
}

func TestGetStatus_EmptyNodes(t *testing.T) {
	mock := &vault.MockStore{
		RetrieveLBInfoFunc: func(distro, clusterID string) ([]map[string]interface{}, string, error) {
			return []map[string]interface{}{}, "", nil
		},
	}

	vip, nodes, err := GetStatus(mock, "rke2", "empty-cluster")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vip != "" {
		t.Errorf("expected empty VIP, got %q", vip)
	}
	if len(nodes) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(nodes))
	}
}
