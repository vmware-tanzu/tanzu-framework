# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../common.mk

IMG_DEFAULT_NAME := capabilities-controller-manager
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
test: fmt vet
	cd client && go test ./... -coverprofile cover.out
	cd controller && go test ./... -coverprofile cover.out

# Build manager binary
manager: fmt vet
	go build -ldflags "$(LD_FLAGS)" -o bin/manager ./controller/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	cd controller && go run -ldflags "$(LD_FLAGS)" ./main.go

# Run go fmt against code
fmt:
	cd controller && go fmt ./...

# Run go vet against code
vet:
	cd controller && go vet ./...

# Generate deployment manifest for the tkr manager
deployment: manifests
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default

manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		$(CRD_OPTIONS) \
		paths=../apis/... \
		output:crd:artifacts:config=../../../config/crd/bases

controller-gen: ## Download controller-gen
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	$(GO) mod init tmp ;\
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.5 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

.PHONY: docker-build
docker-build:
	cd .. && docker build -t $(IMG) -f capabilities/controller/Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" --build-arg DISTROLESS_BASE_IMAGE="$(DISTROLESS_BASE_IMAGE)" --build-arg GOPROXY="$(GOPROXY)" --build-arg GOSUMDB="$(GOSUMDB)" .

.PHONY: docker-publish
docker-publish: ## Publish docker image
	docker push $(IMG)

.PHONY: kbld-image-replace
kbld-image-replace: ## Add newImage in kbld-config.yaml
	cd ../hack/packages/kbld-image-replace && \
		go run main.go -kbld-config ../../../packages/capabilities/kbld-config.yaml $(IMG_DEFAULT_NAME_TAG) $(IMG)

.PHONY: docker-image-names
docker-image-names:
	@echo $(IMG)

.PHONY: docker-build-and-publish
docker-build-and-publish: docker-build docker-publish kbld-image-replace
