package lb

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/michielvha/edgectl/pkg/vault/rke2"
)

type LoadBalancerConfig struct {
	ClusterID string
	IsMain    bool
	Interface string
	VIP       string
	Hostnames []string
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

	fmt.Println("🔧 Installing HAProxy and KeepAlived...")
	if err := installPackages(); err != nil {
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	fmt.Println("📄 Generating HAProxy config...")
	haproxyConfig, err := generateHAProxyConfig(cfg.Hostnames)
	if err != nil {
		return err
	}
	if err := os.WriteFile("/etc/haproxy/haproxy.cfg", []byte(haproxyConfig), 0644); err != nil {
		return fmt.Errorf("failed to write haproxy config: %w", err)
	}

	keepalivedConfig := generateKeepalivedConfig(cfg.Interface, cfg.VIP, state, priority)
	if err := os.WriteFile("/etc/keepalived/keepalived.conf", []byte(keepalivedConfig), 0644); err != nil {
		return fmt.Errorf("failed to write keepalived config: %w", err)
	}

	fmt.Println("🚀 Restarting services...")
	if err := restartService("haproxy"); err != nil {
		return err
	}
	if err := restartService("keepalived"); err != nil {
		return err
	}

	fmt.Println("✅ Load balancer stack configured with VIP", cfg.VIP)
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
	b.WriteString("frontend k3s-frontend\n    bind *:6443\n    mode tcp\n    option tcplog\n    default_backend k3s-backend\n\n")
	b.WriteString("backend k3s-backend\n    mode tcp\n    option tcp-check\n    balance roundrobin\n    default-server inter 10s downinter 5s\n")

	for _, host := range hostnames {
		ip, err := net.LookupIP(host)
		if err != nil || len(ip) == 0 {
			return "", fmt.Errorf("could not resolve IP for host %s: %v", host, err)
		}
		b.WriteString(fmt.Sprintf("    server %s %s:6443 check\n", host, ip[0].String()))
	}

	return b.String(), nil
}

func generateKeepalivedConfig(iface, vip, state, priority string) string {
	return fmt.Sprintf(`global_defs {
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
	out, err := exec.Command("bash", "-c", fmt.Sprintf("ip route get %s | grep -o 'dev [^ ]*' | awk '{print $2}'", vip)).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
