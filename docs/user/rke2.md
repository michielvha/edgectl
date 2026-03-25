# RKE2 Cluster Management

## 🧭 Overview

This CLI enables automated deployment of **RKE2 Kubernetes clusters** across bare-metal or hybrid environments using a simple one-liner. It leverages:

- **OpenBao** (secret store) to securely store and retrieve join tokens
- A persistent **cluster ID** for every RKE2 control plane
- Embedded bash scripts for modular system-level execution

---

## 🧩 Core Concepts

### ✅ Cluster ID

- A **unique identifier** generated at control plane bootstrap (e.g., `rke2-abc12345`)
- Stored persistently at `/etc/edgectl/cluster-id`
- Used as the secret store key for all token-related operations
- Ensures agents can connect to the right control plane without handling raw tokens

---

### 🔐 Secret Store Integration

- All tokens are stored/retrieved using the **Cluster ID** path:
  ```
  kv/data/rke2/<cluster-id>/token
  ```
- Users never manually copy tokens
- Tokens are only exposed during bootstrap and handled programmatically afterward

---

## ⚙️ Workflow

### 🧪 1. `edgectl rke2 server`

```bash
edgectl rke2 server
```

- Installs the RKE2 control plane via embedded bash script
- If `--token` is **not provided**:
  - Generates a new Cluster ID (`rke2-xxxxxxx`)
  - Persists it to `/etc/edgectl/cluster-id`
  - Optionally stores the generated join token in the secret store
- If `--token` **is provided**:
  - Skips Cluster ID generation (assumes it's a secondary master)

---

### 🧑‍🤝‍🧑 2. `edgectl rke2 agent`

```bash
edgectl rke2 agent --cluster-id rke2-abc12345
```

- Requires `--cluster-id` (which is actually the **Cluster ID**)
- Uses the provided Cluster ID to fetch the join token from the secret store
- Joins the agent to the control plane securely
- Token never passed around or embedded in files/scripts

---

## 🔄 Token Lifecycle

| Step                     | Action                                                           |
|--------------------------|------------------------------------------------------------------|
| Server bootstrap         | Token generated and stored in secret store under `/rke2/<cluster-id>` |
| Agent installation       | Token retrieved from secret store using Cluster ID              |
| Additional master nodes  | Optionally use the same Cluster ID for HA setup                 |

---

## 📁 File Layout

| Path                                | Purpose                                |
|-------------------------------------|----------------------------------------|
| `/etc/edgectl/cluster-id`          | Stores generated Cluster ID            |
| `kv/data/rke2/<cluster-id>/` (OpenBao) | Join token, kubeconfig, masters, LB info for that cluster |
| `scripts/rke2.sh` (embedded)        | Bash functions for RKE2 lifecycle      |

---

## 🔮 Future Plans

- Auto-store cluster metadata (creation time, hostname, IP) in the secret store
- Add support for multi-tenant environments via OpenBao namespaces or tags
- Abstract even more bash logic into Go
- Support `edgectl upgrade`, `edgectl add-master`, etc.