# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)

NUL = /dev/null
ifeq ($(GOHOSTOS),windows)
	NUL = NUL
endif

ifeq ($(GOHOSTOS), linux)
XDG_DATA_HOME := ${HOME}/.local/share
endif
ifeq ($(GOHOSTOS), darwin)
XDG_DATA_HOME := "$${HOME}/Library/Application Support"
endif

# Directories
TOOLS_DIR := $(abspath hack/tools)
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin
BIN_DIR := bin
ROOT_DIR := $(shell git rev-parse --show-toplevel)
ADDONS_DIR := addons
UI_DIR := pkg/v1/tkg/web

# Add tooling binaries here and in hack/tools/Makefile
GOLANGCI_LINT := $(TOOLS_BIN_DIR)/golangci-lint
GOIMPORTS := $(TOOLS_BIN_DIR)/goimports
GOBINDATA := $(TOOLS_BIN_DIR)/gobindata
KUBEBUILDER := $(TOOLS_BIN_DIR)/kubebuilder
YTT := $(TOOLS_BIN_DIR)/ytt
KUBEVAL := $(TOOLS_BIN_DIR)/kubeval
GINKGO := $(TOOLS_BIN_DIR)/ginkgo
VALE := $(TOOLS_BIN_DIR)/vale
TOOLING_BINARIES := $(GOLANGCI_LINT) $(YTT) $(KUBEVAL) $(GOIMPORTS) $(GOBINDATA) $(GINKGO) $(VALE)

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
ifndef IS_OFFICIAL_BUILD
IS_OFFICIAL_BUILD = ""
endif

ifndef TANZU_PLUGIN_UNSTABLE_VERSIONS
TANZU_PLUGIN_UNSTABLE_VERSIONS = "experimental"
endif

# NPM registry to use for downloading node modules for UI build
CUSTOM_NPM_REGISTRY ?= $(shell git config tkg.npmregistry)

# TKG Compatibility Image repo and  path related configuration
ifndef TKG_DEFAULT_IMAGE_REPOSITORY
TKG_DEFAULT_IMAGE_REPOSITORY = "projects-stg.registry.vmware.com/tkg"
endif
ifndef TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH
# TODO change it to "tkg-compatibility" once the image is pushed to registry
TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH = "v1.4.0-zshippable/tkg-compatibility"
endif

DOCKER_DIR := /app
SWAGGER=docker run --rm -v ${PWD}:${DOCKER_DIR} quay.io/goswagger/swagger:v0.21.0

PRIVATE_REPOS="github.com/vmware-tanzu"
GO := GOPRIVATE=${PRIVATE_REPOS} go

# Add supported OS-ARCHITECTURE combinations here
ENVS := linux-amd64 windows-amd64 darwin-amd64

.DEFAULT_GOAL:=help

BUILD_SHA ?= $$(git describe --match=$(git rev-parse --short HEAD) --always --dirty)
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
BUILD_VERSION ?= $(shell git describe --tags --abbrev=0 2>$(NUL))
ifeq ($(strip $(BUILD_VERSION)),)
BUILD_VERSION = dev
endif
# BUILD_EDITION is the Tanzu Edition, the plugin should be built for.
# Valid values for BUILD_EDITION are 'tce' and 'tkg'. Default value of BUILD_EDITION is 'tkg'.
ifneq ($(BUILD_EDITION), tce)
BUILD_EDITION = tkg
endif

LD_FLAGS = -s -w
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli.BuildDate=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli.BuildSHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli.BuildVersion=$(BUILD_VERSION)'
LD_FLAGS += -X 'main.BuildEdition=$(BUILD_EDITION)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/buildinfo.IsOfficialBuild=$(IS_OFFICIAL_BUILD)'

BUILD_TAGS ?=

ARTIFACTS_DIR ?= ./artifacts

XDG_CACHE_HOME := ${HOME}/.cache

