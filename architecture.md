# Architectural Overview of EdgeCTL

## Overview
EdgeCTL is a CLI tool designed to manage edge cloud infrastructure. It provides functionality for provisioning Kubernetes clusters, managing secrets, and interacting with load balancers. The architecture is modular, leveraging Go packages and external tools like HashiCorp Vault and RKE2.

---

## High-Level Architecture

```mermaid
graph TD
    A[Main CLI (edgectl)] --> B[Command Handlers]
    B --> C[RKE2 Commands]
    B --> D[Vault Commands]
    B --> E[Version Command]
    B --> F[Load Balancer Commands]

    C --> G[Server Installation Logic]
    C --> H[Agent Installation Logic]
    C --> I[Status Check]
    C --> J[Uninstall Logic]

    D --> K[Vault Client]
    D --> L[Cluster Metadata]

    F --> M[Load Balancer Setup]

    subgraph CorePackages
        P1[Logger]
        P2[Common Utilities]
        P3[Vault Integration]
        P4[RKE2 Server Logic]
        P5[Load Balancer Handler]
    end

    A --> P1
    B --> P2
    D --> P3
    G --> P4
    M --> P5
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