/*
Copyright © 2025 EDGEFORGE contact@edgeforge.eu
*/
package lb

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/michielvha/edgectl/pkg/logger"
	vault "github.com/michielvha/edgectl/pkg/vault"
)

// LoadBalancerConfig struct defines the configuration for a load balancer
type LoadBalancerConfig struct {
	ClusterID string
	IsMain    bool
	Interface string
	VIP       string
	Hostnames []string
	HostIPs   map[string]string
}

// CreateLoadBalancer creates a new load balancer for the RKE2 cluster
// It determines if this node should be the primary or backup LB node
// and configures HAProxy and Keepalived accordingly
func CreateLoadBalancer(clusterID, vip string) error {
	logger.Debug("Creating load balancer for RKE2 cluster")
	fmt.Printf("Creating load balancer for RKE2 cluster %s\n", clusterID)

	// Get the current hostname
	hostname, err := os.Hostname()
	if err != nil {
		return fmt.Errorf("failed to get hostname: %w", err)
	}

	// Connect to Vault
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	// First check if there are any existing load balancers
	existingLBs, existingVIP, err := client.RetrieveLBInfo(clusterID)

	// isFirst is true if there are no existing load balancers
	isFirst := err != nil || len(existingLBs) == 0

	logger.Debug("Load balancer first node check: isFirst=%v, error=%v, existingLBCount=%d",
		isFirst, err, len(existingLBs))

	// Retrieve server nodes from Vault for HAProxy configuration
	hosts, masterVIP, hostIPs, err := client.RetrieveMasterInfo(clusterID)
	if err != nil {
		logger.Debug("No master nodes found, this might be a new cluster: %v", err)
	}

	// Determine which VIP to use (priority: provided VIP > existing LB VIP > master VIP)
	effectiveVIP := vip
	if effectiveVIP == "" && existingVIP != "" {
		effectiveVIP = existingVIP
	} else if effectiveVIP == "" && masterVIP != "" {
		effectiveVIP = masterVIP
	}

	// If no VIP was determined, error out
	if effectiveVIP == "" {
		return fmt.Errorf("no VIP provided and no existing VIP found in Vault")
	}

	// Determine network interface for VIP
	iface, err := detectInterfaceForVIP(effectiveVIP)
	if err != nil {
		return fmt.Errorf("could not detect network interface for VIP %s: %w", effectiveVIP, err)
	}

	// Configure this node as the main LB if it's the first one
	isMain := isFirst

	// Store the current LB info in Vault
	err = client.StoreLBInfo(clusterID, hostname, effectiveVIP, isMain)
	if err != nil {
		return fmt.Errorf("failed to store load balancer info in Vault: %w", err)
	}

	// Bootstrap the load balancer
	return BootstrapLB(LoadBalancerConfig{
		ClusterID: clusterID,
		IsMain:    isMain,
		Interface: iface,
		VIP:       effectiveVIP,
		Hostnames: hosts,
		HostIPs:   hostIPs,
	})
}

func BootstrapLBFromVault(clusterID string, isMain bool) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	hosts, vip, hostIPs, err := client.RetrieveMasterInfo(clusterID)
	if err != nil {
		return fmt.Errorf("failed to fetch master info from Vault: %w", err)
	}

	iface, err := detectInterfaceForVIP(vip)
	if err != nil {
		return fmt.Errorf("could not detect network interface for VIP %s: %w", vip, err)
	}

	return BootstrapLB(LoadBalancerConfig{
		ClusterID: clusterID,
		IsMain:    isMain,
		Interface: iface,
		VIP:       vip,
		Hostnames: hosts,
		HostIPs:   hostIPs,
	})
}