export XDG_DATA_HOME
export XDG_CACHE_HOME

## --------------------------------------
## API/controller building and generation
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

all: manager ui-build build-cli

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

generate-go: $(COUNTERFEITER) ## Generate code via go generate.
	PATH=$(abspath hack/tools/bin):$(PATH) go generate ./...

generate: controller-gen ## Generate code via controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt",year=$(shell date +%Y) paths="./..."
	$(MAKE) fmt

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

## --------------------------------------
## Tooling Binaries
## --------------------------------------

tools: $(TOOLING_BINARIES) ## Build tooling binaries
.PHONY: $(TOOLING_BINARIES)
$(TOOLING_BINARIES):
	make -C $(TOOLS_DIR) $(@F)

## --------------------------------------
## Version
## --------------------------------------

.PHONY: version
version: ## Show version
	@echo $(BUILD_VERSION)

## --------------------------------------
## Build prerequisites
## --------------------------------------

.PHONY: ensure-pinniped-repo
ensure-pinniped-repo:
	@rm -rf pinniped
	@mkdir -p pinniped
	@GIT_TERMINAL_PROMPT=0 git clone -q --depth 1 --branch $(PINNIPED_GIT_COMMIT) ${PINNIPED_GIT_REPOSITORY} pinniped > ${NUL} 2>&1

.PHONY: prep-build-cli
prep-build-cli: ensure-pinniped-repo
	$(GO) mod download
	$(GO) mod tidy
	EMBED_PROVIDERS_TAG=embedproviders
ifeq "${BUILD_TAGS}" "${EMBED_PROVIDERS_TAG}"
	make -C pkg/v1/providers -f Makefile generate-provider-bundle-zip
endif

.PHONY: configure-buildtags-%
configure-buildtags-%:
ifeq ($(strip $(BUILD_TAGS)),)
	$(eval TAGS = $(word 1,$(subst -, ,$*)))
	$(eval BUILD_TAGS=$(TAGS))
endif
	@echo "BUILD_TAGS set to '$(BUILD_TAGS)'"

## --------------------------------------
## Build binaries and plugins
## --------------------------------------

# Dynamically generate the OS-ARCH targets to allow for parallel execution
CLI_JOBS := $(addprefix build-cli-,${ENVS})
RELEASE_JOBS := $(addprefix release-,${ENVS})

.PHONY: build-cli
build-cli: build-plugin-admin ${CLI_JOBS} ## Build Tanzu CLI
	@rm -rf pinniped

.PHONY: build-plugin-admin
build-plugin-admin:
	@echo build version: $(BUILD_VERSION)
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile $(addprefix --target ,$(subst -,_,${ENVS})) --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin/${GOHOSTOS}/${GOHOSTARCH}/cli

.PHONY: build-cli-%
build-cli-%: prep-build-cli
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))

	@if [ "$(filter $(OS)-$(ARCH),$(ENVS))" = "" ]; then\
		printf "\n\n======================================\n";\
		printf "! $(OS)-$(ARCH) is not an officially supported platform!\n";\
		printf "! Make sure to perform a full build to make sure expected plugins are available!\n";\
		printf "======================================\n\n";\
	fi

	./hack/embed-pinniped-binary.sh go ${OS} ${ARCH}
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --corepath "cmd/cli/tanzu" --artifacts artifacts/${OS}/${ARCH}/cli --target  ${OS}_${ARCH}

## --------------------------------------
## Build locally
## --------------------------------------

# By default `make build-cli-local` ensures that the cluster and management-cluster plugins are build
# with embedded provider templates. This is used only for dev build and not for production builds.
# When using embedded providers, `~/.config/tanzu/tkg/providers` directory always gets overwritten with the
# embdedded providers. To skip the provider updates, specify `SUPPRESS_PROVIDERS_UPDATE` environment variable.
# Note: If any local builds want to skip embedding providers and want utilize providers from TKG BoM file,
# To skip provider embedding, pass `BUILD_TAGS=skipembedproviders` to make target (`make BUILD_TAGS=skipembedproviders build-cli-local)
.PHONY: build-cli-local
build-cli-local: configure-buildtags-embedproviders build-cli-${GOHOSTOS}-${GOHOSTARCH} ## Build Tanzu CLI locally. cluster and management-cluster plugins are built with embedded providers.
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin/${GOHOSTOS}/${GOHOSTARCH}/cli --target local

