# Based on https://github.com/hashicorp/terraform-provider-scaffolding-framework/blob/main/GNUmakefile

# Detect OS and set shell/binary paths
ifeq ($(OS),Windows_NT)
	SHELL := pwsh.exe
	.SHELLFLAGS := -NoProfile -Command
	CUSTOM_GCL := ./bin/custom-gcl.exe
	RM_CMD := rm -Recurse -Force -ErrorAction SilentlyContinue
else
	CUSTOM_GCL := ./bin/custom-gcl
	RM_CMD := rm -rf
endif

default: fmt lint build test generate
ai: fmt lint build test

build:
	go build -v ./...

install: build
	go install -v ./...

lint: custom-gcl # custom GolangCi-Linter build containing our custom linter modules
	@echo "Running linters (includes custom linters: executewithretry, panichandler, & unknowncheck)"
	@echo "  For any false positives in the custom linters either ask AI to fix it at /custom-linters/<linter> (and add a test) or disable it in .golangci.yml"
	cd custom-linters && $(CUSTOM_GCL) run --allow-parallel-runners ./... # lint the custom linters
	./custom-linters/$(CUSTOM_GCL) run --allow-parallel-runners ./...

custom-gcl:
	@golangci-lint custom

lint-fix:
	cd custom-linters && $(CUSTOM_GCL) run --allow-parallel-runners --fix ./... # lint the custom linters
	./custom-linters/$(CUSTOM_GCL) run --allow-parallel-runners --fix ./...

generate:
	go generate ./...
	@echo "Running documentation validation..."
	@go run custom-linters/scripts/validate-docs/main.go

fmt:
	gofmt -s -w -e .

test:
	go test -short -cover -timeout=120s -parallel=10 ./...

#testacc:
#	TF_ACC=1 go test -cover -timeout 120m ./...

clean:
	go clean ./...
	go clean -cache -modcache
	golangci-lint cache clean
	$(RM_CMD) custom-linters/bin

help:
	@echo "Available targets:"
	@echo "  make           - Run fmt, lint, build, test, and generate (default)"
	@echo "  make build     - Build all packages"
	@echo "  make fmt       - Format all code with gofmt"
	@echo "  make generate  - Generate documentation"
	@echo "  make lint      - Run linters with golangci-lint and validate documentation"
	@echo "  make test      - Run all unit tests"
	@echo "  make testacc   - Run all acceptance tests"
	@echo "  make clean     - Clean build artifacts"
	@echo "  make help      - Show this help message"

.PHONY: default ai build lint custom-gcl generate fmt test testacc clean help