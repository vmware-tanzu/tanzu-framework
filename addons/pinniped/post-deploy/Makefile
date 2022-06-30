# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# If you update this file, please follow:
# https://suva.sh/posts/well-documented-makefiles/

.DEFAULT_GOAL:=help

GIT_VERSION ?= $(shell git describe --always --tags)
GIT_REF_LONG = $(shell git rev-parse --verify HEAD)
# Release version
VERSION ?= $(GIT_VERSION)

.PHONY: test
test: fmt vet ## Run tests
	go test ./... -coverprofile coverage.txt -v 2

.PHONY: fmt
fmt: ## Format the code base
	go fmt ./...

.PHONY: vet
vet: ## Vet codebase
	go vet ./...

native: ## Build binary
	CGO_ENABLED=0 go build -o tkg-pinniped-post-deploy-job -ldflags="-s -w -X $(GOTARGET)/pkg/buildinfo.Version=$(GIT_VERSION) -X github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/buildinfo.GitSHA=$(GIT_REF_LONG)" github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/cmd/job
	CGO_ENABLED=0 go build -o tkg-pinniped-post-deploy-controller -ldflags="-s -w -X $(GOTARGET)/pkg/buildinfo.Version=$(GIT_VERSION) -X github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/buildinfo.GitSHA=$(GIT_REF_LONG)" github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/cmd/controller

.PHONY: build-images
build-images: ## Build tkg-pinniped-post-deploy docker images
	VERSION=$(VERSION) ./hack/scripts/build-images.sh

run:
	go run ./cmd/job \
        --supervisor-namespace=pinniped-supervisor \
        --concierge-namespace=pinniped-concierge \
        --supervisor-svc-name=pinniped-supervisor \
        --federationdomain-name=pinniped-federation-domain \
        --jwtauthenticator-name=tkg-jwt-authenticator \
        --supervisor-cert-name=pinniped-cert \
        --dex-namespace=tanzu-system-auth \
        --dex-svc-name=dexsvc \
        --dex-cert-name=dex-cert \
        --dex-configmap-name=dex \
        --is-dex-required=false