.PHONY: build-install-cli-local
build-install-cli-local: clean-catalog-cache clean-cli-plugins build-cli-local install-cli-plugins install-cli ## Local build and install the CLI plugins

## --------------------------------------
## manage cli mocks
## --------------------------------------

.PHONY: build-cli-mocks
build-cli-mocks: ## Build Tanzu CLI mocks
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile $(addprefix --target ,$(subst -,_,${ENVS})) --version 0.0.1 --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./test/cli/mock/plugin-old --artifacts ./test/cli/mock/artifacts-old
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile $(addprefix --target ,$(subst -,_,${ENVS})) --version 0.0.2 --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./test/cli/mock/plugin-new --artifacts ./test/cli/mock/artifacts-new
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile $(addprefix --target ,$(subst -,_,${ENVS})) --version 0.0.3 --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./test/cli/mock/plugin-alt --artifacts ./test/cli/mock/artifacts-alt

## --------------------------------------
## install binaries and plugins
## --------------------------------------

.PHONY: install-cli
install-cli: ## Install Tanzu CLI
	$(GO) install -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu

# Note: Invoking this target will update the unstableVersionSelector config
# file setting to 'experimental' by default. Use TANZU_PLUGIN_UNSTABLE_VERSIONS to
# override if necessary.
.PHONY: install-cli-plugins
install-cli-plugins: set-unstable-versions  ## Install Tanzu CLI plugins
	TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
    		plugin install all --local $(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli
	TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR)-admin/$(GOHOSTOS)/$(GOHOSTARCH)/cli
	TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go \
		test fetch --local $(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli --local $(ARTIFACTS_DIR)-admin/$(GOHOSTOS)/$(GOHOSTARCH)/cli

.PHONY: set-unstable-versions
set-unstable-versions:  ## Configures the unstable versions
	TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/cli/tanzu/main.go config set unstable-versions $(TANZU_PLUGIN_UNSTABLE_VERSIONS)

.PHONY: build-install-cli-all ## Build and install the CLI plugins
build-install-cli-all: clean-catalog-cache clean-cli-plugins build-cli install-cli-plugins install-cli ## Build and install Tanzu CLI plugins

# This target is added as some tests still relies on tkg cli.
# TODO: Remove this target when all tests are migrated to use tanzu cli
.PHONY: tkg-cli ## Builds tkg-cli binary
tkg-cli: configure-buildtags-embedproviders configure-bom prep-build-cli ## Build tkg CLI binary only, and without rebuilding ui bits (providers are embedded to the binary)
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/tkg-${GOHOSTOS}-${GOHOSTARCH} -ldflags "${LD_FLAGS}" -tags "${BUILD_TAGS}" cmd/cli/tkg/main.go

.PHONY: build-cli-image
build-cli-image: ## Build the CLI image
	docker build -t projects.registry.vmware.com/tanzu/cli:latest -f Dockerfile.cli .

## --------------------------------------
## Release binaries
## --------------------------------------

# TODO (pbarker): should work this logic into the builder plugin
.PHONY: release
release: ensure-pinniped-repo ${RELEASE_JOBS}
	@rm -rf pinniped

.PHONY: release-%
release-%:
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))

	./hack/embed-pinniped-binary.sh go ${OS} ${ARCH}
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --corepath "cmd/cli/tanzu" --artifacts artifacts/${OS}/${ARCH}/cli --target  ${OS}_${ARCH}

