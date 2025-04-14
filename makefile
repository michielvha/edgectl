# Variables
GO_RUN=go run .
CLI=edgectl

# Default target
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make server      		    Run 'rke2 server install'"
	@echo "  make server-join           Run 'rke2 server install --cluster-id rke2-03db202f'"
	@echo "  make agent       			Run 'rke2 agent --cluster-id rke2-03db202f'"
	@echo "  make lb-create              Run 'rke2 lb create --cluster-id rke2-03db202f'"
	@echo "  make lb-status              Run 'rke2 lb status --cluster-id rke2-03db202f'"
	@echo "  make purge        		    Run 'rke2 purge'"
	@echo "  make config                Run 'rke2 config'"
	@echo "  make test func             Test a Go function with a sample input"
	@echo "  make clean                 Remove temporary files (optional)"

# Commands
.PHONY: server
server:
	$(GO_RUN) rke2 server install --vip 172.16.12.232

.PHONY: server-join
server-join:
	$(GO_RUN) rke2 server install --cluster-id $(CLUSTER_ID)

.PHONY: server-test
server-verify:
	sudo cat /etc/rancher/rke2/config.yaml # | grep 'token:'

# Commands
.PHONY: agent
agent:
	$(GO_RUN) rke2 agent --cluster-id $(CLUSTER_ID)

# Load balancer commands
.PHONY: lb-create
lb-create:
	$(GO_RUN) rke2 lb create --cluster-id $(CLUSTER_ID) --vip 172.16.12.232

.PHONY: lb-status
lb-status:
	$(GO_RUN) rke2 lb status --cluster-id $(CLUSTER_ID)

.PHONY: purge
purge:
	$(GO_RUN) rke2 purge

.PHONY: config
config:
	$(GO_RUN) rke2 config --cluster-id $(CLUSTER_ID)

.PHONY: config
status:
	$(GO_RUN) rke2 status

.PHONY: test-func
test func:
	@echo "🔍 Testing individual function..."
	go run ./cmd/debug/test.go

.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	rm -rf output.log temp/