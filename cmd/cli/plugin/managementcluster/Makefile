# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

ROOT_DIR := $(shell git rev-parse --show-toplevel)
MODULE_ROOT_DIR := $(ROOT_DIR)/tkg

# Golang specific variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
GO := go

# Directories
TOOLS_DIR := $(abspath $(MODULE_ROOT_DIR)/hack/tools)
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

# Add tooling binaries here and in hack/tools/Makefile
GOIMPORTS        := $(TOOLS_BIN_DIR)/goimports
TOOLING_BINARIES := $(GOIMPORTS)

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help (default)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m\033[32m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## All
## --------------------------------------

.PHONY: all
all: test ## Tests the library

## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: fmt vet ## Run Tests
	${GO} test -timeout 60m -coverprofile coverage.txt -v `go list ./...`

.PHONY: vet
vet: ## Vet codebase
	${GO} vet ./...

fmt: $(GOIMPORTS) ## Run goimports
	$(GOIMPORTS) -w -local github.com/vmware-tanzu ./

## --------------------------------------
## Tooling Binaries
## --------------------------------------

tools: $(TOOLING_BINARIES) ## Build tooling binaries
.PHONY: $(TOOLING_BINARIES)
$(TOOLING_BINARIES):
	make -C $(TOOLS_DIR) $(@F)

