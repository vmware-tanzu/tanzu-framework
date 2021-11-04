# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ./common.mk

# Image URL to use all building/pushing image targets
IMG ?= controller:latest
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd"

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
YTT_TESTS_DIR := pkg/v1/providers/tests
PACKAGES_SCRIPTS_DIR := $(abspath hack/packages/scripts)
UI_DIR := pkg/v1/tkg/web

# Add tooling binaries here and in hack/tools/Makefile
GOLANGCI_LINT      := $(TOOLS_BIN_DIR)/golangci-lint
GOIMPORTS          := $(TOOLS_BIN_DIR)/goimports
GOBINDATA          := $(TOOLS_BIN_DIR)/gobindata
KUBEBUILDER        := $(TOOLS_BIN_DIR)/kubebuilder
YTT                := $(TOOLS_BIN_DIR)/ytt
KBLD               := $(TOOLS_BIN_DIR)/kbld
VENDIR             := $(TOOLS_BIN_DIR)/vendir
IMGPKG             := $(TOOLS_BIN_DIR)/imgpkg
KAPP               := $(TOOLS_BIN_DIR)/kapp
KUBEVAL            := $(TOOLS_BIN_DIR)/kubeval
GINKGO             := $(TOOLS_BIN_DIR)/ginkgo
VALE               := $(TOOLS_BIN_DIR)/vale
YQ                 := $(TOOLS_BIN_DIR)/yq
TOOLING_BINARIES   := $(GOLANGCI_LINT) $(YTT) $(KBLD) $(VENDIR) $(IMGPKG) $(KAPP) $(KUBEVAL) $(GOIMPORTS) $(GOBINDATA) $(GINKGO) $(VALE) $(YQ)

export REPO_VERSION ?= $(BUILD_VERSION)

PINNIPED_GIT_REPOSITORY = https://github.com/vmware-tanzu/pinniped.git
PINNIPED_VERSIONS = v0.4.4 v0.12.0

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
TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH = "framework-zshippable/tkg-compatibility"
endif

ifndef DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH
DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH = "packages/management/standalone-cliplugins"
endif
ifndef DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG
DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG = "${BUILD_VERSION}"
endif

DOCKER_DIR := /app
SWAGGER=docker run --rm -v ${PWD}:${DOCKER_DIR} quay.io/goswagger/swagger:v0.21.0

# OCI registry for hosting tanzu framework components (containers and packages)
OCI_REGISTRY ?= projects.registry.vmware.com/tanzu_framework

.DEFAULT_GOAL:=help

LD_FLAGS += -X 'main.BuildEdition=$(BUILD_EDITION)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.IsOfficialBuild=$(IS_OFFICIAL_BUILD)'

