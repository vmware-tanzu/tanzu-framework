# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

ROOT_DIR := $(shell git rev-parse --show-toplevel)
MODULE_ROOT_DIR := $(ROOT_DIR)/cli/runtime

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

NUL = /dev/null
ifeq ($(GOHOSTOS),windows)
	NUL = NUL
endif

BUILD_SHA ?= $$(git describe --match=$(git rev-parse --short HEAD) --always --dirty)
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
BUILD_VERSION ?= $(shell git describe --tags --abbrev=0 2>$(NUL))

ifeq ($(strip $(BUILD_VERSION)),)
BUILD_VERSION = dev
endif


# Directories
TOOLS_DIR := $(abspath $(MODULE_ROOT_DIR)/hack/tools)
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

# Add tooling binaries here and in hack/tools/Makefile
GOIMPORTS        := $(TOOLS_BIN_DIR)/goimports
CONTROLLER_GEN   := $(TOOLS_BIN_DIR)/controller-gen
TOOLING_BINARIES := $(GOIMPORTS) $(CONTROLLER_GEN)

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


.PHONY: configure-version
configure-version: ## Configure plugin runtime version
	# Update plugin runtime version variable in version pkg
	sed "s+TANZU_PLUGIN_RUNTIME_VERSION+${BUILD_VERSION}+g"  hack/update-bundled-plugin-runtime-version/update-bundled-plugin-runtime-version.txt  > version/zz_generated_plugin_runtime_version.go

## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: fmt vet ## Run Tests
	${GO} test ./... -timeout 60m -race -coverprofile coverage.txt -v

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

## --------------------------------------
## Generators
## --------------------------------------

CONTROLLER_GEN_SRC ?= "./..."

generate-controller-code: $(CONTROLLER_GEN) $(GOIMPORTS) ## Generate code via controller-gen
	$(CONTROLLER_GEN) $(GENERATOR) object:headerFile="$(MODULE_ROOT_DIR)/hack/boilerplate.go.txt",year=$(shell date +%Y) paths="$(CONTROLLER_GEN_SRC)" $(OPTIONS)
	$(MAKE) fmt

generate-manifests:
	$(MAKE) generate-controller-code GENERATOR=crd OPTIONS="output:crd:artifacts:config=$(MODULE_ROOT_DIR)/generated/crd/bases" CONTROLLER_GEN_SRC=$(CONTROLLER_GEN_SRC)
