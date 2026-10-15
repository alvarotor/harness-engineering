# Single entry point for developers and CI (D-0002, D-0013).
# Must run on the OS make of macOS (GNU Make 3.81) and Linux (P-0009): no .SHELLFLAGS, .ONESHELL or other 3.82+ features.
SHELL := bash
.DEFAULT_GOAL := help

COVERAGE_MIN := 80
HARNESS_DIR := tools/harness

# Use the tool versions pinned in .tool-versions (installed by scripts/bootstrap.sh), even if mise is not activated in the shell.
MISE_SHIMS := $(or $(MISE_DATA_DIR),$(HOME)/.local/share/mise)/shims
ifneq ($(wildcard $(MISE_SHIMS)),)
export PATH := $(MISE_SHIMS):$(PATH)
endif

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

doctor: ## Check tool versions against .tool-versions and Docker; points to scripts/bootstrap.sh on failure
	@fail=0; while read -r tool want; do \
	  [[ -z "$$tool" || "$$tool" == \#* ]] && continue; \
	  case "$$tool" in \
	    golang) have=$$(go env GOVERSION 2>/dev/null | sed 's/^go//') ;; \
	    nodejs) have=$$(node --version 2>/dev/null | sed 's/^v//') ;; \
	    opentofu) have=$$(tofu version 2>/dev/null | awk 'NR==1 {sub("v","",$$2); print $$2}') ;; \
	    buf) have=$$(buf --version 2>/dev/null) ;; \
	    golangci-lint) have=$$(golangci-lint version 2>/dev/null | awk '{print $$4}') ;; \
	    npm:@fission-ai/openspec) have=$$(openspec --version 2>/dev/null) ;; \
	    *) echo "?  $$tool: unknown to doctor"; continue ;; \
	  esac; \
	  if [[ "$$have" == "$$want" ]]; then echo "ok $$tool $$have"; \
	  else echo "!! $$tool: want $$want, have $${have:-missing}"; fail=1; fi; \
	done < .tool-versions; \
	if docker compose version >/dev/null 2>&1; then echo "ok docker compose"; \
	else echo "!! docker compose: missing or not running"; fail=1; fi; \
	if [[ $$fail == 1 ]]; then echo "fix: run ./scripts/bootstrap.sh (see CONTRIBUTING.md#prerequisites)"; fi; \
	exit $$fail
