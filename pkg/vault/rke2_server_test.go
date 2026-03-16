package vault

import (
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
