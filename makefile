# Variables
GO_RUN=go run .
CLI=edgectl

# Default target
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make build                 Build the edgectl binary"
	@echo "  make server      		    Run 'rke2 server install'"
	@echo "  make server-join           Run 'rke2 server install --cluster-id rke2-03db202f'"
	@echo "  make agent       			Run 'rke2 agent --cluster-id rke2-03db202f'"
	@echo "  make lb-create             Run 'rke2 lb create --cluster-id rke2-03db202f'"
	@echo "  make lb-status             Run 'rke2 lb status --cluster-id rke2-03db202f'"
	@echo "  make purge        		    Run 'rke2 purge'"
	@echo "  make config                Run 'rke2 config'"
	@echo "  make test                  Run all unit tests"
	@echo "  make test-cover            Run unit tests with coverage report"
	@echo "  make test-integration      Run integration tests (requires Docker)"
	@echo "  make test-func             Test a Go function with a sample input"
	@echo "  make clean                 Remove temporary files (optional)"

# Build
.PHONY: build
build:
	go build -o $(CLI) .

# Commands
.PHONY: server
server:
	$(GO_RUN) rke2 server install --vip 172.16.12.232

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
	$(GO_RUN) rke2 lb create --cluster-id $(CLUSTER_ID) --vip 172.16.12.232

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