## --------------------------------------
## Testing, verification, formating and cleanup
## --------------------------------------

.PHONY: test
test: generate fmt vet manifests build-cli-mocks ## Run tests
	## Skip running TKG integration tests
	$(GO) test -coverprofile cover.out -v `go list ./... | grep -v github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test`
	$(MAKE) kubebuilder -C $(TOOLS_DIR)
	KUBEBUILDER_ASSETS=$(ROOT_DIR)/$(KUBEBUILDER)/bin GOPRIVATE=$(PRIVATE_REPOS) $(MAKE) test -C addons

.PHONY: test-cli
test-cli: build-cli-mocks ## Run tests
	$(GO) test ./...

fmt: tools ## Run goimports
	$(GOIMPORTS) -w -local github.com/vmware-tanzu ./

vet: ## Run go vet
	$(GO) vet ./...

lint: tools doc-lint ## Run linting checks
	# Linter runs per module, add each one here and make sure they match
	# in .github/workflows/main.yaml for CI coverage
	$(GOLANGCI_LINT) run -v
	cd $(ADDONS_DIR); $(GOLANGCI_LINT) run -v
	cd $(ADDONS_DIR)/pinniped/post-deploy/; $(GOLANGCI_LINT) run -v

doc-lint: tools ## Run linting checks for docs
	$(VALE) --config=.vale/config.ini --glob='*.md' ./

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	$(GO) mod tidy
	cd $(ADDONS_DIR); $(GO) mod tidy
	cd $(ADDONS_DIR)/pinniped/post-deploy/; $(GO) mod tidy
	cd $(TOOLS_DIR); $(GO) mod tidy

.PHONY: verify
verify: ## Run all verification scripts
	./hack/verify-dirty.sh