ifneq ($(strip $(TANZU_CORE_BUCKET)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.CoreBucketName=$(TANZU_CORE_BUCKET)'
endif

ifeq ($(TANZU_FORCE_NO_INIT), true)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/core.forceNoInit=true'
endif

ifneq ($(strip $(TKG_DEFAULT_IMAGE_REPOSITORY)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.DefaultStandaloneDiscoveryRepository=$(TKG_DEFAULT_IMAGE_REPOSITORY)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.DefaultStandaloneDiscoveryImagePath=$(DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.DefaultStandaloneDiscoveryImageTag=$(DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG)'
endif

BUILD_TAGS ?=

ARTIFACTS_DIR ?= ./artifacts

XDG_CACHE_HOME := ${HOME}/.cache
XDG_CONFIG_HOME :=${HOME}/.config

export XDG_DATA_HOME
export XDG_CACHE_HOME
export XDG_CONFIG_HOME
export OCI_REGISTRY

## --------------------------------------
## API/controller building and generation
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-25s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

all: manager ui-build build-cli

manager: generate fmt vet ## Build manager binary
	$(GO) build -ldflags "$(LD_FLAGS)" -o bin/manager main.go

run: generate fmt vet manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	$(GO) run -ldflags "$(LD_FLAGS)" ./main.go

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

controller-gen: ## Download controller-gen
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	$(GO) mod init tmp ;\
	$(GO) get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0 ;\
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
	@GIT_TERMINAL_PROMPT=0 git clone -q ${PINNIPED_GIT_REPOSITORY} pinniped > ${NUL} 2>&1

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
CLI_ADMIN_JOBS := $(addprefix build-plugin-admin-,${ENVS})
RELEASE_JOBS := $(addprefix release-,${ENVS})

.PHONY: build-cli
build-cli: ${CLI_ADMIN_JOBS} ${CLI_JOBS} ## Build Tanzu CLI
	@rm -rf pinniped

.PHONY: build-plugin-admin
build-plugin-admin: ${CLI_ADMIN_JOBS}

.PHONY: build-plugin-admin-%
build-plugin-admin-%:
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))

	@if [ "$(filter $(OS)-$(ARCH),$(ENVS))" = "" ]; then\
		printf "\n\n======================================\n";\
		printf "! $(OS)-$(ARCH) is not an officially supported platform!\n";\
		printf "! Make sure to perform a full build to make sure expected plugins are available!\n";\
		printf "======================================\n\n";\
	fi

	@echo build version: $(BUILD_VERSION)
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin/${OS}/${ARCH}/cli --target ${OS}_${ARCH}

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

	./hack/embed-pinniped-binary.sh go ${OS} ${ARCH} ${PINNIPED_VERSIONS}
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
## Build and publish CLIPlugin Discovery resource files and binaries
## --------------------------------------

STANDALONE_PLUGINS := login management-cluster package pinniped-auth
CONTEXT_PLUGINS := cluster kubernetes-release secret

.PHONY: publish-plugins-all-local
publish-plugins-all-local: ## Publish CLI Plugins locally for all supported os-arch
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type local --plugins "$(STANDALONE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --local-output-discovery-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/discovery/standalone" --local-output-distribution-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/distribution" --input-artifact-dir $(ARTIFACTS_DIR)
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type local --plugins "$(CONTEXT_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --local-output-discovery-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/discovery/context" --local-output-distribution-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/distribution" --input-artifact-dir $(ARTIFACTS_DIR)

.PHONY: publish-plugins-local
publish-plugins-local: ## Publish CLI Plugins locally for current host os-arch only
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type local --plugins "$(STANDALONE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --local-output-discovery-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/discovery/standalone" --local-output-distribution-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/distribution" --input-artifact-dir $(ARTIFACTS_DIR)
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type local --plugins "$(CONTEXT_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --local-output-discovery-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/discovery/context" --local-output-distribution-dir "$(XDG_CONFIG_HOME)/tanzu-plugins/distribution" --input-artifact-dir $(ARTIFACTS_DIR)

.PHONY: publish-plugins-all-oci
publish-plugins-all-oci: ## Publish CLI Plugins as OCI image for all supported os-arch
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type oci --plugins "$(STANDALONE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/standalone:v0.0.1 --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type oci --plugins "$(CONTEXT_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/context:v0.0.1 --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)

.PHONY: publish-plugins-oci
publish-plugins-oci: ## Publish CLI Plugins as OCI image for current host os-arch only
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type oci --plugins "$(STANDALONE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/standalone:v0.0.1 --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go publish --type oci --plugins "$(CONTEXT_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/context:v0.0.1 --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)

.PHONY: build-publish-plugins-all-local
build-publish-plugins-all-local: clean-catalog-cache clean-cli-plugins build-cli publish-plugins-all-local ## Build and Publish CLI Plugins locally for all supported os-arch
	
.PHONY: build-publish-plugins-local
build-publish-plugins-local: clean-catalog-cache clean-cli-plugins build-cli-local publish-plugins-local ## Build and Publish CLI Plugins locally for current host os-arch only

.PHONY: build-publish-plugins-all-oci
build-publish-plugins-all-oci: clean-catalog-cache clean-cli-plugins build-cli publish-plugins-all-oci ## Build and Publish CLI Plugins as OCI image for all supported os-arch
	
.PHONY: build-publish-plugins-oci
build-publish-plugins-oci: clean-catalog-cache clean-cli-plugins build-cli-local publish-plugins-oci ## Build and Publish CLI Plugins as OCI image for current host os-arch only

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

	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin/${OS}/${ARCH}/cli --target ${OS}_${ARCH}
	./hack/embed-pinniped-binary.sh go ${OS} ${ARCH} ${PINNIPED_VERSIONS}
	$(GO) run ./cmd/cli/plugin-admin/builder/main.go cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --corepath "cmd/cli/tanzu" --artifacts artifacts/${OS}/${ARCH}/cli --target  ${OS}_${ARCH}

## --------------------------------------
## Testing, verification, formating and cleanup
## --------------------------------------

.PHONY: test
test: generate fmt vet manifests build-cli-mocks ## Run tests
	## Skip running TKG integration tests
	$(MAKE) ytt -C $(TOOLS_DIR)

	## Test the YTT cluster templates
	echo "Changing into the provider test directory to verify ytt cluster templates..."
	cd ./pkg/v1/providers/tests/unit && PATH=$(abspath hack/tools/bin):"$(PATH)" $(GO) test -v -timeout 30s ./
	echo "... ytt cluster template verification complete!"

	PATH=$(abspath hack/tools/bin):"$(PATH)" $(GO) test -coverprofile cover.out -v `go list ./... | grep -v github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/test`


	$(MAKE) kubebuilder -C $(TOOLS_DIR)
	KUBEBUILDER_ASSETS=$(ROOT_DIR)/$(KUBEBUILDER)/bin $(MAKE) test -C addons

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

	# Linting for the addons...
	$(GOLANGCI_LINT) run -v
	cd $(ADDONS_DIR); $(GOLANGCI_LINT) run -v
	cd $(ADDONS_DIR)/pinniped/post-deploy/; $(GOLANGCI_LINT) run -v

	# Linting for the YTT generation test code...
	cd $(YTT_TESTS_DIR); $(GOLANGCI_LINT) run -v

	# Check licenses in shell scripts and Makefile
	hack/check-license.sh

doc-lint: tools ## Run linting checks for docs
	$(VALE) --config=.vale/config.ini --glob='*.md' ./
	# mdlint rules with possible errors and fixes can be found here:
	# https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md
	# Additional configuration can be found in the .markdownlintrc file.
	hack/check-mdlint.sh

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	$(GO) mod tidy
	cd $(ADDONS_DIR); $(GO) mod tidy
	cd $(ADDONS_DIR)/pinniped/post-deploy/; $(GO) mod tidy
	cd $(TOOLS_DIR); $(GO) mod tidy
	cd $(YTT_TESTS_DIR); $(GO) mod tidy

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

## --------------------------------------
## Docker build
## --------------------------------------

# These are the components in this repo that need to have a docker image built.
# This variable refers to directory paths that contain a Makefile with `docker-build`, `docker-publish` and
# `kbld-image-replace` targets that can build and push a docker image for that component.
COMPONENTS := cliplugins pkg/v1/sdk/features

.PHONY: docker-build
docker-build: TARGET=docker-build
docker-build: $(COMPONENTS)

.PHONY: docker-publish
docker-publish: TARGET=docker-publish
docker-publish: $(COMPONENTS)

.PHONY: kbld-image-replace
kbld-image-replace: TARGET=kbld-image-replace
kbld-image-replace: $(COMPONENTS)

.PHONY: $(COMPONENTS)
$(COMPONENTS):
	$(MAKE) -C $@ $(TARGET)

.PHONY: docker-all
docker-all: docker-build docker-publish kbld-image-replace

## --------------------------------------
## Packages
## --------------------------------------

.PHONY: create-package
create-package: ## Stub out new package directories and manifests. Usage: make create-management-package PACKAGE_NAME=foobar
	@hack/packages/scripts/create-package.sh $(PACKAGE_REPOSITORY) $(PACKAGE_NAME)

.PHONY: package-bundle
package-bundle: ## Build one specific tar bundle package, needs PACKAGE_NAME VERSION
	PACKAGE_REPOSITORY=$(PACKAGE_REPOSITORY) PACKAGE_NAME=$(PACKAGE_NAME) $(PACKAGES_SCRIPTS_DIR)/package-utils.sh generate_single_imgpkg_lock_output
	PACKAGE_REPOSITORY=$(PACKAGE_REPOSITORY) PACKAGE_NAME=$(PACKAGE_NAME) PACKAGE_SUB_VERSION=$(PACKAGE_SUB_VERSION) $(PACKAGES_SCRIPTS_DIR)/package-utils.sh create_single_package_bundle

.PHONY: package-bundles
package-bundles: management-package-bundles ## Build tar bundles for multiple packages

.PHONY: management-package-bundles
management-package-bundles: tools management-imgpkg-lock-output ## Build tar bundles for packages
	PACKAGE_REPOSITORY="management" $(PACKAGES_SCRIPTS_DIR)/package-utils.sh create_package_bundles localhost:5000

.PHONY: package-repo-bundle
package-repo-bundle: ## Build tar bundles for package repo with given package-values.yaml file
	PACKAGE_REPOSITORY=$(PACKAGE_REPOSITORY) REGISTRY=$(OCI_REGISTRY)/packages/$(PACKAGE_REPOSITORY) PACKAGE_VALUES_FILE=$(PACKAGE_VALUES_FILE) $(PACKAGES_SCRIPTS_DIR)/package-utils.sh create_package_repo_bundles

.PHONY: push-package-bundles
push-package-bundles: push-management-package-bundles  ## Push package bundles

.PHONY: push-package-repo-bundles
push-package-repo-bundles: push-management-package-repo-bundle ## Push package repo bundles

.PHONY: push-management-package-bundles
push-management-package-bundles: tools ## Push management package bundles
	PACKAGE_REPOSITORY="management" REGISTRY=$(OCI_REGISTRY)/packages/management $(PACKAGES_SCRIPTS_DIR)/package-utils.sh push_package_bundles

.PHONY: push-management-package-repo-bundle
push-management-package-repo-bundle: tools ## Push management package repo bundles
	PACKAGE_REPOSITORY="management" REGISTRY=$(OCI_REGISTRY)/packages/management $(PACKAGES_SCRIPTS_DIR)/package-utils.sh push_package_repo_bundles

.PHONY: management-imgpkg-lock-output
management-imgpkg-lock-output: tools ## Generate imgpkg lock output for packages
	PACKAGE_REPOSITORY="management" $(PACKAGES_SCRIPTS_DIR)/package-utils.sh generate_imgpkg_lock_output

.PHONY: clean-registry
clean-registry: ## Stops and removes local docker registry
	docker container stop registry && docker container rm -v registry || true

.PHONY: local-registry
local-registry: clean-registry ## Starts up a local docker registry
	docker run -d -p 5000:5000 --name registry mirror.gcr.io/library/registry:2

.PHONY: trivy-scan
trivy-scan: ## Trivy scan images used in packages
	make -C $(TOOLS_DIR) trivy
	$(PACKAGES_SCRIPTS_DIR)/package-utils.sh trivy_scan

.PHONY: management-package-vendir-sync
management-package-vendir-sync: ## Performs a `vendir sync` for each management package
	@cd packages/management && for package in *; do\
		printf "\n===> syncing $${package}\n";\
		pushd $${package}/bundle;\
		$(TOOLS_BIN_DIR)/vendir sync >> /dev/null;\
		popd;\
	done
