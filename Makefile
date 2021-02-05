# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

ifeq ($(OS),Windows_NT)
	build_OS := Windows
	NUL = NUL
else
	build_OS := $(shell uname -s 2>/dev/null || echo Unknown)
	NUL = /dev/null
endif

# Directories
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

GOBINDATA := $(TOOLS_BIN_DIR)/go-bindata-$(GOOS)-$(GOARCH)

PINNIPED_GIT_REPOSITORY = https://github.com/vmware-tanzu/pinniped.git
ifeq ($(strip $(PINNIPED_GIT_COMMIT)),)
PINNIPED_GIT_COMMIT = v0.4.0
endif

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

PRIVATE_REPOS="github.com/vmware-tanzu-private"
GO := GOPRIVATE=${PRIVATE_REPOS} go

.DEFAULT_GOAL:=help

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

all: manager build-cli

test: generate fmt vet manifests build-cli-mocks ## Run tests
	$(GO) test ./... -coverprofile cover.out

manager: generate fmt vet ## Build manager binary
	$(GO) build -o bin/manager main.go

run: generate fmt vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	$(GO) run ./main.go

install: manifests ## Install CRDs into a cluster
	kustomize build config/crd | kubectl apply -f -

uninstall: manifests ## Uninstall CRDs from a cluster
	kustomize build config/crd | kubectl delete -f -

deploy: manifests ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

manifests: controller-gen ## Generate manifests e.g. CRD, RBAC etc.
	$(CONTROLLER_GEN) \
		$(CRD_OPTIONS) \
		paths=./apis/... \
		output:crd:artifacts:config=config/crd/bases

fmt: ## Run go fmt
	$(GO) fmt ./...

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run linting checks
# Make sure you have golangci-lint installed: https://golangci-lint.run/usage/install/
	golangci-lint run ./...
	golint -set_exit_status ./...

generate: controller-gen ## Generate code via controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt",year=$(shell date +%Y) paths="./..."

docker-build: test ## Build the docker image
	docker build . -t ${IMG}

docker-push: ## Push the docker image
	docker push ${IMG}

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


BUILD_SHA ?= $$(git describe --match=$(git rev-parse --short HEAD) --always --dirty)
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
BUILD_VERSION := $(shell git describe --tags --abbrev=0 2>$(NUL))
ifeq ($(strip $(BUILD_VERSION)),)
BUILD_VERSION = dev
endif

LD_FLAGS = -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildDate=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildSHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu-private/core/pkg/v1/cli.BuildVersion=$(BUILD_VERSION)'


ARTIFACTS_DIR ?= ./artifacts

ifeq ($(build_OS), Linux)
XDG_DATA_HOME := ${HOME}/.local/share
endif
ifeq ($(build_OS), Darwin)
XDG_DATA_HOME := ${HOME}/Library/ApplicationSupport
endif

export XDG_DATA_HOME


.PHONY: install-cli
install-cli: ## Install Tanzu CLI
	$(GO) install -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu

# TODO (vuil) main artifacts build is split into individual targets for now to build/prepare
# target-specific pinniped binary for the pinniped-auth plugin. Collapse to
# single invocation once that is no longer needed.
.PHONY: build-cli
build-cli: prep-build-cli ## Build Tanzu CLI
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) linux amd64
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --target linux_amd64
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) windows amd64
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --target windows_amd64
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) darwin amd64
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --target darwin_amd64
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) linux 386
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --target linux_386
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) windows 386
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --target windows_386
	@rm -rf pinniped

.PHONY: build-cli-local
build-cli-local: prep-build-cli ## Build Tanzu CLI locally
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --target local
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin --target local

.PHONY: build-cli-mocks
build-cli-mocks: ## Build Tanzu CLI mocks
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version 0.0.1 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-old --artifacts ./test/cli/mock/artifacts-old 
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version 0.0.2 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-new --artifacts ./test/cli/mock/artifacts-new
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version 0.0.3 --ldflags "$(LD_FLAGS)" --path ./test/cli/mock/plugin-alt --artifacts ./test/cli/mock/artifacts-alt

.PHONY: build-cli-image
build-cli-image: ## Build the CLI image
	docker build -t projects.registry.vmware.com/tanzu/cli:latest -f Dockerfile.cli .

.PHONY: test-cli
test-cli: build-cli-mocks ## Run tests
	$(GO) test ./...

.PHONY: build-install-cli-all ## Build and install the CLI plugins
build-install-cli-all: clean-cli-plugins build-cli install-cli-plugins install-cli ## Build and install Tanzu CLI plugins

install-cli-plugins: TANZU_CLI_NO_INIT=true

.PHONY: install-cli-plugins
install-cli-plugins:  ## Install Tanzu CLI plugins 
	$(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR) --local $(ARTIFACTS_DIR)-admin
	$(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		test fetch --local $(ARTIFACTS_DIR) --local $(ARTIFACTS_DIR)-admin

.PHONY: clean-cli-plugins
clean-cli-plugins: ## Remove Tanzu CLI plugins
	- rm -rf ${XDG_DATA_HOME}/tanzu-cli/*

.PHONY: ensure-pinniped-repo
ensure-pinniped-repo: $(GOBINDATA)
	@rm -rf pinniped
	@mkdir -p pinniped
	@GIT_TERMINAL_PROMPT=0 git clone -q --depth 1 --branch $(PINNIPED_GIT_COMMIT) ${PINNIPED_GIT_REPOSITORY} pinniped

.PHONY: generate-pinniped-bindata
generate-pinniped-bindata: ensure-pinniped-repo
	cd pinniped && $(GO) build -o pinniped ./cmd/pinniped
	$(GOBINDATA) -mode=420 -modtime=1 -o=pkg/v1/auth/tkg/zz_generated.bindata.go -pkg=tkgauth pinniped/pinniped
	git update-index --assume-unchanged pkg/v1/auth/tkg/zz_generated.bindata.go
	@rm -rf pinniped

.PHONY: prep-build-cli
prep-build-cli: ensure-pinniped-repo
	$(GO) mod download


$(GOBINDATA): $(TOOLS_DIR)/go.mod # Build go-bindata from tools folder
	mkdir -p $(TOOLS_BIN_DIR)
	cd $(TOOLS_DIR); $(GO) build -tags=tools -o ../../$(TOOLS_BIN_DIR) github.com/shuLhan/go-bindata/... ; mv ../../$(TOOLS_BIN_DIR)/go-bindata ../../$(GOBINDATA)


# TODO (pbarker): should work this logic into the builder plugin
.PHONY: release
release: ensure-pinniped-repo
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) linux amd64
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --artifacts artifacts/linux/amd64/cli --target linux_amd64
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) windows amd64
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --artifacts artifacts/windows/amd64/cli --target windows_amd64
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) darwin amd64
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --artifacts artifacts/darwin/amd64/cli --target darwin_amd64
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) linux 386
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --artifacts artifacts/linux/386/cli --target linux_386
	./hack/generate-pinniped-bindata.sh go $(GOBINDATA) windows 386
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --corepath "cmd/cli/tanzu" --artifacts artifacts/windows/386/cli --target windows_386
	@rm -rf pinniped
