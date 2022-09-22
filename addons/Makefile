# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../common.mk

.DEFAULT_GOAL := help

ADDONS_DIR	:= $(dir $(shell pwd))

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE := on

# Image URL to use all building/pushing image targets
IMG_DEFAULT_NAME ?= addons-controller
IMG_DEFAULT_TAG := $(shell git describe --always --dirty --tags)
IMG_DEFAULT_NAME_TAG := $(IMG_DEFAULT_NAME):latest

IMG_VERSION_OVERRIDE ?= $(shell git describe --always --dirty --tags)

ifeq ($(strip $(OCI_REGISTRY)),)
	IMG ?= $(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
else
	IMG ?= $(OCI_REGISTRY)/$(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
endif

# Directories
BIN_DIR := bin
KUBEBUILDER_BIN_DIR := $(TOOLS_BIN_DIR)/kubebuilder/bin

export PATH := $(abspath $(BIN_DIR)):$(PATH)

# Binaries
MANAGER       := $(BIN_DIR)/manager

.PHONY: all
all: test manager ## Tests and builds the binaries

## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: fmt vet template-tests ## Run Tests
	$(MAKE) kubebuilder -C $(TOOLS_DIR)
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_BIN_DIR) go test ./... -timeout 60m -race -coverprofile coverage.txt -v 2

.PHONY: test-verbose
test-verbose: ## Verbose tests with streaming output for debugging
	$(MAKE) kubebuilder -C $(TOOLS_DIR)
	KUBEBUILDER_ASSETS=$(KUBEBUILDER_BIN_DIR) $(TOOLS_DIR)/bin/ginkgo -v ./... -timeout 60m -coverprofile coverage.txt -v 2

.PHONY: template-tests
template-tests: ## Sanity tests to verify the addon config CR templates
	./config/hack/test.sh verifyAllAddonConfigTemplates

## --------------------------------------
## Binaries
## --------------------------------------

.PHONY: manager-binary
manager-binary: $(MANAGER) ## Build manager binary
$(MANAGER):
	go build -o $@ -ldflags "$(LD_FLAGS) "'-extldflags "-static"'

.PHONY: manager
manager: fmt vet manager-binary ## Build manager binary

## --------------------------------------
## Linting and fixing linter errors
## --------------------------------------

.PHONY: vet
vet: ## Vet codebase
	go vet ./...

## --------------------------------------
## Development - local
## --------------------------------------

.PHONY: run
run: fmt vet ## Run locally
	go run -ldflags "$(LD_FLAGS)" ./main.go

## --------------------------------------
## Docker
## --------------------------------------

.PHONY: docker-build
docker-build: ## Build the docker image
	cd .. && docker build -t $(IMG) -f addons/Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" .

.PHONY: docker-push
docker-push: ## Push the docker image
	docker push $(IMG)

.PHONY: docker-publish ## Push the docker image
docker-publish: docker-push

.PHONY: docker-remove
docker-remove: ## Remove the docker image
	docker rmi $(IMG) 2>/dev/null || true

.PHONY: docker-build-and-publish
docker-build-and-publish: docker-build docker-publish kbld-image-replace

## --------------------------------------
## Update kbld-config
## --------------------------------------

.PHONY: kbld-image-replace
kbld-image-replace: ## Add newImage in kbld-config.yaml
	cd ../hack/packages/kbld-image-replace && \
	  go run main.go -kbld-config ../../../packages/addons-manager/kbld-config.yaml $(IMG_DEFAULT_NAME_TAG) $(IMG)

## --------------------------------------
## Clean
## --------------------------------------

.PHONY: clean ## Remove all generated files
clean: docker-remove
	rm -rf bin

TOOLS_BIN_DIR := hack/tools/bin
$(TOOLS_BIN_DIR):
	mkdir -p $(TOOLS_BIN_DIR)

COUNTERFEITER := $(TOOLS_BIN_DIR)/counterfeiter
$(COUNTERFEITER): $(TOOLS_BIN_DIR) go.mod go.sum ## Build counterfeiter
	cd hack/tools && $(GO) build -tags=tools -o bin/ github.com/maxbrunsfeld/counterfeiter/v6

generate-fakes: $(COUNTERFEITER) ## Generate fakes for unit tests
	PATH=$(PWD)/$(TOOLS_BIN_DIR):"$(PATH)" $(GO) generate ./...
	$(MAKE) fmt
