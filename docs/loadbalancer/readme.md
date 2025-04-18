# Load Balancer

**TODO: explain how to auto generate load balancers**

## troubleshooting

Here are some tests you can run to verify your load balancer is working correctly:

### 1. Verify the virtual IP is assigned

Check if the VIP is properly assigned to an interface:
```bash
ip addr show | grep -A2 "inet.*172\."
```

### 2. Test Kubernetes API connectivity

Test connectivity to the Kubernetes API through the load balancer:
```bash
curl -k https://<your-vip>:6443
```

You should get a response like "Unauthorized" or "Forbidden" which is expected without credentials, but means the connection works.

### 3. Check HAProxy stats

Check the HAProxy backend health status:
```bash
echo "show stat" | socat unix-connect:/run/haproxy/admin.sock stdio
```

### 4. Test high availability

You can test HA functionality by simulating a failure:

- Stop HAProxy on the master node:
  ```bash
  systemctl stop haproxy
  ```

- Check if Keepalived detects the failure and transitions to BACKUP state:
  ```bash
  journalctl -u keepalived -n 20
  ```

- Verify the VIP moves to another node (if you have multiple LB nodes)

- Restart HAProxy and check if the VIP is reclaimed:
  ```bash
  systemctl start haproxy
  journalctl -u keepalived -n 20
  ```

### 5. Check load balancing to your RKE2 servers

To verify traffic is being distributed to all your RKE2 servers, check the logs:
```bash
tcpdump -i any port 6443 -n
```

While running this, connect to the Kubernetes API via the VIP in another terminal. You should see traffic flowing to the backend servers.

### 6. Test with kubectl

If you have a kubeconfig file configured to use the VIP, try basic kubectl commands:
```bash
kubectl --kubeconfig=/path/to/kubeconfig get nodes
```

These tests should confirm that your load balancer is working properly, providing HA for your RKE2 cluster, and correctly routing traffic to your servers.