# Variables
GO_RUN=go run .
CLI=edgectl

# Default target
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make install-server      Run 'rke2 server install --cluster-id rke2-03db202f'"
	@echo "  make install-agent       Run 'rke2 instal agent --cluster-id rke2-03db202f'"
	@echo "  make purge        		  Run 'rke2 purge'"
	@echo "  make config              Run 'rke2 config"
	@echo "  make test func           Test a Go function with a sample input"
	@echo "  make clean               Remove temporary files (optional)"

# Commands
.PHONY: install-server
install-server:
	$(GO_RUN) rke2 server install --cluster-id rke2-03db202f

# Commands
.PHONY: install-agent
install-agent:
	$(GO_RUN) rke2 agent install --cluster-id rke2-03db202f

.PHONY: purge
purge:
	$(GO_RUN) rke2 purge

.PHONY: config
list:
	$(GO_RUN) config kube --cluster-id rke2-03db202f

.PHONY: test-func
test func:
	@echo "🔍 Testing individual function..."
	go run ./cmd/debug/test.go

.PHONY: clean
clean:
	@echo "🧹 Cleaning up..."
	rm -rf output.log temp/