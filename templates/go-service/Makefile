# Per-deployable targets. Same names in every deployable (docs/standards/engineering.md).
# Must run on GNU Make 3.81 (macOS): no .ONESHELL, no ::=, no != .
SHELL := bash
COVERAGE_MIN := 80
GOLANGCI_CONFIG ?= ../../.golangci.yml
IMAGE ?= harness-engineering-__NAME__:local
# Baked into the image and logged as the version attribute; CI passes the git SHA.
VERSION ?= dev

# Never let a Go command download a toolchain implicitly.
export GOTOOLCHAIN := local

.PHONY: lint test build image

lint:
	@test -z "$$(gofmt -l .)" || { gofmt -l .; echo "gofmt: format the files above"; exit 1; }
	go vet ./...
	golangci-lint run --config $(GOLANGCI_CONFIG) ./...

test:
	go test -race -coverprofile=coverage.out ./internal/...
	@set -o pipefail; go tool cover -func=coverage.out | awk -v min=$(COVERAGE_MIN) \
	  '/^total:/ { sub("%","",$$3); if ($$3+0 < min) { printf "coverage %s%% < %s%%\n", $$3, min; exit 1 } \
	  printf "coverage %s%% (min %s%%)\n", $$3, min }'

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/service ./cmd/service

image:
	docker build --build-arg VERSION=$(VERSION) -t $(IMAGE) .
