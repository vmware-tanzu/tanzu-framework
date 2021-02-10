# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Build tooling that is used for the main project

# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL := /usr/bin/env bash

.DEFAULT_GOAL := help

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq (,$(strip $(GOPROXY)))
GOPROXY := https://proxy.golang.org
endif
export GOPROXY

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE := on

# Directories.
BIN_DIR := bin
SRCS := go.mod go.sum

KUBEBUILDER_VERSION=2.3.1

# Host information.
HOST_OS=$(shell go env GOOS)
HOST_ARCH=$(shell go env GOARCH)


# Binaries.
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOLINT := $(BIN_DIR)/golint
KUBEBUILDER        := $(BIN_DIR)/kubebuilder

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
## --------------------------------------
## Binaries
## --------------------------------------

golangci-lint: $(GOLANGCI_LINT) $(SRCS) ## Build golangci-lint
$(GOLANGCI_LINT): go.mod
	go build -tags=tools -o $@ github.com/golangci/golangci-lint/cmd/golangci-lint

golint: $(GOLINT) $(SRC) ## Build golint
$(GOLINT): go.mod
	go build -tags=tools -o $@ golang.org/x/lint/golint

kubebuilder: $(KUBEBUILDER) ## Install kubebuilder
$(KUBEBUILDER):
	curl -sL https://go.kubebuilder.io/dl/$(KUBEBUILDER_VERSION)/$(HOST_OS)/$(HOST_ARCH) | tar -xz -C /tmp/
	mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(HOST_OS)_$(HOST_ARCH) $(@)

## --------------------------------------
## Cleanup
## --------------------------------------

.PHONY: clean
clean: ## Remove all generated binaries
	rm -rf bin