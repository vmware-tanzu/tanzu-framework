# Copyright 2021 VMware, Inc. All Rights Reserved.
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

all: manager

# Run tests
test: fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: fmt vet
	go build -ldflags "$(LD_FLAGS)" -o bin/manager ./controller/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run:
	cd controller && go run -ldflags "$(LD_FLAGS)" ./main.go

# Generate deployment manifest for the tkr manager
deployment: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default

manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		$(CRD_OPTIONS) \
		paths=../apis/config/... \
		output:crd:artifacts:config=./config/crds

.PHONY: docker-build
docker-build: ## Build docker image
	cd ../ && docker build -t $(IMG) -f featuregate/controller/Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" .

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
