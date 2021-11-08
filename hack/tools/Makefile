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
ROOT_DIR:=$(shell git rev-parse --show-toplevel)
BIN_DIR := bin
TOOLS_SCRIPTS_DIR := $(ROOT_DIR)/hack/tools
SRCS := go.mod go.sum

YQ_VERSION=v4.5.0
KUBEBUILDER_VERSION=2.3.1
YTT_VERSION=v0.35.1
KBLD_VERSION=v0.31.0
KAPP_VERSION=v0.42.0
VENDIR_VERSION=v0.23.0
KUBEVAL_VERSION=0.15.0
VALE_VERSION=2.10.3

# Host information.
HOST_OS=$(shell go env GOOS)
HOST_ARCH=$(shell go env GOARCH)

# Binaries.
GINKGO             := $(BIN_DIR)/ginkgo
GOLANGCI_LINT      := $(BIN_DIR)/golangci-lint
GOIMPORTS          := $(BIN_DIR)/goimports
KUBEBUILDER        := $(BIN_DIR)/kubebuilder
KUBEVAL            := $(BIN_DIR)/kubeval
GOBINDATA          := $(BIN_DIR)/gobindata
VALE               := $(BIN_DIR)/vale
YTT                := $(BIN_DIR)/ytt
KBLD               := $(BIN_DIR)/kbld
VENDIR             := $(BIN_DIR)/vendir
IMGPKG             := $(BIN_DIR)/imgpkg
YQ                 := $(BIN_DIR)/yq
TRIVY              := $(BIN_DIR)/trivy
KAPP               := $(BIN_DIR)/kapp

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Binaries
## --------------------------------------

ginkgo: $(GINKGO)
$(GINKGO): go.mod
	mkdir -p $(BIN_DIR)
	go build -tags=tools -o $(BIN_DIR)/ginkgo github.com/onsi/ginkgo/ginkgo

COUNTERFEITER := $(BIN_DIR)/counterfeiter
$(COUNTERFEITER): $(BIN_DIR) go.mod go.sum
	go build -tags=tools -o $@ github.com/maxbrunsfeld/counterfeiter/v6

golangci-lint: $(GOLANGCI_LINT) $(SRCS) ## Build golangci-lint
$(GOLANGCI_LINT): go.mod
	go build -tags=tools -o $@ github.com/golangci/golangci-lint/cmd/golangci-lint

goimports: $(GOIMPORTS) $(SRCS) ## Build goimports
$(GOIMPORTS): go.mod
	go build -tags=tools -o $@ golang.org/x/tools/cmd/goimports

kubebuilder: $(KUBEBUILDER) ## Install kubebuilder
$(KUBEBUILDER):
	mkdir -p $(BIN_DIR)
	curl -L https://github.com/kubernetes-sigs/kubebuilder/releases/download/v${KUBEBUILDER_VERSION}/kubebuilder_${KUBEBUILDER_VERSION}_${HOST_OS}_${HOST_ARCH}.tar.gz | tar -xz -C /tmp/
	mv /tmp/kubebuilder_$(KUBEBUILDER_VERSION)_$(HOST_OS)_$(HOST_ARCH) $(@)

kubeval: $(KUBEVAL) ## Install kubeval
$(KUBEVAL):
	mkdir -p $(BIN_DIR)
	curl -sL https://github.com/instrumenta/kubeval/releases/download/$(KUBEVAL_VERSION)/kubeval-$(HOST_OS)-$(HOST_ARCH).tar.gz  | tar -xz -C /tmp/
	mv /tmp/kubeval $(@)
	chmod a+x $(@)

gobindata: $(GOBINDATA)
$(GOBINDATA): go.mod # Build go-bindata
	mkdir -p $(BIN_DIR)
	go build -tags=tools -o $(BIN_DIR) github.com/shuLhan/go-bindata/...; mv $(BIN_DIR)/go-bindata $(GOBINDATA)

vale: $(VALE)
$(VALE):
	mkdir -p $(BIN_DIR)
	curl -sfL https://install.goreleaser.com/github.com/ValeLint/vale.sh | sh -s v$(VALE_VERSION)

yq: $(YQ) ## Download yq
$(YQ):
	mkdir -p $(BIN_DIR)
	echo $(HOST_OS) $(HOST_ARCH)
	curl -sL https://github.com/mikefarah/yq/releases/download/$(YQ_VERSION)/yq_$(HOST_OS)_$(HOST_ARCH) -o $(@) && \
	chmod a+x $(@)

## --------------------------------------------------------------------------------------------------------
## Carvel tools
## --------------------------------------------------------------------------------------------------------

ytt: $(YTT) ## Install ytt
$(YTT):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-ytt/releases/download/$(YTT_VERSION)/ytt-$(HOST_OS)-$(HOST_ARCH)
	mv ytt-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

kbld: $(KBLD) ## Download kbld
$(KBLD):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-kbld/releases/download/$(KBLD_VERSION)/kbld-$(HOST_OS)-$(HOST_ARCH)
	mv kbld-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

vendir: $(VENDIR) ## Download vendir
$(VENDIR):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-vendir/releases/download/$(VENDIR_VERSION)/vendir-$(HOST_OS)-$(HOST_ARCH)
	mv vendir-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

imgpkg: $(IMGPKG) ## Download imgpkg
$(IMGPKG):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-imgpkg/releases/download/v0.11.0/imgpkg-$(HOST_OS)-$(HOST_ARCH)
	mv imgpkg-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

kapp: $(KAPP) ## Download kapp
$(KAPP):
	mkdir -p $(BIN_DIR)
	curl -LO https://github.com/vmware-tanzu/carvel-kapp/releases/download/$(KAPP_VERSION)/kapp-$(HOST_OS)-$(HOST_ARCH)
	mv kapp-$(HOST_OS)-$(HOST_ARCH) $(@)
	chmod a+x $(@)

## --------------------------------------
## Trivy
## --------------------------------------

trivy: $(TRIVY) ## Download trivy
$(TRIVY):
	curl -sfL https://raw.githubusercontent.com/aquasecurity/trivy/main/contrib/install.sh | sh -s -- -b $(BIN_DIR)

## --------------------------------------
## Cleanup
## --------------------------------------

.PHONY: clean
clean: ## Remove all generated binaries
	rm -rf bin
