# Single entry point for developers and CI.
# Must run on macOS, Linux and Termux Make; avoid GNU Make 3.82+ features for macOS compatibility.
SHELL := bash
.DEFAULT_GOAL := help

COVERAGE_MIN := 80
HARNESS_DIR := tools/harness
ROOT := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

# Never let a Go command download a toolchain implicitly; developers install Go themselves.
export GOTOOLCHAIN := local

.PHONY: help check harness harness-test harness-lint spec-validate doctor template-test new-service \
	up down ps logs compose-config secrets-scan vuln image-scan infra-images deployables deploy rename

help: ## List targets
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  %-16s %s\n", $$1, $$2}'

check: harness harness-test harness-lint spec-validate template-test ## Run every repository validation
	@echo "make check: all validations passed"

harness: ## Validate repo structure, registry, catalog, agents, docs
	@go -C $(HARNESS_DIR) run ./cmd/harness check -root $(CURDIR)

harness-test: ## Harness unit tests with coverage gate
	@go -C $(HARNESS_DIR) test -race -coverprofile=coverage.out ./internal/...
	@set -o pipefail; go -C $(HARNESS_DIR) tool cover -func=coverage.out | awk -v min=$(COVERAGE_MIN) \
	  '/^total:/ { sub("%","",$$3); if ($$3+0 < min) { printf "coverage %s%% < %s%%\n", $$3, min; exit 1 } \
	  printf "coverage %s%% (min %s%%)\n", $$3, min }'

harness-lint: ## golangci-lint on the harness
	@cd $(HARNESS_DIR) && golangci-lint run --config $(CURDIR)/.golangci.yml ./...

spec-validate: ## OpenSpec strict validation of all changes and specs
	@openspec validate --all --strict --no-interactive

doctor: ## Check required developer tools are present (runs the diagnostic-only scripts/bootstrap.sh)
	@$(ROOT)scripts/bootstrap.sh

# Local stack and delivery. Every Compose run merges the base file, one overlay and each
# deployable fragment; paths are relative to the repository root.
COMPOSE_ENV := infra/env/local/compose.env
COMPOSE_FRAGMENTS := $(wildcard services/*/compose.yaml) $(wildcard apps/*/compose.yaml)
# Socket group of the host's Docker daemon: the socket proxy runs as non-root and joins this group (override if the
# socket group differs, for example on Docker Desktop).
DOCKER_GID ?= $(shell stat -L -c %g /var/run/docker.sock 2>/dev/null || echo 999)
COMPOSE := DOCKER_GID=$(DOCKER_GID) docker compose --project-directory . --env-file $(COMPOSE_ENV) -f infra/compose/compose.yaml -f infra/compose/compose.local.yaml $(foreach f,$(COMPOSE_FRAGMENTS),-f $(f))
COMPOSE_DEPLOY := DOCKER_GID=$(DOCKER_GID) docker compose --project-directory . --env-file $(COMPOSE_ENV) -f infra/compose/compose.yaml -f infra/compose/compose.deploy.yaml $(foreach f,$(COMPOSE_FRAGMENTS),-f $(f))
BASE ?= origin/main

# Scanners run in CI only (not in check): pinned by digest, tag kept for humans.
GITLEAKS_IMAGE ?= zricethezav/gitleaks:v8.30.1@sha256:c00b6bd0aeb3071cbcb79009cb16a60dd9e0a7c60e2be9ab65d25e6bc8abbb7f
TRIVY_IMAGE ?= aquasec/trivy:0.76.2@sha256:8f6bb947dbc528902e11f2f5be52f7b519f6a39d5d148f41166f2040499768fb
GOVULNCHECK_VERSION ?= v1.9.0
GO_MODULES := tools/harness $(patsubst %/go.mod,%,$(wildcard services/*/go.mod apps/*/go.mod))

up: ## Build and start the local stack, wait until healthy (needs Docker; port 80 free)
	@$(COMPOSE) up -d --build --wait

down: ## Stop the local stack (volumes are kept)
	@$(COMPOSE) down

ps: ## Show local stack status
	@$(COMPOSE) ps

logs: ## Follow local stack logs (SERVICE=<name> for one service)
	@$(COMPOSE) logs -f --tail=100 $(SERVICE)

