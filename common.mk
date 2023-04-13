# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

ROOT_DIR := $(shell git rev-parse --show-toplevel)
RELATIVE_ROOT ?= .
CONTROLLER_GEN_SRC ?= "./..."

# Framework has lots of components, and many build steps that are disparate.
# Use docker buildkit so that we get faster image builds and skip redundant work.
# TODO: We need to measure what this speeds up, and what it doesn't (and why).
export DOCKER_BUILDKIT := 1

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

ifdef DEBUG
LD_FLAGS = -s
GC_FLAGS = all=-N -l
else
LD_FLAGS = -s -w
GC_FLAGS =
endif

# Set buildinfo for modules
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.Date=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.SHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/util/buildinfo.Version=$(BUILD_VERSION)'

# Add supported OS-ARCHITECTURE combinations here
ENVS ?= linux-amd64 windows-amd64 darwin-amd64

# Hosts running SELinux need :z added to volume mounts
SELINUX_ENABLED := $(shell cat /sys/fs/selinux/enforce 2> /dev/null || echo 0)

ifeq ($(SELINUX_ENABLED),1)
  DOCKER_VOL_OPTS?=:z
endif

# Support DISTROLESS_BASE_IMAGE,GOPROXY,GOPROXY change while building image
GOPROXY ?= "https://proxy.golang.org,direct"
GOPROXY ?= "sum.golang.org"
DISTROLESS_BASE_IMAGE ?= gcr.io/distroless/static:nonroot

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
##@ Tooling Binaries
## --------------------------------------

.PHONY: tools
tools: $(TOOLING_BINARIES) ## Build tooling binaries
.PHONY: $(TOOLING_BINARIES)
$(TOOLING_BINARIES):
	make -C $(TOOLS_DIR) $(@F)

## --------------------------------------
##@ API/controller building and generation
## --------------------------------------

.PHONY: generate-controller-code
generate-controller-code: $(CONTROLLER_GEN) $(GOIMPORTS) ## Generate code via controller-gen
	$(CONTROLLER_GEN) $(GENERATOR) object:headerFile="$(ROOT_DIR)/hack/boilerplate.go.txt",year=$(shell date +%Y) paths="$(CONTROLLER_GEN_SRC)" $(OPTIONS)

.PHONY: generate-manifests
generate-manifests:
	$(MAKE) generate-controller-code GENERATOR=crd OPTIONS="output:crd:artifacts:config=$(RELATIVE_ROOT)/config/crd/bases" CONTROLLER_GEN_SRC=$(CONTROLLER_GEN_SRC)
