# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../../../common.mk

IMG ?= capabilities-controller-manager:latest
CRD_OPTIONS ?= "crd"

all: manager

# Run tests
test: fmt vet manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: fmt vet
	go build -ldflags "$(LD_FLAGS)" -o bin/manager ../../../../cmd/managers/capabilities/main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run -ldflags "$(LD_FLAGS)" ../../../../cmd/managers/capabilities/main.go

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

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
	cd ../../../../ && docker build -t $(IMG) -f pkg/v1/sdk/capabilities/Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" .
