# Architectural Overview of EdgeCTL

## Overview
EdgeCTL is a CLI tool designed to manage edge cloud infrastructure. It provides functionality for provisioning Kubernetes clusters, managing secrets, and interacting with load balancers. The architecture is modular, leveraging Go packages and external tools like HashiCorp Vault and RKE2.

---

## High-Level Architecture

```mermaid
graph TD
    A[Main CLI (edgectl)] -->|Commands| B[Command Handlers]
    B -->|RKE2 Management| C[RKE2 Commands]
    B -->|Vault Integration| D[Vault Commands]
    B -->|Version Info| E[Version Command]
    B -->|Load Balancer| F[Load Balancer Commands]

    C -->|Server Install| G[Server Installation Logic]
    C -->|Agent Install| H[Agent Installation Logic]
    C -->|Status| I[Status Check]
    C -->|Uninstall| J[Uninstall Logic]

    D -->|Secrets Management| K[Vault Client]
    D -->|Cluster Info| L[Cluster Metadata]

    F -->|HAProxy + Keepalived| M[Load Balancer Setup]

    subgraph "Core Packages"
        P1[Logger]
        P2[Common Utilities]
        P3[Vault Integration]
        P4[RKE2 Server Logic]
        P5[Load Balancer Handler]
    end

    A -->|Uses| P1
    B -->|Uses| P2
    D -->|Uses| P3
    G -->|Uses| P4
    M -->|Uses| P5
```

---

## Components

### 1. **Main CLI**
- **File:** `main.go`
- **Description:** Entry point for the CLI. Delegates execution to the `cmd` package.

### 2. **Command Handlers**
- **Directory:** `cmd/`
- **Description:** Contains subcommands for managing RKE2, Vault, and load balancers.
  - `rke2.go`: Handles RKE2 cluster operations.
  - `vault.go`: Manages secrets in HashiCorp Vault.
  - `version.go`: Displays CLI version.
  - `rke2/lb/commands.go`: Manages load balancer setup and status.

### 3. **Core Packages**
- **Logger**
  - **File:** `pkg/logger/log.go`
  - **Description:** Provides logging functionality using `zerolog`.

- **Common Utilities**
  - **File:** `pkg/common/`
  - **Description:** Contains shared utilities like embedded scripts and helper functions.

- **Vault Integration**
  - **File:** `pkg/vault/`
  - **Description:** Handles interactions with HashiCorp Vault for secrets management.

- **RKE2 Server Logic**
  - **File:** `pkg/rke2/server/install.go`
  - **Description:** Implements logic for installing and managing RKE2 servers.

- **Load Balancer Handler**
  - **File:** `pkg/lb/handler.go`
  - **Description:** Manages HAProxy and Keepalived configurations for load balancing.

---

## External Dependencies

### 1. **HashiCorp Vault**
- Used for secure storage and retrieval of secrets.

### 2. **RKE2**
- Lightweight Kubernetes distribution for edge environments.

### 3. **HAProxy + Keepalived**
- Provides high availability and load balancing for RKE2 clusters.

---

## Future Enhancements
- Add support for Fedora-based architectures.
- Implement a debug command for connectivity verification.
- Introduce multi-tenant support via Vault namespaces.

---

## Diagram Legend
- **Main CLI:** Entry point for the application.
- **Command Handlers:** Subcommands for specific functionalities.
- **Core Packages:** Reusable logic and utilities.
- **External Dependencies:** Third-party tools and services.