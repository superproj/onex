# ==============================================================================
# Makefile helper functions for lintes
#

## Tool Binaries
GOLANGCI_LINT := golangci-lint
KUBE_LINT := kube-linter

.PHONY: lint.run
lint.run: lint.ci lint.kubefiles lint.dockerfiles lint.charts ## Run all available linters.

.PHONY: lint.ci
lint.ci: lint.golangci-lint lint.onex ## Run CI-related linters.

.PHONY: lint.golangci-lint
lint.golangci-lint: tools.verify.golangci-lint ## Run golangci to lint source codes.
	@echo "===========> Run golangci to lint source codes"
	@$(GOLANGCI_LINT) run -c $(ONEX_ROOT)/.golangci.yaml $(ONEX_ROOT)/...

.PHONY: lint.onex
lint.onex: ## Run linters developed by onex developers.
	@$(GO) run cmd/lint-kubelistcheck/main.go $(ONEX_ROOT)/...

.PHONY: lint.kubefiles
lint.kubefiles: tools.verify.kube-linter ## Lint protobuf files.
	@$(KUBE_LINT) lint $(ONEX_ROOT)/deployments

.PHONY: lint.dockerfiles 
lint.dockerfiles: image.verify go.build.verify ## Lint dockerfiles.
	@$(SCRIPTS_DIR)/ci-lint-dockerfiles.sh $(HADOLINT_VER) $(HADOLINT_FAILURE_THRESHOLD)

.PHONY: lint.charts
lint.charts: tools.verify.helm ## Lint helm charts.
	$(MAKE) chart.lint

# In actual development, many logs are difficult to follow the logcheck specifications and also do not 
# need to. Here we only have a basic understanding, and it is not recommended to use lint.logcheck rule.
.PHONY: lint.logcheck
lint.logcheck: tools.verify.logcheck ## Tool to check logging calls.
	@logcheck -check-contextual $(ONEX_ROOT)/...
	@logcheck -check-structured $(ONEX_ROOT)/...