compose-config: ## Validate the local and deploy Compose merges (needs Docker)
	@$(COMPOSE) config -q
	@$(COMPOSE_DEPLOY) config -q
	@echo "compose-config: local and deploy merges are valid"

secrets-scan: ## gitleaks over the full git history (CI; needs Docker and fetch-depth 0)
	@docker run --rm -v "$(CURDIR):/repo:ro" $(GITLEAKS_IMAGE) detect --source /repo \
	  --config /repo/infra/gitleaks.toml --redact --no-banner --exit-code 1

vuln: ## govulncheck on tools/harness and every Go deployable (CI)
	@set -e; for d in $(GO_MODULES); do \
	  echo "govulncheck $$d"; \
	  (cd $$d && GOTOOLCHAIN=local go run golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION) ./...); \
	done

image-scan: ## Trivy scan of IMAGE: fails on fixable HIGH/CRITICAL CVEs (CI; usage: make image-scan IMAGE=<image>)
	@test -n "$(IMAGE)" || { echo "usage: make image-scan IMAGE=<image>"; exit 2; }
	@docker run --rm -v /var/run/docker.sock:/var/run/docker.sock $(TRIVY_IMAGE) image \
	  --exit-code 1 --severity HIGH,CRITICAL --ignore-unfixed --no-progress "$(IMAGE)"

infra-images: ## List the third-party images pinned in infra/compose (CI reports their CVEs, not gated)
	@grep -h '^    image:' infra/compose/compose.yaml infra/compose/compose.local.yaml infra/compose/compose.deploy.yaml \
	  | awk '{print $$2}' | sort -u

deployables: ## Deployable paths as a JSON array: changed since BASE, or all with ALL=1 (CI matrix)
	@set -eu; \
	all=$$(ls -d services/*/ apps/*/ 2>/dev/null | sed 's:/$$::'); \
	if [ "$(ALL)" = 1 ]; then list="$$all"; \
	else \
	  changed=$$(git diff --name-only "$(BASE)...HEAD"); \
	  if printf '%s\n' "$$changed" | grep -Eq '^(templates|packages|contracts)/'; then list="$$all"; \
	  else \
	    list=""; \
	    for d in $$(printf '%s\n' "$$changed" | sed -n -E 's#^((services|apps)/[^/]+)/.*#\1#p' | sort -u); do \
	      if [ -d "$$d" ]; then list="$$list $$d"; fi; \
	    done; \
	  fi; \
	fi; \
	printf '%s\n' $$list | awk 'NF { p = p (p == "" ? "" : ",") "\"" $$0 "\"" } END { print "[" p "]" }'

deploy: ## Deploy ENV=dev|prod to HOST (CI calls this; inputs and host-key setup in infra/README.md)
	@test -n "$(ENV)" -a -n "$(HOST)" -a -n "$(SSH_USER)" -a -n "$(ENV_FILE)" -a -n "$(SECRETS_FILE)" -a -n "$(KNOWN_HOSTS)" || { echo "usage: make deploy ENV=dev|prod HOST=<host> SSH_USER=<user> ENV_FILE=<file> SECRETS_FILE=<file> KNOWN_HOSTS=<file> [SSH_KEY=<file>]"; exit 2; }
	@sh infra/deploy/deploy.sh "$(ENV)" "$(HOST)" "$(SSH_USER)" "$(ENV_FILE)" "$(SECRETS_FILE)" "$(KNOWN_HOSTS)" "$(SSH_KEY)"

new-service: ## Generate services/<name> from templates/go-service (name=<kebab-case>)
	@test -n "$(name)" || { echo "usage: make new-service name=<kebab-case>"; exit 2; }
	@./scripts/new-service.sh "$(name)"

rename: ## Rename the project slug in every file, once per new product repository (slug=<kebab-case>)
	@test -n "$(slug)" || { echo "usage: make rename slug=<kebab-case>"; exit 2; }
	@./scripts/rename.sh "$(slug)"

template-test: ## Generate a demo service in a temporary root and run its lint and test targets
	@tmp=$$(mktemp -d) && trap 'rm -rf "$$tmp"' EXIT && \
	./scripts/new-service.sh demo "$$tmp" >/dev/null && \
	$(MAKE) -C "$$tmp/services/demo" lint test GOLANGCI_CONFIG=$(CURDIR)/.golangci.yml
