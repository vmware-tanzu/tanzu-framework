# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../../common.mk

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

all: manager

# Run tests
test: fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: fmt vet
	go build -ldflags "$(LD_FLAGS)" -o bin/manager ../../../../cmd/managers/featuregate/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run -ldflags "$(LD_FLAGS)" ../../../../cmd/managers/featuregate/main.go

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

.PHONY: docker-build
docker-build: ## Build docker image
	cd ../../../../ && docker build -t $(IMG) -f pkg/v1/sdk/features/Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" .

.PHONY: docker-publish
docker-publish: ## Publish docker image
	docker push $(IMG)

.PHONY: kbld-image-replace
kbld-image-replace: ## Add newImage in kbld-config.yaml
	cd ../../../../hack/packages/kbld-image-replace && $(MAKE) run IMAGE=$(IMG_DEFAULT_NAME_TAG) NEW_IMAGE=$(IMG)

.PHONY: docker-image-names
docker-image-names:
	@echo $(IMG)

.PHONY: docker-build-and-publish
docker-build-and-publish: docker-build docker-publish kbld-image-replace
