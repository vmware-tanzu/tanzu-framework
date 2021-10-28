# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

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

# BUILD_EDITION is the Tanzu Edition, the plugin should be built for.
# Valid values for BUILD_EDITION are 'tce' and 'tkg'. Default value of BUILD_EDITION is 'tkg'.
ifneq ($(BUILD_EDITION), tce)
BUILD_EDITION = tkg
endif

LD_FLAGS = -s -w
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Date=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.SHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Version=$(BUILD_VERSION)'

# Add supported OS-ARCHITECTURE combinations here
ENVS := linux-amd64 windows-amd64 darwin-amd64
STANDALONE_PLUGINS := login management-cluster package
CONTEXTAWARE_PLUGINS := cluster kubernetes-release pinniped-auth secret
PLUGINS := $(STANDALONE_PLUGINS) $(CONTEXTAWARE_PLUGINS)