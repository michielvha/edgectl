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

type LoadBalancerConfig struct {
	ClusterID string
	IsMain    bool
	Interface string
	VIP       string
	Hostnames []string
}

// CreateLoadBalancer creates a new load balancer for the RKE2 cluster
// It determines if this node should be the primary or backup LB node
// and configures HAProxy and Keepalived accordingly
func CreateLoadBalancer(clusterID, vip string) error {
	logger.Info("Creating load balancer for RKE2 cluster")

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

	// Retrieve server nodes from Vault or set up initial config if this is the first LB
	hosts, existingVIP, err := client.RetrieveMasterInfo(clusterID)
	isFirst := err != nil || existingVIP == ""

	// If no VIP was provided and no existing VIP was found, error out
	if vip == "" && existingVIP == "" {
		return fmt.Errorf("no VIP provided and no existing VIP found in Vault")
	}

	// Use existing VIP if none provided
	if vip == "" {
		vip = existingVIP
	}

	// Determine network interface for VIP
	iface, err := detectInterfaceForVIP(vip)
	if err != nil {
		return fmt.Errorf("could not detect network interface for VIP %s: %w", vip, err)
	}

	// Configure this node as the main LB if it's the first one
	isMain := isFirst

	// Store the current LB info in Vault
	err = client.StoreLBInfo(clusterID, hostname, vip, isMain)
	if err != nil {
		return fmt.Errorf("failed to store load balancer info in Vault: %w", err)
	}

	// Bootstrap the load balancer
	return BootstrapLB(LoadBalancerConfig{
		ClusterID: clusterID,
		IsMain:    isMain,
		Interface: iface,
		VIP:       vip,
		Hostnames: hosts,
	})
}

func BootstrapLBFromVault(clusterID string, isMain bool) error {
	client, err := vault.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create Vault client: %w", err)
	}

	hosts, vip, err := client.RetrieveMasterInfo(clusterID)
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
	})
}

func BootstrapLB(cfg LoadBalancerConfig) error {
	priority := "100"
	state := "BACKUP"
	if cfg.IsMain {
		priority = "200"
		state = "MASTER"
	}

	logger.Info("🔧 Installing HAProxy and KeepAlived...")
	if err := installPackages(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	logger.Info("📄 Generating HAProxy config...")
	haproxyConfig, err := generateHAProxyConfig(cfg.Hostnames)
	if err != nil {
		return err
	}
	if err := os.WriteFile("/etc/haproxy/haproxy.cfg", []byte(haproxyConfig), 0644); err != nil {
		return fmt.Errorf("failed to write haproxy config: %w", err)
	}

	logger.Info("📄 Generating Keepalived config...")
	keepalivedConfig := generateKeepalivedConfig(cfg.Interface, cfg.VIP, state, priority)
	if err := os.WriteFile("/etc/keepalived/keepalived.conf", []byte(keepalivedConfig), 0644); err != nil {
		return fmt.Errorf("failed to write keepalived config: %w", err)
	}

	logger.Info("🚀 Restarting services...")
	if err := restartService("haproxy"); err != nil {
		return err
	}
	if err := restartService("keepalived"); err != nil {
		return err
	}

	logger.Info("%s", fmt.Sprintf("✅ Load balancer stack configured with VIP %s", cfg.VIP))
	return nil
}

func installPackages() error {
	cmd := exec.Command("bash", "-c", "apt-get update && apt-get install -y haproxy keepalived")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generateHAProxyConfig(hostnames []string) (string, error) {
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

	for _, host := range hostnames {
		ip, err := net.LookupIP(host)
		if err != nil || len(ip) == 0 {
			return "", fmt.Errorf("could not resolve IP for host %s: %v", host, err)
		}
		b.WriteString(fmt.Sprintf("    server %s %s:6443 check\n", host, ip[0].String()))
	}

	// Add supervisor API backend
	b.WriteString("\nbackend rke2-supervisor-backend\n    mode tcp\n    option tcp-check\n    balance roundrobin\n    default-server inter 10s downinter 5s rise 3 fall 3\n")

	for _, host := range hostnames {
		ip, err := net.LookupIP(host)
		if err != nil || len(ip) == 0 {
			return "", fmt.Errorf("could not resolve IP for host %s: %v", host, err)
		}
		b.WriteString(fmt.Sprintf("    server %s %s:9345 check\n", host, ip[0].String()))
	}

	return b.String(), nil
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
