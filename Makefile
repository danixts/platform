SHELL := /bin/bash

GO         ?= go
GOFMT      ?= gofmt
GOIMPORTS  ?= goimports
GOLANGCI   ?= golangci-lint
MODULE     := $(shell $(GO) list -m)
PKGS       := ./...

.PHONY: help format imports tidy deps update lint vet check build test clean tools

help:
	@echo "Targets:"
	@echo "  format   - gofmt -s -w ."
	@echo "  imports  - goimports -local $(MODULE) -w ."
	@echo "  tidy     - go mod tidy"
	@echo "  deps     - go mod download"
	@echo "  update   - go get -u ./... && go mod tidy"
	@echo "  vet      - go vet ./..."
	@echo "  lint     - golangci-lint run ./..."
	@echo "  check    - fmt + imports + vet + lint + build"
	@echo "  build    - go build ./..."
	@echo "  test     - go test ./..."
	@echo "  clean    - go clean -cache -testcache"
	@echo "  tools    - install goimports and golangci-lint"

format:
	$(GOFMT) -s -w .

imports:
	$(GOIMPORTS) -local $(MODULE) -w .

tidy:
	$(GO) mod tidy

deps:
	$(GO) mod download

update:
	$(GO) get -u $(PKGS)
	$(GO) mod tidy

vet:
	$(GO) vet $(PKGS)

lint:
	$(GOLANGCI) run $(PKGS)

build:
	$(GO) build $(PKGS)

test:
	$(GO) test -race -count=1 $(PKGS)

clean:
	$(GO) clean -cache -testcache

check: fmt imports vet lint build

tools:
	$(GO) install golang.org/x/tools/cmd/goimports@latest
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
