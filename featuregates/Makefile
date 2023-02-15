# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../common.mk

IMG_DEFAULT_NAME := featuregates-controller-manager
IMG_DEFAULT_TAG := latest
IMG_DEFAULT_NAME_TAG := $(IMG_DEFAULT_NAME):$(IMG_DEFAULT_TAG)

IMG_VERSION_OVERRIDE ?= $(IMG_DEFAULT_TAG)

ifeq ($(strip $(OCI_REGISTRY)),)
	IMG ?= $(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
else
	IMG ?= $(OCI_REGISTRY)/$(IMG_DEFAULT_NAME):$(IMG_VERSION_OVERRIDE)
endif

CRD_OPTIONS ?= "crd"

#Adding Support for GOPROXY and GOSUMDB
GOPROXY ?= direct
GOSUMDB ?= off
DISTROLESS_BASE_IMAGE ?= gcr.io/distroless/static:nonroot

all: manager

# Run tests
test:
	cd client && go test ./... -coverprofile cover.out
	cd controller && go test ./... -coverprofile cover.out

# Build manager binary
manager:
	go build -ldflags "$(LD_FLAGS)" -o bin/manager ./controller/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run:
	cd controller && go run -ldflags "$(LD_FLAGS)" ./main.go

.PHONY: docker-build
docker-build: ## Build docker image
	cd ../ && docker build -t $(IMG) -f featuregates/controller/Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" --build-arg DISTROLESS_BASE_IMAGE="$(DISTROLESS_BASE_IMAGE)" -e GOPROXY="$(GOPROXY)" -e GOSUMDB="$(GOSUMDB)" .

.PHONY: docker-publish
docker-publish: ## Publish docker image
	docker push $(IMG)

.PHONY: kbld-image-replace
kbld-image-replace: ## Add newImage in kbld-config.yaml
	cd ../hack/packages/kbld-image-replace && \
		go run main.go -kbld-config ../../../packages/featuregates/kbld-config.yaml $(IMG_DEFAULT_NAME_TAG) $(IMG)

.PHONY: docker-image-names
docker-image-names:
	@echo $(IMG)

.PHONY: docker-build-and-publish
docker-build-and-publish: docker-build docker-publish kbld-image-replace