.PHONY: clean-catalog-cache
clean-catalog-cache: ## Cleans catalog cache
	@rm -rf ${XDG_CACHE_HOME}/tanzu/*

.PHONY: clean-cli-plugins
clean-cli-plugins: ## Remove Tanzu CLI plugins
	- rm -rf ${XDG_DATA_HOME}/tanzu-cli/*

## --------------------------------------
## UI Build & Test
## --------------------------------------
.PHONY: update-npm-registry
update-npm-registry: ## set alternate npm registry
ifneq ($(strip $(CUSTOM_NPM_REGISTRY)),)
	npm config set registry $(CUSTOM_NPM_REGISTRY) web
endif

.PHONY: ui-dependencies
ui-dependencies: update-npm-registry  ## install UI dependencies (node modules)
	cd $(UI_DIR); NG_CLI_ANALYTICS=ci npm ci; cd ../

.PHONY: ui-build
ui-build: ui-dependencies ## install dependencies, then compile client UI for production
	cd $(UI_DIR); npm run build:prod; cd ../
	$(MAKE) generate-ui-bindata

.PHONY: ui-build-and-test
ui-build-and-test: ui-dependencies ## compile client UI for production and run tests
	cd $(UI_DIR); npm run build:ci; cd ../
	$(MAKE) generate-ui-bindata

.PHONY: verify-ui-bindata
verify-ui-bindata: ## Run verification for ui bindata
	git diff --exit-code pkg/v1/tkg/manifest/server/zz_generated.bindata.go

## --------------------------------------
## Generate files
## --------------------------------------

.PHONY: cobra-docs
cobra-docs:
	TANZU_CLI_NO_INIT=true TANZU_CLI_NO_COLOR=true $(GO) run ./cmd/cli/tanzu generate-all-docs
	sed -i.bak -E 's/\/[A-Za-z]*\/([a-z]*)\/.config\/tanzu\/pinniped\/sessions.yaml/~\/.config\/tanzu\/pinniped\/sessions.yaml/g' docs/cli/commands/tanzu_pinniped-auth_login.md

.PHONY: generate-fakes
generate-fakes: ## Generate fakes for writing unit tests
	$(GO) generate ./...
	$(MAKE) fmt

.PHONY: generate-ui-bindata
generate-ui-bindata: $(GOBINDATA) ## Generate go-bindata for ui files
	$(GOBINDATA) -mode=420 -modtime=1 -o=pkg/v1/tkg/manifest/server/zz_generated.bindata.go -pkg=server $(UI_DIR)/dist/...
	$(MAKE) fmt

.PHONY: generate-telemetry-bindata
generate-telemetry-bindata: $(GOBINDATA) ## Generate telemetry bindata
	$(GOBINDATA) -mode=420 -modtime=1 -o=pkg/v1/tkg/manifest/telemetry/zz_generated.bindata.go -pkg=telemetry pkg/v1/tkg/manifest/telemetry/...
	$(MAKE) fmt

 # TODO: Remove bindata dependency and use go embed
.PHONY: generate-bindata ## Generate go-bindata files
generate-bindata: generate-telemetry-bindata generate-ui-bindata

.PHONY: configure-bom
configure-bom:
	# Update default BoM Filename variable in tkgconfig pkg
	sed "s+TKG_DEFAULT_IMAGE_REPOSITORY+${TKG_DEFAULT_IMAGE_REPOSITORY}+g"  hack/update-bundled-bom-filename/update-bundled-default-bom-files-configdata.txt | \
	sed "s+TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH+${TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH}+g" | \
	sed "s+TKG_MANAGEMENT_CLUSTER_PLUGIN_VERSION+${BUILD_VERSION}+g"  > pkg/v1/tkg/tkgconfigpaths/zz_bundled_default_bom_files_configdata.go

.PHONY: generate-ui-swagger-api
generate-ui-swagger-api: ## Generate swagger files for UI backend
	rm -rf ${UI_DIR}/server/client  ${UI_DIR}/server/models ${UI_DIR}/server/restapi/operations
	${SWAGGER} generate server -q -A kickstartUI -t $(DOCKER_DIR)/${UI_DIR}/server -f $(DOCKER_DIR)/${UI_DIR}/api/spec.yaml --exclude-main
	${SWAGGER} generate client -q -A kickstartUI -t $(DOCKER_DIR)/${UI_DIR}/server -f $(DOCKER_DIR)/${UI_DIR}/api/spec.yaml
	# reset the server.go file to avoid goswagger overwritting our custom changes.
	git reset HEAD ${UI_DIR}/server/restapi/server.go
	git checkout HEAD ${UI_DIR}/server/restapi/server.go
	$(MAKE) fmt

## --------------------------------------
## Provider templates/overlays
## --------------------------------------

.PHONY: clustergen
clustergen:
	CLUSTERGEN_BASE=${CLUSTERGEN_BASE} make -C pkg/v1/providers -f Makefile cluster-generation-diffs

.PHONY: generate-embedproviders
generate-embedproviders:
	make -C pkg/v1/providers -f Makefile generate-provider-bundle-zip

## --------------------------------------
## TKG integration tests
## --------------------------------------

GINKGO_NODES  ?= 1
GINKGO_NOCOLOR ?= false

.PHONY: e2e-tkgctl-docker
e2e-tkgctl-docker: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders pkg/v1/tkg/test/tkgctl/docker

.PHONY: e2e-tkgctl-azure
e2e-tkgctl-azure: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders pkg/v1/tkg/test/tkgctl/azure

.PHONY: e2e-tkgctl-aws
e2e-tkgctl-aws: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders pkg/v1/tkg/test/tkgctl/aws

.PHONY: e2e-tkgctl-vc67
e2e-tkgctl-vc67: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders pkg/v1/tkg/test/tkgctl/vsphere67

.PHONY: e2e-tkgpackageclient-docker
e2e-tkgpackageclient-docker: $(GINKGO) generate-embedproviders ## Run ginkgo tkgpackageclient E2E tests
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders pkg/v1/tkg/test/tkgpackageclient