func BootstrapLB(cfg LoadBalancerConfig) error {
	priority := "100"
	state := "BACKUP"
	if cfg.IsMain {
		priority = "200"
		state = "MASTER"
	}

	fmt.Print("🔧 Installing HAProxy and KeepAlived... \n")
	if err := installPackages(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	fmt.Print("📄 Generating HAProxy config... \n")
	haproxyConfig, err := generateHAProxyConfig(cfg.Hostnames, cfg.HostIPs)
	if err != nil {
		return err
	}
	if err := os.WriteFile("/etc/haproxy/haproxy.cfg", []byte(haproxyConfig), 0o644); err != nil {
		return fmt.Errorf("failed to write haproxy config: %w", err)
	}

	fmt.Print("📄 Generating Keepalived config... \n")
	keepalivedConfig := generateKeepalivedConfig(cfg.Interface, cfg.VIP, state, priority)
	if err := os.WriteFile("/etc/keepalived/keepalived.conf", []byte(keepalivedConfig), 0o644); err != nil {
		return fmt.Errorf("failed to write keepalived config: %w", err)
	}

	fmt.Print("🚀 Restarting services... \n")
	if err := restartService("haproxy"); err != nil {
		return err
	}
	if err := restartService("keepalived"); err != nil {
		return err
	}

	fmt.Printf("✅ Load balancer stack configured with VIP %s \n", cfg.VIP)
	return nil
}

func installPackages() error {
	cmd := exec.Command("bash", "-c", "apt-get update && apt-get install -y haproxy keepalived")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generateHAProxyConfig(hostnames []string, hostIPs map[string]string) (string, error) {
	var b strings.Builder
	b.WriteString(`# HAProxy Configuration for RKE2 Load Balancing
global
    log /dev/log local0
    log /dev/log local1 notice
    chroot /var/lib/haproxy
    stats socket /run/haproxy/admin.sock mode 660 level admin expose-fd listeners
    stats timeout 30s
    user haproxy
    group haproxy
    daemon

defaults
    log     global
    mode    tcp
    option  tcplog
    option  dontlognull
    timeout connect 5000
    timeout client  50000
    timeout server  50000
    errorfile 400 /etc/haproxy/errors/400.http
    errorfile 403 /etc/haproxy/errors/403.http
    errorfile 408 /etc/haproxy/errors/408.http
    errorfile 500 /etc/haproxy/errors/500.http
    errorfile 502 /etc/haproxy/errors/502.http
    errorfile 503 /etc/haproxy/errors/503.http
    errorfile 504 /etc/haproxy/errors/504.http

frontend k3s-frontend
    bind *:6443
    mode tcp
    option tcplog
    default_backend k3s-backend

# Frontend for RKE2 supervisor API
frontend rke2-supervisor-frontend
    bind *:9345
    mode tcp
    option tcplog
    default_backend rke2-supervisor-backend

backend k3s-backend
    mode tcp
    option tcp-check
    balance roundrobin
    default-server inter 10s downinter 5s rise 3 fall 3

`)

	// Add servers to the k3s backend (port 6443)
	addServersToBackend(&b, hostnames, hostIPs, 6443)

	// Add supervisor API backend
	b.WriteString("\nbackend rke2-supervisor-backend\n    mode tcp\n    option tcp-check\n    balance roundrobin\n    default-server inter 10s downinter 5s rise 3 fall 3\n")

	// Add servers to the supervisor backend (port 9345)
	addServersToBackend(&b, hostnames, hostIPs, 9345)

	return b.String(), nil
}

// Helper function to add servers to a HAProxy backend
func addServersToBackend(b *strings.Builder, hostnames []string, hostIPs map[string]string, port int) {
	for _, host := range hostnames {
		// First try to get IP from the hostIPs map (cached IPs from Vault)
		if ip, ok := hostIPs[host]; ok {
			fmt.Fprintf(b, "    server %s %s:%d check\n", host, ip, port)
			logger.Debug("Using IP %s from Vault for host %s", ip, host)
			continue
		}

		// Fallback to DNS lookup if IP not found in Vault
		ipAddrs, err := net.LookupIP(host)
		if err != nil || len(ipAddrs) == 0 {
			logger.Warn("Could not resolve IP for host %s via DNS, skipping: %v", host, err)
			// Instead of failing, skip this host and continue
			continue
		}
		fmt.Fprintf(b, "    server %s %s:%d check\n", host, ipAddrs[0].String(), port)
		logger.Debug("Using IP %s from DNS for host %s", ipAddrs[0].String(), host)
	}
}

func generateKeepalivedConfig(iface, vip, state, priority string) string {
	return fmt.Sprintf(`# Keepalived configuration for RKE2 VIP
global_defs {
  enable_script_security
  script_user root
}

vrrp_script chk_haproxy {
    script 'killall -0 haproxy'
    interval 2
}

vrrp_instance haproxy-vip {
    interface %s
    state %s
    priority %s
    virtual_router_id 51

    virtual_ipaddress {
        %s/24
    }

    track_script {
        chk_haproxy
    }
}
`, iface, state, priority, vip)
}

func restartService(name string) error {
	cmd := exec.Command("systemctl", "restart", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func detectInterfaceForVIP(vip string) (string, error) {
	// Try to find the interface that would be used to reach the VIP
	out, err := exec.Command("bash", "-c", fmt.Sprintf("ip route get %s | grep -o 'dev [^ ]*' | awk '{print $2}'", vip)).Output()
	if err != nil {
		// If that fails, try to find the primary interface
		out, err = exec.Command("bash", "-c", "ip route | grep default | grep -o 'dev [^ ]*' | awk '{print $2}'").Output()
		if err != nil {
			return "", fmt.Errorf("failed to detect network interface: %w", err)
		}
	}
	return strings.TrimSpace(string(out)), nil
}
