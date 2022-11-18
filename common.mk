# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

ROOT_DIR := $(shell git rev-parse --show-toplevel)
RELATIVE_ROOT ?= .
CONTROLLER_GEN_SRC ?= "./..."

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

# BUILD_EDITION is the Tanzu Edition, the plugin should be built for.
# Valid values for BUILD_EDITION are 'tce' and 'tkg'. Default value of BUILD_EDITION is 'tkg'.
ifneq ($(BUILD_EDITION), tce)
BUILD_EDITION = tkg
endif

ifdef DEBUG
LD_FLAGS = -s
GC_FLAGS = all=-N -l
else
LD_FLAGS = -s -w
GC_FLAGS =
endif

# Remove old package path for buildinfo vars when it's no longer in use by plugin binaries.
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Date=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.SHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Version=$(BUILD_VERSION)'

# Set buildinfo for plugins
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo.Date=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo.SHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/runtime/buildinfo.Version=$(BUILD_VERSION)'

# Set buildinfo for tkr, object-propagation and (potentially) other modules
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.Date=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.SHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.Version=$(BUILD_VERSION)'

# Add supported OS-ARCHITECTURE combinations here
ENVS ?= linux-amd64 windows-amd64 darwin-amd64
STANDALONE_PLUGINS ?= login management-cluster:k8s package:k8s pinniped-auth secret:k8s telemetry:k8s
CONTEXTAWARE_PLUGINS ?= cluster:k8s kubernetes-release:k8s feature:k8s
ADMIN_PLUGINS ?= builder codegen test
PLUGINS ?= $(STANDALONE_PLUGINS) $(CONTEXTAWARE_PLUGINS)

# Hosts running SELinux need :z added to volume mounts
SELINUX_ENABLED := $(shell cat /sys/fs/selinux/enforce 2> /dev/null || echo 0)

ifeq ($(SELINUX_ENABLED),1)
  DOCKER_VOL_OPTS?=:z
endif


# Directories
TOOLS_DIR := $(abspath $(ROOT_DIR)/hack/tools)
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin


# Add tooling binaries here and in hack/tools/Makefile
CONTROLLER_GEN     := $(TOOLS_BIN_DIR)/controller-gen
GOLANGCI_LINT      := $(TOOLS_BIN_DIR)/golangci-lint
GOIMPORTS          := $(TOOLS_BIN_DIR)/goimports
GOBINDATA          := $(TOOLS_BIN_DIR)/gobindata
KUBEBUILDER        := $(TOOLS_BIN_DIR)/kubebuilder
KUSTOMIZE          := $(TOOLS_BIN_DIR)/kustomize
YTT                := $(TOOLS_BIN_DIR)/ytt
KBLD               := $(TOOLS_BIN_DIR)/kbld
VENDIR             := $(TOOLS_BIN_DIR)/vendir
IMGPKG             := $(TOOLS_BIN_DIR)/imgpkg
KAPP               := $(TOOLS_BIN_DIR)/kapp
GINKGO             := $(TOOLS_BIN_DIR)/ginkgo
VALE               := $(TOOLS_BIN_DIR)/vale
YQ                 := $(TOOLS_BIN_DIR)/yq
CONVERSION_GEN     := $(TOOLS_BIN_DIR)/conversion-gen
COUNTERFEITER      := $(TOOLS_BIN_DIR)/counterfeiter
TOOLING_BINARIES   := $(CONTROLLER_GEN) $(GOLANGCI_LINT) $(YTT) $(KBLD) $(VENDIR) $(IMGPKG) $(KAPP) $(KUSTOMIZE) $(GOIMPORTS) $(GOBINDATA) $(GINKGO) $(VALE) $(YQ) $(CONVERSION_GEN) $(COUNTERFEITER)

## --------------------------------------
##@ API/controller building and generation
## --------------------------------------

help: ## Display this help (default)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m\033[32m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


## --------------------------------------
##@ Tooling Binaries
## --------------------------------------

tools: $(TOOLING_BINARIES) ## Build tooling binaries
.PHONY: $(TOOLING_BINARIES)
$(TOOLING_BINARIES):
	make -C $(TOOLS_DIR) $(@F)

generate-controller-code: $(CONTROLLER_GEN) $(GOIMPORTS) ## Generate code via controller-gen
	$(CONTROLLER_GEN) $(GENERATOR) object:headerFile="$(ROOT_DIR)/hack/boilerplate.go.txt",year=$(shell date +%Y) paths="$(CONTROLLER_GEN_SRC)" $(OPTIONS)
	$(MAKE) fmt

generate-manifests:
	$(MAKE) generate-controller-code GENERATOR=crd OPTIONS="output:crd:artifacts:config=$(RELATIVE_ROOT)/config/crd/bases" CONTROLLER_GEN_SRC=$(CONTROLLER_GEN_SRC)

fmt: $(GOIMPORTS) ## Run goimports
	$(GOIMPORTS) -w -local github.com/vmware-tanzu ./
