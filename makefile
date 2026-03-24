# Variables
GO_RUN=go run .
CLI=edgectl
VIP=100.111.180.80

# Default target
.PHONY: help build server server-join agent lb-create lb-status purge config status test test-cover test-integration test-func clean lint
.PHONY: k3s-server k3s-server-join k3s-agent k3s-lb-create k3s-lb-status k3s-lb-cleanup k3s-purge k3s-config k3s-status k3s-bash
help:
	@echo "Usage:"
	@echo "  make build                 Build the edgectl binary"
	@echo ""
	@echo "  RKE2 Commands:"
	@echo "  make server                Run 'rke2 server install'"
	@echo "  make server-join           Run 'rke2 server install --cluster-id <id>'"
	@echo "  make agent                 Run 'rke2 agent install --cluster-id <id>'"
	@echo "  make lb-create             Run 'rke2 lb create --cluster-id <id>'"
	@echo "  make lb-status             Run 'rke2 lb status --cluster-id <id>'"
	@echo "  make purge                 Run 'rke2 system purge'"
	@echo "  make config                Run 'rke2 system kubeconfig --cluster-id <id>'"
	@echo "  make status                Run 'rke2 system status'"
	@echo ""
	@echo "  K3s Commands:"
	@echo "  make k3s-server            Run 'k3s server install'"
	@echo "  make k3s-server-join       Run 'k3s server install --cluster-id <id>'"
	@echo "  make k3s-agent             Run 'k3s agent install --cluster-id <id>'"
	@echo "  make k3s-lb-create         Run 'k3s lb create --cluster-id <id>'"
	@echo "  make k3s-lb-status         Run 'k3s lb status --cluster-id <id>'"
	@echo "  make k3s-lb-cleanup        Run 'k3s lb cleanup --cluster-id <id>'"
	@echo "  make k3s-purge             Run 'k3s system purge'"
	@echo "  make k3s-config            Run 'k3s system kubeconfig --cluster-id <id>'"
	@echo "  make k3s-status            Run 'k3s system status'"
	@echo "  make k3s-bash              Run 'k3s system bash'"
	@echo ""
	@echo "  Testing & Tooling:"
	@echo "  make test                  Run all unit tests"
	@echo "  make test-cover            Run unit tests with coverage report"
	@echo "  make test-integration      Run integration tests (requires Docker)"
	@echo "  make test-func             Test a Go function with a sample input"
	@echo "  make clean                 Remove temporary files (optional)"
	@echo "  make lint                  Run linter with auto-fix"


# Build
.PHONY: build
build:
	go build -o $(CLI) .

# Commands
.PHONY: server
server:
	$(GO_RUN) rke2 server install --vip $(VIP)

.PHONY: server-join
server-join:
	$(GO_RUN) rke2 server install --cluster-id $(CLUSTER_ID)

.PHONY: server-verify
server-verify:
	sudo cat /etc/rancher/rke2/config.yaml # | grep 'token:'

# Commands
.PHONY: agent
agent:
	$(GO_RUN) rke2 agent install --cluster-id $(CLUSTER_ID)

# Load balancer commands
.PHONY: lb-create
lb-create:
	$(GO_RUN) rke2 lb create --cluster-id $(CLUSTER_ID) --vip $(VIP)

.PHONY: lb-status
lb-status:
	$(GO_RUN) rke2 lb status --cluster-id $(CLUSTER_ID)

.PHONY: purge
purge:
	$(GO_RUN) rke2 system purge

.PHONY: config
config:
	$(GO_RUN) rke2 system kubeconfig --cluster-id $(CLUSTER_ID)

.PHONY: status
status:
	$(GO_RUN) rke2 system status

# K3s Commands
.PHONY: k3s-server
k3s-server:
	$(GO_RUN) k3s server install --vip $(VIP)

.PHONY: k3s-server-join
k3s-server-join:
	$(GO_RUN) k3s server install --cluster-id $(CLUSTER_ID)

.PHONY: k3s-agent
k3s-agent:
	$(GO_RUN) k3s agent install --cluster-id $(CLUSTER_ID)

.PHONY: k3s-lb-create
k3s-lb-create:
	$(GO_RUN) k3s lb create --cluster-id $(CLUSTER_ID) --vip $(VIP)

.PHONY: k3s-lb-status
k3s-lb-status:
	$(GO_RUN) k3s lb status --cluster-id $(CLUSTER_ID)

.PHONY: k3s-lb-cleanup
k3s-lb-cleanup:
	$(GO_RUN) k3s lb cleanup --cluster-id $(CLUSTER_ID)

.PHONY: k3s-purge
k3s-purge:
	$(GO_RUN) k3s system purge

.PHONY: k3s-config
k3s-config:
	$(GO_RUN) k3s system kubeconfig --cluster-id $(CLUSTER_ID)

.PHONY: k3s-status
k3s-status:
	$(GO_RUN) k3s system status

.PHONY: k3s-bash
k3s-bash:
	$(GO_RUN) k3s system bash

.PHONY: test-func
test-func:
	@echo "🔍 Testing individual function..."
	go run ./cmd/debug/test.go

# Test targets
.PHONY: test
test:
	@echo "🧪 Running unit tests..."
	go test ./... -v

.PHONY: test-cover
test-cover:
	@echo "🧪 Running unit tests with coverage..."
	go test ./... -v -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report written to coverage.html"

.PHONY: test-integration
test-integration:
	@echo "🧪 Running integration tests (requires Docker)..."
	go test ./pkg/vault/ -tags=integration -v -count=1

.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	rm -rf output.log temp/

.PHONY: lint
lint:
	@echo "Running golangci-lint with auto-fix on backend..."
	@export PATH="$$(go env GOPATH)/bin:$$PATH" && \
	golangci-lint run --fix