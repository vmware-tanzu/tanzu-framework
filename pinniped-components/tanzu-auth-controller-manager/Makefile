# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# This Makefile is here to follow step 4 of this README:
#   https://github.com/vmware-tanzu/tanzu-framework/blob/5cdba3181afbc31b05d4df51052fb47f3fcb5b92/docs/packages/add-new-package-to-package-repo.md

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

LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pinniped-components/pkg/buildinfo.Date=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pinniped-components/pkg/buildinfo.SHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pinniped-components/pkg/buildinfo.Version=$(BUILD_VERSION)'

IMG_DEFAULT_NAME := tanzu-auth-controller-manager
IMG_DEFAULT_TAG := latest
IMG_DEFAULT_NAME_TAG := $(IMG_DEFAULT_NAME):$(IMG_DEFAULT_TAG)

IMG_VERSION_OVERRIDE ?= $(IMG_DEFAULT_TAG)

#Adding Support for GOPROXY and GOSUMDB
GOPROXY ?= direct
GOSUMDB ?= off
DISTROLESS_BASE_IMAGE ?= gcr.io/distroless/static:nonroot

ifeq ($(strip $(OCI_REGISTRY)),)
	IMG ?= $(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
else
	IMG ?= $(OCI_REGISTRY)/$(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
endif

.PHONY: docker-build
docker-build: ## Build docker image
	docker build -t $(IMG) -f Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" --build-arg DISTROLESS_BASE_IMAGE="$(DISTROLESS_BASE_IMAGE)" -e GOPROXY="$(GOPROXY)" -e GOSUMDB="$(GOSUMDB)" .

.PHONY: docker-publish
docker-publish: ## Publish docker image
	docker push $(IMG)

.PHONY: kbld-image-replace
kbld-image-replace: ## Add newImage in kbld-config.yaml
	cd ../../hack/packages/kbld-image-replace && \
		go run main.go -kbld-config ../../../packages/tanzu-auth/kbld-config.yaml $(IMG_DEFAULT_NAME_TAG) $(IMG)
