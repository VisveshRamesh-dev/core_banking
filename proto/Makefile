.PHONY: api lint build test vet tools help

# Default target.
.DEFAULT_GOAL := help

# ── Proto code generation ─────────────────────────────────────────────────────

## api: generate all Go stubs from .proto contracts (gRPC + HTTP + Kratos errors)
##      Prerequisites: run `make tools` once to install the plugin binaries.
api:
	buf generate

## lint: lint all .proto files against buf's DEFAULT ruleset
lint:
	buf lint

# ── Go targets ────────────────────────────────────────────────────────────────

## build: compile all Go packages
build:
	go build ./...

## test: run the full test suite with the race detector enabled
test:
	go test -race ./...

## vet: run go vet across all packages
vet:
	go vet ./...

## fmt: format all Go source files
fmt:
	gofmt -w .

# ── Toolchain bootstrap ───────────────────────────────────────────────────────

## tools: install all code-generation binaries into GOPATH/bin
##        Run this once before the first `make api`.
tools:
	go install github.com/bufbuild/buf/cmd/buf@latest
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-http/v2@latest
	go install github.com/go-kratos/kratos/cmd/protoc-gen-go-errors/v2@latest

# ── Help ──────────────────────────────────────────────────────────────────────

## help: print this help message
help:
	@echo "Usage: make <target>"
	@echo ""
	@grep -E '^## ' Makefile | sed 's/^## /  /'
