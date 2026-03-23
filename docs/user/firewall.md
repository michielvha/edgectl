# Firewall Configuration

## Overview

EdgeCTL automatically configures firewall rules during cluster installation. It supports multiple firewall backends and auto-detects which one is available on the host, making it compatible with both Debian-based and Fedora/RHEL-based distributions.

---

## Supported Firewall Backends

| Backend | Detection | Typical Distributions | Persistence |
|---------|-----------|----------------------|-------------|
| **UFW** | `ufw` command available | Ubuntu, Debian | Persistent by default |
| **firewalld** | `firewall-cmd` command available | Fedora, RHEL, CentOS, Rocky Linux, AlmaLinux | Persistent (`--permanent` flag) |
| **iptables** | `iptables` command available | Any Linux (fallback) | Not persistent (rules lost on reboot) |

If no supported firewall is detected, edgectl will skip firewall configuration and print a warning.

### Detection Priority

Firewalls are detected in the following order:

1. **UFW** — checked first (preferred on Debian-based systems)
2. **firewalld** — checked second (default on Fedora/RHEL)
3. **iptables** — fallback if neither UFW nor firewalld is installed
4. **none** — no firewall configured; ports are skipped with a warning

---

## Supported Distributions

| Distribution | Default Firewall | Status |
|--------------|-----------------|--------|
| Ubuntu | UFW | Fully supported |
| Debian | UFW | Fully supported |
| Fedora | firewalld | Fully supported |
| RHEL | firewalld | Fully supported |
| CentOS | firewalld | Fully supported |
| Rocky Linux | firewalld | Fully supported |
| AlmaLinux | firewalld | Fully supported |

---

## Port Requirements

### RKE2 Server Node

| Port | Protocol | Purpose |
|------|----------|---------|
| 22 | TCP | SSH access |
| 6443 | TCP | Kubernetes API Server |
| 9345 | TCP | RKE2 Supervisor API |
| 10250 | TCP | kubelet metrics |
| 2379 | TCP | etcd client |
| 2380 | TCP | etcd peer |
| 2381 | TCP | etcd metrics |
| 30000-32767 | TCP | Kubernetes NodePort range |

### K3s Server Node

| Port | Protocol | Purpose |
|------|----------|---------|
| 22 | TCP | SSH access |
| 6443 | TCP | Kubernetes API Server |
| 10250 | TCP | kubelet metrics |
| 2379 | TCP | etcd client |
| 2380 | TCP | etcd peer |
| 30000-32767 | TCP | Kubernetes NodePort range |

> **Note:** K3s does not use a separate supervisor port — both API and supervisor traffic go through 6443.

### Agent Node (RKE2 & K3s)

| Port | Protocol | Purpose |
|------|----------|---------|
| 22 | TCP | SSH access |
| 10250 | TCP | kubelet metrics |
| 30000-32767 | TCP | Kubernetes NodePort range |

---

## How It Works

During `edgectl <distro> server install` or `edgectl <distro> agent install`, the embedded scripts:

1. **Detect** the available firewall backend using `detect_firewall()`
2. **Allow** each required port using `firewall_allow_port()`
3. **Enable/reload** the firewall using `firewall_enable()`

### UFW Example

```
ufw allow proto tcp from any to any port 6443 comment "Kubernetes API Server"
ufw allow proto tcp from any to any port 9345 comment "RKE2 Supervisor API"
ufw --force enable
```

UFW rules include descriptive comments for easy identification with `ufw status verbose`.

### firewalld Example

```
firewall-cmd --permanent --add-port=6443/tcp
firewall-cmd --permanent --add-port=9345/tcp
firewall-cmd --reload
```

Rules are added with the `--permanent` flag and applied via `--reload`.

### iptables Example

```
iptables -A INPUT -p tcp --dport 6443 -j ACCEPT
iptables -A INPUT -p tcp --dport 30000:32767 -m multiport -j ACCEPT
```

Port ranges use the `-m multiport` module. Note that iptables rules are **not persisted** across reboots — consider installing `iptables-persistent` or use UFW/firewalld instead.

---

## Verifying Firewall Rules

### UFW

```bash
sudo ufw status verbose
```

### firewalld

```bash
sudo firewall-cmd --list-ports
sudo firewall-cmd --list-all
```

### iptables

```bash
sudo iptables -L INPUT -n --line-numbers
```

---

## Troubleshooting

### Nodes can't join the cluster

Verify that the required ports are open on the server node:

```bash
# Test API server connectivity from the agent
curl -k https://<server-ip>:6443

# For RKE2, also test the supervisor port
curl -k https://<server-ip>:9345
```

### Firewall rules not persisting (iptables)

If you're using the iptables fallback, rules are lost on reboot. Install a persistence mechanism:

```bash
# Debian/Ubuntu
sudo apt install iptables-persistent
sudo netfilter-persistent save

# Fedora/RHEL
sudo dnf install iptables-services
sudo service iptables save
```

Or switch to UFW or firewalld for built-in persistence.

### No firewall detected

If edgectl reports "No supported firewall detected", install one:

```bash
# Debian/Ubuntu
sudo apt install ufw

# Fedora/RHEL
sudo dnf install firewalld
sudo systemctl enable --now firewalld
```
