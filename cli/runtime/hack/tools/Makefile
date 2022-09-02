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

# Directories.
ROOT_DIR:=$(shell git rev-parse --show-toplevel)
BIN_DIR := bin
SRCS := go.mod go.sum

# Binaries.
GOIMPORTS := $(BIN_DIR)/goimports

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Binaries
## --------------------------------------

goimports: $(GOIMPORTS) $(SRCS) ## Build goimports
$(GOIMPORTS): go.mod go.sum
	go build -tags=tools -o $@ golang.org/x/tools/cmd/goimports

## --------------------------------------
## Cleanup
## --------------------------------------

.PHONY: clean
clean: ## Remove all generated binaries
	rm -rf bin
