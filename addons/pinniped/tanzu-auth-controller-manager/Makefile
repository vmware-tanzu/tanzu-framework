# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# This Makefile is here to follow step 4 of this README:
#   https://github.com/vmware-tanzu/tanzu-framework/blob/5cdba3181afbc31b05d4df51052fb47f3fcb5b92/docs/packages/add-new-package-to-package-repo.md

IMG_DEFAULT_NAME := tanzu-auth-controller-manager
IMG_DEFAULT_TAG := latest
IMG_DEFAULT_NAME_TAG := $(IMG_DEFAULT_NAME):$(IMG_DEFAULT_TAG)

IMG_VERSION_OVERRIDE ?= $(IMG_DEFAULT_TAG)

ifeq ($(strip $(OCI_REGISTRY)),)
	IMG ?= $(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
else
	IMG ?= $(OCI_REGISTRY)/$(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
endif

.PHONY: docker-build
docker-build: ## Build docker image
	docker build -t $(IMG) -f Dockerfile .

.PHONY: docker-publish
docker-publish: ## Publish docker image
	docker push $(IMG)

.PHONY: kbld-image-replace
kbld-image-replace: ## Add newImage in kbld-config.yaml
	cd ../../../hack/packages/kbld-image-replace && $(MAKE) run IMAGE=$(IMG_DEFAULT_NAME_TAG) NEW_IMAGE=$(IMG)
