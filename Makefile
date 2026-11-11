# Single entry point for developers and CI (D-0002, D-0013).
# Must run on macOS, Linux and Termux Make; avoid GNU Make 3.82+ features for macOS compatibility (P-0012).
SHELL := bash
.DEFAULT_GOAL := help

COVERAGE_MIN := 80
HARNESS_DIR := tools/harness

# Never let a Go command download a toolchain implicitly; developers install Go themselves.
export GOTOOLCHAIN := local

.PHONY: help check harness harness-test harness-lint spec-validate doctor

help: ## List targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-16s %s\n", $$1, $$2}'

check: harness harness-test harness-lint spec-validate ## Run every repository validation
	@echo "make check: all validations passed"

harness: ## Validate repo structure, registry, catalog, agents, docs
	@go -C $(HARNESS_DIR) run ./cmd/harness check -root $(CURDIR)

harness-test: ## Harness unit tests with coverage gate
	@go -C $(HARNESS_DIR) test -race -coverprofile=coverage.out ./internal/...
	@set -o pipefail; go -C $(HARNESS_DIR) tool cover -func=coverage.out | awk -v min=$(COVERAGE_MIN) \
	  '/^total:/ { sub("%","",$$3); if ($$3+0 < min) { printf "coverage %s%% < %s%% (D-0017)\n", $$3, min; exit 1 } \
	  printf "coverage %s%% (min %s%%)\n", $$3, min }'

harness-lint: ## golangci-lint on the harness
	@cd $(HARNESS_DIR) && golangci-lint run --config $(CURDIR)/.golangci.yml ./...

spec-validate: ## OpenSpec strict validation of all changes and specs
	@openspec validate --all --strict --no-interactive

doctor: ## Check required developer tools are present; points to setup guidance on failure
	@fail=0; for tool in go node tofu buf golangci-lint openspec; do \
	  if command -v "$$tool" >/dev/null 2>&1; then echo "ok $$tool"; \
	  else echo "!! $$tool: missing"; fail=1; fi; \
	done; \
	if command -v docker >/dev/null 2>&1 && docker compose version >/dev/null 2>&1; then echo "ok docker compose"; \
	else echo "!! docker compose: missing"; fail=1; fi; \
	if [[ $$fail == 1 ]]; then echo "fix: see CONTRIBUTING.md#prerequisites; ./scripts/bootstrap.sh only reports missing tools"; fi; \
	exit $$fail
