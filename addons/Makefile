# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help

ROOT_DIR	:= $(dir $(shell pwd))

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE := on

# Image URL to use all building/pushing image targets
IMG ?= addons-controller:$(shell git describe --always --dirty --tags)

# Directories
BIN_DIR       := bin
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
KUBEBUILDER_BIN_DIR := $(TOOLS_BIN_DIR)/kubebuilder/bin

export PATH := $(abspath $(BIN_DIR)):$(PATH)

# Binaries
MANAGER       := $(BIN_DIR)/manager

# ssh private key file name
SSH_PRIVATE_KEY ?= .private_ssh_key

.PHONY: all
all: test manager ## Tests and builds the binaries

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: fmt vet ## Run Tests
	$(MAKE) kubebuilder -C ${ROOT_DIR}$(TOOLS_DIR)
	KUBEBUILDER_ASSETS=${ROOT_DIR}$(KUBEBUILDER_BIN_DIR) go test ./... -coverprofile cover.out -v 2

.PHONY: test-verbose
test-verbose: ## Verbose tests with streaming output for debugging
	ginkgo -v ./...

## --------------------------------------
## Binaries
## --------------------------------------

.PHONY: manager-binary
manager-binary: $(MANAGER) ## Build manager binary
$(MANAGER):
	go build -o $@ -ldflags '-extldflags "-static"'

.PHONY: manager
manager: fmt vet manager-binary ## Build manager binary

## --------------------------------------
## Linting and fixing linter errors
## --------------------------------------

.PHONY: fmt
fmt: ## Fmt codebase
	go fmt ./...

.PHONY: vet
vet: ## Vet codebase
	go vet ./...

## --------------------------------------
## Development - local
## --------------------------------------

.PHONY: run
run: fmt vet ## Run locally
	go run ./main.go

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-build
docker-build: ## Build the docker image
	@[ -f ../$(SSH_PRIVATE_KEY) ] || { echo "an SSH key file named $(SSH_PRIVATE_KEY) must exist and have read access to github.com" 1>&2; exit 1; }
	cd .. && docker build -t $(IMG) -f addons/Dockerfile --build-arg SSH_PRIVATE_KEY=$(SSH_PRIVATE_KEY) .

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(IMG)

.PHONY: docker-remove
docker-remove: ## Remove the docker image
	docker rmi $(IMG) 2>/dev/null || true

## --------------------------------------
## Development - git
## --------------------------------------

.PHONY:	git-config
git-config: ## Configure gitlab access
	git config --global url.git@github.com:.insteadof https://github.com/

## --------------------------------------
## Clean
## --------------------------------------

.PHONY: clean ## Remove all generated files
clean: docker-remove
	rm -rf bin
