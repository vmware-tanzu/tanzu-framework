# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help

ROOT_DIR	:= $(dir $(shell pwd))

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE := on

# Image URL to use all building/pushing image targets
IMG ?= addons-controller:latest
IMAGE_REPO ?= vmware.io

# Tag for template images, shoudld end with "-templates"
IMAGES_OUTPUT_DIR ?= template-images
TEMPLATES_IMG_TAG ?= $(shell git describe --always --dirty --tags)

# Template image names
ADDONS_MANAGER_TEMPLATES_IMG_NAME ?= tanzu-addons-manager-templates
KAPP_CONTROLLER_TEMPLATES_IMG_NAME ?= kapp-controller-templates
METRICS_SERVER_TEMPLATES_IMG_NAME ?= metrics-server-templates
VSPHERE_CPI_TEMPLATES_IMG_NAME ?= vsphere-cpi-templates
PINNIPED_ADDON_TEMPLATES_IMG_NAME ?= pinniped-templates
VSPHERE_CSI_TEMPLATES_IMG_NAME ?= vsphere-csi-templates
ANTREA_TEMPLATES_IMG_NAME ?= antrea-templates
CALICO_TEMPLATES_IMG_NAME ?= calico-templates

# Directories
BIN_DIR       := bin
export PATH := $(abspath $(BIN_DIR)):$(PATH)

# Binaries
MANAGER       := $(BIN_DIR)/manager

# golang container for building binaries
GO_CONTAINER ?= mirror.gcr.io/library/golang:1.14

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
	go test ./... -coverprofile cover.out -v 2

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

.PHONY: clean
clean: docker-remove ## Remove all generated files
	rm -rf bin
	rm -rf artifacts

## --------------------------------------
## Addon templates
## --------------------------------------
build-addon-template-images:
	IMG_NAME="$(IMAGE_REPO)/$(ADDONS_MANAGER_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C addons-manager
	IMG_NAME="$(IMAGE_REPO)/$(KAPP_CONTROLLER_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C kapp-controller
	IMG_NAME="$(IMAGE_REPO)/$(VSPHERE_CPI_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C vsphere_cpi
	IMG_NAME="$(IMAGE_REPO)/$(PINNIPED_ADDON_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C pinniped
	IMG_NAME="$(IMAGE_REPO)/$(METRICS_SERVER_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C metrics-server
	IMG_NAME="$(IMAGE_REPO)/$(VSPHERE_CSI_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C vsphere_csi
	IMG_NAME="$(IMAGE_REPO)/$(ANTREA_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C antrea
	IMG_NAME="$(IMAGE_REPO)/$(CALICO_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) template-image -C calico
  	
save-addon-template-images:
	mkdir -p "${ROOT_DIR}addons/artifacts"
	docker save "$(IMAGE_REPO)/${ADDONS_MANAGER_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${ADDONS_MANAGER_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"
	docker save "$(IMAGE_REPO)/${KAPP_CONTROLLER_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${KAPP_CONTROLLER_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"
	docker save "$(IMAGE_REPO)/${VSPHERE_CPI_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${VSPHERE_CPI_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"
	docker save "$(IMAGE_REPO)/${PINNIPED_ADDON_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${PINNIPED_ADDON_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"
	docker save "$(IMAGE_REPO)/${METRICS_SERVER_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${METRICS_SERVER_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"
	docker save "$(IMAGE_REPO)/${VSPHERE_CSI_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${VSPHERE_CSI_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"
	docker save "$(IMAGE_REPO)/${ANTREA_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${ANTREA_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"
	docker save "$(IMAGE_REPO)/${CALICO_TEMPLATES_IMG_NAME}:${TEMPLATES_IMG_TAG}" | gzip -c > "${ROOT_DIR}addons/artifacts/${CALICO_TEMPLATES_IMG_NAME}-${TEMPLATES_IMG_TAG}.tar.gz"

push-addon-template-images: ## Push template images
	IMG_NAME="$(IMAGE_REPO)/$(ADDONS_MANAGER_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C addons-manager
	IMG_NAME="$(IMAGE_REPO)/$(KAPP_CONTROLLER_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C kapp-controller
	IMG_NAME="$(IMAGE_REPO)/$(VSPHERE_CPI_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C vsphere_cpi
	IMG_NAME="$(IMAGE_REPO)/$(PINNIPED_ADDON_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C pinniped
	IMG_NAME="$(IMAGE_REPO)/$(METRICS_SERVER_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C metrics-server
	IMG_NAME="$(IMAGE_REPO)/$(VSPHERE_CSI_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C vsphere_csi
	IMG_NAME="$(IMAGE_REPO)/$(ANTREA_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C antrea
	IMG_NAME="$(IMAGE_REPO)/$(CALICO_TEMPLATES_IMG_NAME)" IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) push-template-image -C calico

## --------------------------------------
## BOM yaml that contains all add-ons
## --------------------------------------
bom-yaml:
	mkdir -p "${ROOT_DIR}addons/artifacts"
	IMG_TAG=$(TEMPLATES_IMG_TAG) $(MAKE) generate-addon-bom -C bom
