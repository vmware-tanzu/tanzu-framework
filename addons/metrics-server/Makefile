# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help

ROOT_DIR	:= $(dir $(shell pwd))
IMG_CATEGORY := metrics
IMG_CLUSTER_TYPE := management workload
ADDON_NAME := metrics-server-addon

IMG_NAME ?= metrics-server-templates
IMG_TAG ?= latest

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Build templates
## --------------------------------------

.PHONY: template-image
template-image: ## Build template image
	docker build -t $(IMG_NAME):$(IMG_TAG) -f ../Dockerfile.templates .
