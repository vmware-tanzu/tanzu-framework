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

.PHONY: docker-build
docker-build:
	cd ../../../../ && docker build -t $(IMG) -f pkg/v1/sdk/capabilities/Dockerfile --build-arg LD_FLAGS="$(LD_FLAGS)" .
