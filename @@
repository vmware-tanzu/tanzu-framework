# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ./common.mk

SHELL := /usr/bin/env bash

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
ADDONS_DIR := addons
YTT_TESTS_DIR := providers/tests
PACKAGES_SCRIPTS_DIR := $(abspath hack/packages/scripts)
UI_DIR := tkg/web
GO_MODULES=$(shell find . -path "*/go.mod" | grep -v "^./pinniped" | xargs -I _ dirname _)
PROVIDER_BUNDLE_ZIP = providers/client/manifest/providers.zip
TKG_PROVIDER_BUNDLE_ZIP = tkg/tkgctl/client/manifest/providers.zip

PINNIPED_GIT_REPOSITORY = https://github.com/vmware-tanzu/pinniped.git
PINNIPED_VERSIONS = v0.4.4 v0.12.1

ifndef IS_OFFICIAL_BUILD
IS_OFFICIAL_BUILD = ""
endif

ifndef TANZU_PLUGIN_UNSTABLE_VERSIONS
TANZU_PLUGIN_UNSTABLE_VERSIONS = "experimental"
endif

# NPM registry to use for downloading node modules for UI build
CUSTOM_NPM_REGISTRY ?= $(shell git config tkg.npmregistry)

# TKG Compatibility Image repo and path related configuration
# These set the defaults after a fresh install in ~/.config/tanzu/config.yaml
# Users can change these values by running commands like:
# tanzu config set cli.edition tce
ifndef TKG_DEFAULT_IMAGE_REPOSITORY
TKG_DEFAULT_IMAGE_REPOSITORY = "projects-stg.registry.vmware.com/tkg"
endif
ifndef TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH
# TODO change it to "tkg-compatibility" once the image is pushed to registry
TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH = "framework-zshippable/tkg-compatibility"
endif

ifndef ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY
ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY = "true"
endif
ifndef DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH
DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH = "packages/standalone-plugins"
endif
ifndef DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG
DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG = "${BUILD_VERSION}"
endif
ifndef DEFAULT_STANDALONE_DISCOVERY_TYPE
DEFAULT_STANDALONE_DISCOVERY_TYPE = "local"
endif
ifndef DEFAULT_STANDALONE_DISCOVERY_LOCAL_PATH
DEFAULT_STANDALONE_DISCOVERY_LOCAL_PATH = "standalone"
endif
ifndef TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES
TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES = "projects-stg.registry.vmware.com/tkg"
endif

# Package tooling related variables
PACKAGE_VERSION ?= ${BUILD_VERSION}
REPO_BUNDLE_VERSION ?= ${BUILD_VERSION}

DOCKER_DIR := /app
SWAGGER=docker run --rm -v ${PWD}:${DOCKER_DIR}:$(DOCKER_VOL_OPTS) quay.io/goswagger/swagger:v0.21.0

# OCI registry for hosting tanzu framework components (containers and packages)
OCI_REGISTRY ?= projects.registry.vmware.com/tanzu_framework

.DEFAULT_GOAL:=help

# TODO: Change package path to cli/runtime when this var is moved there.
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.IsOfficialBuild=$(IS_OFFICIAL_BUILD)'

ifneq ($(strip $(TANZU_CORE_BUCKET)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.CoreBucketName=$(TANZU_CORE_BUCKET)'
endif

ifeq ($(TANZU_FORCE_NO_INIT), true)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/command.forceNoInit=true'
endif

ifneq ($(strip $(TKG_DEFAULT_IMAGE_REPOSITORY)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryRepository=$(TKG_DEFAULT_IMAGE_REPOSITORY)'
endif
ifneq ($(strip $(TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultAllowedPluginRepositories=$(TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES)'
endif

ifneq ($(strip $(ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/runtime/config.IsContextAwareDiscoveryEnabled=$(ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryImagePath=$(DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryImageTag=$(DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_LOCAL_PATH)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryLocalPath=$(DEFAULT_STANDALONE_DISCOVERY_LOCAL_PATH)'
endif


BUILD_TAGS ?=

ARTIFACTS_DIR ?= $(ROOT_DIR)/artifacts
ARTIFACTS_ADMIN_DIR ?= $(ROOT_DIR)/artifacts-admin

XDG_CACHE_HOME := ${HOME}/.cache
XDG_CONFIG_HOME := ${HOME}/.config
TANZU_PLUGIN_PUBLISH_PATH ?= $(XDG_CONFIG_HOME)/tanzu-plugins

export XDG_DATA_HOME
export XDG_CACHE_HOME
export XDG_CONFIG_HOME
export OCI_REGISTRY


all: manager ui-build build-cli

manager: generate ## Build manager binary
	$(GO) build -ldflags "$(LD_FLAGS)" -o bin/manager main.go

run: generate manifests ## Run against the configured Kubernetes cluster in ~/.kube/config
	$(GO) run -ldflags "$(LD_FLAGS)" ./main.go

install: manifests tools ## Install CRDs into a cluster
	kustomize build config/crd | kubectl apply -f -

uninstall: manifests ## Uninstall CRDs from a cluster
	kustomize build config/crd | kubectl delete -f -

deploy: manifests tools ## Deploy controller in the configured Kubernetes cluster in ~/.kube/config
	cd config/manager && kustomize edit set image controller=${IMG}
	kustomize build config/default | kubectl apply -f -

manifests: ## Generate manifests e.g. CRD, RBAC etc.
	$(MAKE) generate-manifests CONTROLLER_GEN_SRC=./apis/...
	$(MAKE) -C apis/cli generate-manifests CONTROLLER_GEN_SRC=./...
	$(MAKE) -C apis/config generate-manifests CONTROLLER_GEN_SRC=./...
	$(MAKE) -C cli/runtime generate-manifests

generate-go: $(COUNTERFEITER) ## Generate code via go generate.
	PATH=$(abspath hack/tools/bin):"$(PATH)" go generate ./...

generate: tools ## Generate code (legacy)
	$(MAKE) generate-controller-code
	$(MAKE) -C apis/cli generate-controller-code
	$(MAKE) -C apis/config generate-controller-code
	$(MAKE) -C cli/runtime generate-controller-code

## --------------------------------------
##@ Version
## --------------------------------------

.PHONY: version
version: ## Show version
	@echo $(BUILD_VERSION)

## --------------------------------------
##@ Build prerequisites
## --------------------------------------

.PHONY: ensure-pinniped-repo
ensure-pinniped-repo: ## Clone Pinniped
	@rm -rf pinniped
	@mkdir -p pinniped
	@GIT_TERMINAL_PROMPT=0 git clone -q ${PINNIPED_GIT_REPOSITORY} pinniped > ${NUL} 2>&1

.PHONY: prep-build-cli
prep-build-cli: ensure-pinniped-repo  ## Prepare for building the CLI
	$(GO) mod download
	$(GO) mod tidy -compat=${GOVERSION}
	EMBED_PROVIDERS_TAG=embedproviders
ifeq "${BUILD_TAGS}" "${EMBED_PROVIDERS_TAG}"
	make -C providers -f Makefile generate-provider-bundle-zip
	cp -f ${PROVIDER_BUNDLE_ZIP} $(TKG_PROVIDER_BUNDLE_ZIP)
endif

.PHONY: configure-buildtags-%
configure-buildtags-%: ## Configure build tags
ifeq ($(strip $(BUILD_TAGS)),)
	$(eval TAGS = $(word 1,$(subst -, ,$*)))
	$(eval BUILD_TAGS=$(TAGS))
endif
	@echo "BUILD_TAGS set to '$(BUILD_TAGS)'"

## --------------------------------------
##@ Build binaries and plugins
## --------------------------------------

# Dynamically generate the OS-ARCH targets to allow for parallel execution
CLI_JOBS_OCI_DISCOVERY := $(addprefix build-cli-oci-,${ENVS})
CLI_ADMIN_JOBS_OCI_DISCOVERY := $(addprefix build-plugin-admin-oci-,${ENVS})

CLI_JOBS_LOCAL_DISCOVERY := $(addprefix build-cli-local-,${ENVS})
CLI_ADMIN_JOBS_LOCAL_DISCOVERY := $(addprefix build-plugin-admin-local-,${ENVS})

LOCAL_PUBLISH_PLUGINS_JOBS := $(addprefix publish-plugins-local-,$(ENVS))


RELEASE_JOBS := $(addprefix release-,${ENVS})

BUILDER := $(ROOT_DIR)/bin/builder
BUILDER_SRC := $(shell find cmd/cli/plugin-admin/builder -type f -print)
$(BUILDER): $(BUILDER_SRC)
	cd cmd/cli/plugin-admin/builder && $(GO) build -o $(BUILDER) .

.PHONY: prepare-builder
prepare-builder: $(BUILDER)

.PHONY: build-cli
build-cli: build-cli-with-local-discovery ## Build Tanzu CLI

.PHONY: build-cli-with-oci-discovery
build-cli-with-oci-discovery: ${CLI_ADMIN_JOBS_OCI_DISCOVERY} ${CLI_JOBS_OCI_DISCOVERY} publish-plugins-all-oci publish-admin-plugins-all-oci ## Build Tanzu CLI with OCI standalone discovery
	@rm -rf pinniped

.PHONY: build-cli-with-local-discovery
build-cli-with-local-discovery: ${CLI_ADMIN_JOBS_LOCAL_DISCOVERY} ${CLI_JOBS_LOCAL_DISCOVERY} publish-plugins-all-local publish-admin-plugins-all-local ## Build Tanzu CLI with Local standalone discovery
	@rm -rf pinniped

.PHONY: build-plugin-admin-with-oci-discovery
build-plugin-admin-with-oci-discovery: ${CLI_ADMIN_JOBS_OCI_DISCOVERY} publish-admin-plugins-all-oci ## Build Tanzu CLI admin plugins with OCI standalone discovery

.PHONY: build-plugin-admin-with-local-discovery
build-plugin-admin-with-local-discovery: ${CLI_ADMIN_JOBS_LOCAL_DISCOVERY} publish-admin-plugins-all-local ## Build Tanzu CLI admin plugins with Local standalone discovery

.PHONY: build-plugin-admin-%
build-plugin-admin-%: prepare-builder
	$(eval ARCH = $(word 3,$(subst -, ,$*)))
	$(eval OS = $(word 2,$(subst -, ,$*)))
	$(eval DISCOVERY_TYPE = $(word 1,$(subst -, ,$*)))

	@if [ "$(filter $(OS)-$(ARCH),$(ENVS))" = "" ]; then\
		printf "\n\n======================================\n";\
		printf "! $(OS)-$(ARCH) is not an officially supported platform!\n";\
		printf "! Make sure to perform a full build to make sure expected plugins are available!\n";\
		printf "======================================\n\n";\
	fi

	@echo build version: $(BUILD_VERSION)
	$(BUILDER) cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=${DISCOVERY_TYPE}'" --tags "${BUILD_TAGS}" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin/${OS}/${ARCH}/cli --target ${OS}_${ARCH}

.PHONY: build-cli-%
build-cli-%: prepare-builder prep-build-cli
	$(eval ARCH = $(word 3,$(subst -, ,$*)))
	$(eval OS = $(word 2,$(subst -, ,$*)))
	$(eval DISCOVERY_TYPE = $(word 1,$(subst -, ,$*)))

	@if [ "$(filter $(OS)-$(ARCH),$(ENVS))" = "" ]; then\
		printf "\n\n======================================\n";\
		printf "! $(OS)-$(ARCH) is not an officially supported platform!\n";\
		printf "! Make sure to perform a full build to make sure expected plugins are available!\n";\
		printf "======================================\n\n";\
	fi

	./hack/embed-pinniped-binary.sh go ${OS} ${ARCH} ${PINNIPED_VERSIONS}
	$(BUILDER) cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=${DISCOVERY_TYPE}'" --tags "${BUILD_TAGS}" --path "cmd/cli/plugin" --artifacts artifacts/${OS}/${ARCH}/cli --target  ${OS}_${ARCH}
	$(MAKE) build-tanzu-core-cli-$(DISCOVERY_TYPE)-$(OS)-$(ARCH) -C cli/core

## --------------------------------------
##@ Build locally
## --------------------------------------

# Building CLI and plugins locally with `make build-cli-local` is different in 2 ways compared to official build
# 1. It uses `local` file-system based standalone-discovery for plugin discovery and installation
#    whereas official build uses `OCI` based standalone-discovery for plugin discovery and installation
# 2. On official build, provider templates are published as an OCI image and consumed from BoM file
#    but with `build-cli-local`, it ensures that the cluster and management-cluster plugins are build
#    with embedded provider templates. This is used only for dev build and not for production builds.
#    When using embedded providers, `~/.config/tanzu/tkg/providers` directory always gets overwritten with the
#    embdedded providers. To skip the provider updates, specify `SUPPRESS_PROVIDERS_UPDATE` environment variable.
#    Note: If any local builds want to skip embedding providers and want utilize providers from TKG BoM file,
#    To skip provider embedding, pass `BUILD_TAGS=skipembedproviders` to make target (`make BUILD_TAGS=skipembedproviders build-cli-local)
.PHONY: build-cli-local
build-cli-local: prepare-builder configure-buildtags-embedproviders build-cli-local-${GOHOSTOS}-${GOHOSTARCH} publish-plugins-local ## Build Tanzu CLI with local standalone discovery. cluster and management-cluster plugins are built with embedded providers.
	$(BUILDER) cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=local'" --tags "${BUILD_TAGS}" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin/${GOHOSTOS}/${GOHOSTARCH}/cli --target local
	$(MAKE) publish-admin-plugins-local
.PHONY: build-install-cli-local
build-install-cli-local: clean-catalog-cache clean-cli-plugins build-cli-local install-cli-plugins install-cli ## Local build and install the CLI plugins with local standalone discovery

## --------------------------------------
##@ Build and publish CLI plugin discovery resource files and binaries
## --------------------------------------

.PHONY: publish-plugins-all-local
publish-plugins-all-local: prepare-builder ## Publish CLI plugins locally for all supported os-arch
	$(BUILDER) publish --type local --plugins "$(PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/discovery/standalone" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/distribution" --input-artifact-dir $(ARTIFACTS_DIR)

.PHONY: publish-admin-plugins-all-local
publish-admin-plugins-all-local: prepare-builder ## Publish CLI admin plugins locally for all supported os-arch
	$(BUILDER) publish --type local --plugins "$(ADMIN_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/discovery/admin" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/distribution" --input-artifact-dir $(ARTIFACTS_ADMIN_DIR)

.PHONY: publish-plugins-local
publish-plugins-local: prepare-builder ## Publish CLI plugins locally for current host os-arch only
	$(BUILDER) publish --type local --plugins "$(PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/discovery/standalone" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/distribution" --input-artifact-dir $(ARTIFACTS_DIR)

.PHONY: publish-plugins-local-%
publish-plugins-local-%: prepare-builder ## Publish CLI plugins to local directory that can be shared. Configure TANZU_PLUGIN_PUBLISH_PATH, PLUGINS, ARTIFACTS_DIR, DISCOVERY_NAME variables
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))
	$(BUILDER) publish --type local --plugins "$(PLUGINS)" --version $(BUILD_VERSION) --os-arch "${OS}-${ARCH}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH}-$(DISCOVERY_NAME)/discovery/$(DISCOVERY_NAME)" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/${OS}-${ARCH}-$(DISCOVERY_NAME)/distribution" --input-artifact-dir $(ARTIFACTS_DIR)

.PHONY: publish-plugins-local-generic
publish-plugins-local-generic: ${LOCAL_PUBLISH_PLUGINS_JOBS} ## Publish CLI plugins to local directory that can be shared. Configure TANZU_PLUGIN_PUBLISH_PATH, PLUGINS, ARTIFACTS_DIR, DISCOVERY_NAME variables

.PHONY: publish-admin-plugins-local
publish-admin-plugins-local: prepare-builder ## Publish CLI admin plugins locally for current host os-arch only
	$(BUILDER) publish --type local --plugins "$(ADMIN_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --local-output-discovery-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/discovery/admin" --local-output-distribution-dir "$(TANZU_PLUGIN_PUBLISH_PATH)/distribution" --input-artifact-dir $(ARTIFACTS_ADMIN_DIR)


.PHONY: publish-plugins-all-oci
publish-plugins-all-oci: prepare-builder ## Publish CLI plugins as OCI image for all supported os-arch
	$(BUILDER) publish --type oci --plugins "$(STANDALONE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/standalone:${BUILD_VERSION} --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)
	$(BUILDER) publish --type oci --plugins "$(CONTEXTAWARE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/context:${BUILD_VERSION} --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)

.PHONY: publish-admin-plugins-all-oci
publish-admin-plugins-all-oci: prepare-builder ## Publish CLI admin plugins as OCI image for all supported os-arch
	$(BUILDER) publish --type oci --plugins "$(ADMIN_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${ENVS}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/admin:${BUILD_VERSION} --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_ADMIN_DIR)

.PHONY: publish-plugins-oci
publish-plugins-oci: prepare-builder ## Publish CLI plugins as OCI image for current host os-arch only
	$(BUILDER) publish --type oci --plugins "$(STANDALONE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/standalone:${BUILD_VERSION} --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)
	$(BUILDER) publish --type oci --plugins "$(CONTEXTAWARE_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/context:${BUILD_VERSION} --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_DIR)

.PHONY: publish-admin-plugins-oci
publish-admin-plugins-oci: prepare-builder ## Publish CLI admin plugins as OCI image for current host os-arch only
	$(BUILDER) publish --type oci --plugins "$(ADMIN_PLUGINS)" --version $(BUILD_VERSION) --os-arch "${GOHOSTOS}-${GOHOSTARCH}" --oci-discovery-image ${OCI_REGISTRY}/tanzu-plugins/discovery/admin:${BUILD_VERSION} --oci-distribution-image-repository ${OCI_REGISTRY}/tanzu-plugins/distribution/ --input-artifact-dir $(ROOT_DIR)/$(ARTIFACTS_ADMIN_DIR)


.PHONY: build-publish-plugins-all-local
build-publish-plugins-all-local: clean-catalog-cache clean-cli-plugins build-cli-with-local-discovery ## Build and Publish CLI plugins locally with local standalone discovery for all supported os-arch

.PHONY: build-publish-plugins-local
build-publish-plugins-local: clean-catalog-cache clean-cli-plugins build-cli-local ## Build and publish CLI Plugins locally with local standalone discovery for current host os-arch only

.PHONY: build-publish-plugins-all-oci
build-publish-plugins-all-oci: clean-catalog-cache clean-cli-plugins build-cli publish-plugins-all-oci ## Build and Publish CLI Plugins as OCI image for all supported os-arch

## --------------------------------------
##@ Manage CLI mocks
## --------------------------------------

.PHONY: build-cli-mocks
build-cli-mocks: prepare-builder ## Build Tanzu CLI mocks
	$(BUILDER) cli compile $(addprefix --target ,$(subst -,_,${ENVS})) --version 0.0.1 --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./test/cli/mock/plugin-old --artifacts ./test/cli/mock/artifacts-old
	$(BUILDER) cli compile $(addprefix --target ,$(subst -,_,${ENVS})) --version 0.0.2 --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./test/cli/mock/plugin-new --artifacts ./test/cli/mock/artifacts-new
	$(BUILDER) cli compile $(addprefix --target ,$(subst -,_,${ENVS})) --version 0.0.3 --ldflags "$(LD_FLAGS)" --tags "${BUILD_TAGS}" --path ./test/cli/mock/plugin-alt --artifacts ./test/cli/mock/artifacts-alt

## --------------------------------------
##@ Install binaries and plugins
## --------------------------------------

.PHONY: install-cli
install-cli: install-cli-local ## Install Tanzu CLI with local discovery

.PHONY: install-cli-%
install-cli-%: ## Install Tanzu CLI
	$(eval DISCOVERY_TYPE = $(word 1,$(subst -, ,$*)))
	$(MAKE) install-tanzu-core-cli-$(DISCOVERY_TYPE) -C cli/core

.PHONY: install-cli-plugins
install-cli-plugins: ## Install Tanzu CLI plugins
	@if [ "${ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY}" = "true" ]; then \
		$(MAKE) install-cli-plugins-from-local-discovery ; \
	else \
		$(MAKE) install-cli-plugins-without-discovery ; \
	fi

.PHONY: install-cli-plugins-without-discovery
install-cli-plugins-without-discovery: set-unstable-versions set-context-aware-cli-for-plugins ## Install Tanzu CLI plugins when context-aware discovery is disabled
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/tanzu/main.go \
		plugin install all --local $(ARTIFACTS_DIR)-admin/$(GOHOSTOS)/$(GOHOSTARCH)/cli
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/tanzu/main.go \
		test fetch --local $(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli --local $(ARTIFACTS_DIR)-admin/$(GOHOSTOS)/$(GOHOSTARCH)/cli

.PHONY: install-cli-plugins-from-local-discovery
install-cli-plugins-from-local-discovery: clean-catalog-cache clean-cli-plugins set-context-aware-cli-for-plugins configure-admin-plugins-discovery-source-local ## Install Tanzu CLI plugins from local discovery
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=local'" ./cmd/tanzu/main.go plugin sync

.PHONY: install-cli-plugins-from-oci-discovery
install-cli-plugins-from-oci-discovery: clean-catalog-cache clean-cli-plugins set-context-aware-cli-for-plugins ## Install Tanzu CLI plugins from OCI discovery
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=oci'" ./cmd/tanzu/main.go plugin sync

.PHONY: set-unstable-versions
set-unstable-versions: ## Configures the unstable versions
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/tanzu/main.go config set unstable-versions $(TANZU_PLUGIN_UNSTABLE_VERSIONS)

.PHONY: set-context-aware-cli-for-plugins
set-context-aware-cli-for-plugins: ## Configures the context-aware-cli-for-plugins-beta feature flag
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/tanzu/main.go config set features.global.context-aware-cli-for-plugins $(ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY)

.PHONY: configure-admin-plugins-discovery-source-local
configure-admin-plugins-discovery-source-local: ## Configures the admin plugins discovery source
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/tanzu/main.go plugin source add --name admin-local --type local --uri admin || true
	cd ./cli/core && TANZU_CLI_NO_INIT=true $(GO) run -ldflags "$(LD_FLAGS)" ./cmd/tanzu/main.go plugin source update admin-local --type local --uri admin || true

.PHONY: build-install-cli-all ## Build and install the CLI plugins
build-install-cli-all: build-install-cli-all-with-local-discovery ## Build and install Tanzu CLI plugins

.PHONY: build-install-cli-all-with-local-discovery ## Build and install the CLI plugins with local standalone discovery
build-install-cli-all-with-local-discovery: clean-catalog-cache clean-cli-plugins build-cli-with-local-discovery install-cli-plugins-from-local-discovery install-cli-local ## Build and install Tanzu CLI plugins

.PHONY: build-install-cli-all-with-oci-discovery ## Build and install the CLI plugins with oci standalone discovery
build-install-cli-all-with-oci-discovery: clean-catalog-cache clean-cli-plugins build-cli-with-oci-discovery install-cli-plugins-from-oci-discovery install-cli-oci ## Build and install Tanzu CLI plugins

# This target is added as some tests still relies on tkg cli.
# TODO: Remove this target when all tests are migrated to use tanzu cli
.PHONY: tkg-cli ## Builds TKG-CLI binary
tkg-cli: configure-buildtags-embedproviders configure-bom prep-build-cli ## Build tkg CLI binary only, and without rebuilding ui bits (providers are embedded to the binary)
	GO111MODULE=on $(GO) build -o $(BIN_DIR)/tkg-${GOHOSTOS}-${GOHOSTARCH} --gcflags "${GC_FLAGS}" -ldflags "${LD_FLAGS}" -tags "${BUILD_TAGS}" cmd/cli/tkg/main.go

.PHONY: build-cli-image
build-cli-image: ## Build the CLI image
	docker build -t projects.registry.vmware.com/tanzu/cli:latest -f Dockerfile.cli .

## --------------------------------------
##@ Release binaries
## --------------------------------------

# TODO (pbarker): should work this logic into the builder plugin
.PHONY: release
release: ensure-pinniped-repo ${RELEASE_JOBS} ## Create release binaries
	@rm -rf pinniped

.PHONY: release-%
release-%: prepare-builder ## Create release for a platform
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))

	$(BUILDER) cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=oci'" --tags "${BUILD_TAGS}" --path ./cmd/cli/plugin-admin --artifacts artifacts-admin/${OS}/${ARCH}/cli --target ${OS}_${ARCH}
	./hack/embed-pinniped-binary.sh go ${OS} ${ARCH} ${PINNIPED_VERSIONS}
	$(BUILDER) cli compile --version $(BUILD_VERSION) --ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=oci'" --tags "${BUILD_TAGS}" --path "cmd/cli/plugin" --artifacts artifacts/${OS}/${ARCH}/cli --target  ${OS}_${ARCH}
	$(MAKE) build-tanzu-core-cli-oci-$(OS)-$(ARCH) -C cli/core

## --------------------------------------
##@ Testing, verification, formating and cleanup
## --------------------------------------

.PHONY: test
test: generate manifests build-cli-mocks ## Run tests
	## Skip running TKG integration tests
	$(MAKE) ytt -C $(TOOLS_DIR)

	echo "Verifying cluster-api packages and providers are in sync..."
	make -C hack/providers-sync-tools validate
	echo "... cluster-api packages are in sync"

	## Test the YTT cluster templates
	echo "Changing into the provider test directory to verify ytt cluster templates..."
	cd ./providers/tests/unit && PATH=$(abspath hack/tools/bin):"$(PATH)" $(GO) test -coverprofile coverage1.txt -v -timeout 120s ./
	echo "... ytt cluster template verification complete!"

	echo "Verifying package tests..."
	find ./packages/ -name "test" -type d | \
		xargs -n1  -I {} bash -c 'cd {} && PATH=$(abspath hack/tools/bin):"$(PATH)" $(GO) test -coverprofile coverage2.txt -v -timeout 120s ./...' \;
	echo "... package tests complete!"

	PATH=$(abspath hack/tools/bin):"$(PATH)" $(GO) test -coverprofile coverage3.txt -v `go list ./... | grep -Ev '(github.com/vmware-tanzu/tanzu-framework/tkg/test|github.com/vmware-tanzu/tanzu-framework/cmd/cli/plugin/package/test)'`

	$(MAKE) kubebuilder -C $(TOOLS_DIR)
	KUBEBUILDER_ASSETS=$(ROOT_DIR)/$(KUBEBUILDER)/bin $(MAKE) test -C addons

	# pinniped post-deploy
	$(MAKE) test -C pinniped-components/post-deploy
	# pinniped tanzu-auth-controller-manager
	pinniped-components/tanzu-auth-controller-manager/hack/test.sh

	#Test core cli runtime library
	$(MAKE) test -C cli/runtime

	#Test core cli
	$(MAKE) test -C cli/core

	#Test tkg module
	$(MAKE) test -C tkg

	# Test feature gates
	$(MAKE) test -C featuregates

.PHONY: test-cli
test-cli: build-cli-mocks ## Run tests
	$(GO) test  ./pkg/v1/auth/... ./pkg/v1/builder/...  ./pkg/v1/encoding/... ./pkg/v1/grpc/...
	#Test core cli
	$(MAKE) test -C cli/core

lint: tools go-lint doc-lint misspell yamllint ## Run linting and misspell checks
	# Check licenses in shell scripts and Makefiles
	hack/check-license.sh

misspell:
	hack/check/misspell.sh

yamllint:
	hack/check/check-yaml.sh

go-lint: tools ## Run linting of go source
	@for i in $(GO_MODULES); do \
		echo "-- Linting $$i --"; \
		pushd $${i}; \
		$(GOLANGCI_LINT) run -v --timeout=10m || exit 1; \
		popd; \
	done

	# Prevent use of deprecated ioutils module
	@CHECK=$$(grep -r --include="*.go" --exclude-dir="pinniped" --exclude="zz_generated*" ioutil .); \
	if [ -n "$${CHECK}" ]; then \
		echo "ioutil is deprecated, use io or os replacements"; \
		echo "https://go.dev/doc/go1.16#ioutil"; \
		echo "$${CHECK}"; \
		exit 1; \
	fi

doc-lint: tools ## Run linting checks for docs
	$(VALE) --config=.vale/config.ini --glob='*.md' ./
	# mdlint rules with possible errors and fixes can be found here:
	# https://github.com/DavidAnson/markdownlint/blob/main/doc/Rules.md
	# Additional configuration can be found in the .markdownlintrc file.
	hack/check-mdlint.sh

.PHONY: modules
modules: ## Runs go mod to ensure modules are up to date.
	@for i in $(GO_MODULES); do \
		echo "-- Tidying $$i --"; \
		pushd $${i}; \
		$(GO) mod tidy -compat=${GOVERSION} || exit 1; \
		popd; \
	done

.PHONY: verify
verify: modules ## Run all verification scripts
verify: ## Run all verification scripts
	./packages/tkg-clusterclass/hack/sync-cc.sh
	./hack/verify-dirty.sh

.PHONY: clean-catalog-cache
clean-catalog-cache: ## Cleans catalog cache
	@rm -rf ${XDG_CACHE_HOME}/tanzu/*

.PHONY: clean-cli-plugins
clean-cli-plugins: ## Remove Tanzu CLI plugins
	- rm -rf ${XDG_DATA_HOME}/tanzu-cli/*

## --------------------------------------
##@ UI Build & Test
## --------------------------------------
.PHONY: update-npm-registry
update-npm-registry: ## set alternate npm registry
ifneq ($(strip $(CUSTOM_NPM_REGISTRY)),)
	npm config set registry $(CUSTOM_NPM_REGISTRY) web
endif

.PHONY: ui-dependencies
ui-dependencies: update-npm-registry  ## install UI dependencies (node modules)
	cd $(UI_DIR); NG_CLI_ANALYTICS=ci npm ci --legacy-peer-deps; cd ../

.PHONY: ui-build
ui-build: ui-dependencies ## Install dependencies, then compile client UI for production
	cd $(UI_DIR); npm run build:prod; cd ../
	$(MAKE) generate-ui-bindata

.PHONY: ui-build-and-test
ui-build-and-test: ui-dependencies ## Compile client UI for production and run tests
	cd $(UI_DIR); npm run build:ci; cd ../
	$(MAKE) generate-ui-bindata

.PHONY: verify-ui-bindata
verify-ui-bindata: ## Run verification for UI bindata
	git diff --exit-code tkg/manifest/server/zz_generated.bindata.go

## --------------------------------------
##@ Generate files
## --------------------------------------

.PHONY: cobra-docs
cobra-docs:
	cd cli/core && TANZU_CLI_NO_INIT=true TANZU_CLI_NO_COLOR=true $(GO) run ./cmd/tanzu generate-all-docs --docs-dir "$(ROOT_DIR)/docs/cli/commands"
	sed -i.bak -E 's/\/[A-Za-z]*\/([a-z]*)\/.config\/tanzu\/pinniped\/sessions.yaml/~\/.config\/tanzu\/pinniped\/sessions.yaml/g' docs/cli/commands/tanzu_pinniped-auth_login.md


.PHONY: generate-fakes
generate-fakes: ## Generate fakes for writing unit tests
	$(GO) generate ./...
	$(MAKE) fmt

.PHONY: generate-ui-bindata
generate-ui-bindata: $(GOBINDATA) ## Generate go-bindata for ui files
	$(GOBINDATA) -mode=420 -modtime=1 -o=tkg/manifest/server/zz_generated.bindata.go -pkg=server $(UI_DIR)/dist/...
	$(MAKE) fmt

.PHONY: generate-telemetry-bindata
generate-telemetry-bindata: $(GOBINDATA) ## Generate telemetry bindata
	$(GOBINDATA) -mode=420 -modtime=1 -o=tkg/manifest/telemetry/zz_generated.bindata.go -pkg=telemetry tkg/manifest/telemetry/...
	$(MAKE) fmt

 # TODO: Remove bindata dependency and use go embed
.PHONY: generate-bindata ## Generate go-bindata files
generate-bindata: generate-telemetry-bindata generate-ui-bindata

.PHONY: configure-bom
configure-bom: ## Configure bill of materials
	# Update default BoM Filename variable in tkgconfig pkg
	sed "s+TKG_DEFAULT_IMAGE_REPOSITORY+${TKG_DEFAULT_IMAGE_REPOSITORY}+g"  hack/update-bundled-bom-filename/update-bundled-default-bom-files-configdata.txt | \
	sed "s+TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH+${TKG_DEFAULT_COMPATIBILITY_IMAGE_PATH}+g" | \
	sed "s+TKG_MANAGEMENT_CLUSTER_PLUGIN_VERSION+${BUILD_VERSION}+g"  > tkg/tkgconfigpaths/zz_bundled_default_bom_files_configdata.go

.PHONY: generate-ui-swagger-api
generate-ui-swagger-api: ## Generate swagger files for UI backend
	rm -rf ${UI_DIR}/server/client  ${UI_DIR}/server/models ${UI_DIR}/server/restapi/operations
	${SWAGGER} generate server -q -A kickstartUI -t $(DOCKER_DIR)/${UI_DIR}/server -f $(DOCKER_DIR)/${UI_DIR}/api/spec.yaml --exclude-main
	${SWAGGER} generate client -q -A kickstartUI -t $(DOCKER_DIR)/${UI_DIR}/server -f $(DOCKER_DIR)/${UI_DIR}/api/spec.yaml
	# reset the server.go file to avoid goswagger overwritting our custom changes.
	git reset HEAD ${UI_DIR}/server/restapi/server.go
	git checkout HEAD ${UI_DIR}/server/restapi/server.go
	$(MAKE) fmt

.PHONY: clean-generated-conversions
clean-generated-conversions: ## Remove files generated by conversion-gen from the mentioned dirs. Example SRC_DIRS="./api/run/v1alpha1"
	(IFS=','; for i in $(SRC_DIRS); do find $$i -type f -name 'zz_generated.conversion*' -exec rm -f {} \;; done)

.PHONY: generate-go-conversions
generate-go-conversions: $(CONVERSION_GEN) ## Generate conversions go code
	$(CONVERSION_GEN) \
		-v 3 --logtostderr \
		--input-dirs="./apis/run/v1alpha1,./apis/run/v1alpha2" \
		--build-tag=ignore_autogenerated_core \
		--output-base ./ \
		--output-file-base=zz_generated.conversion \
		--go-header-file=./hack/boilerplate.go.txt

.PHONY: generate-package-config ## Generate the default package config CR e.g. make generate-package-config apiGroup=cni.tanzu.vmware.com kind=AntreaConfig version=v1alpha1 tkr=v1.23.3---vmware.1-tkg.1 namespace=tkg-system
generate-package-config:
	@cd addons/config && \
        ./hack/test.sh verifyAddonConfigTemplateForGVR ${apiGroup} ${version} $(shell echo $(kind) | tr A-Z a-z) $(or $(iaas),default) && \
		$(YTT) --ignore-unknown-comments \
			-f templates/${apiGroup}/${version}/$(shell echo $(kind) | tr A-Z a-z).yaml \
			-f testcases/${apiGroup}/${version}/$(shell echo $(kind) | tr A-Z a-z)/$(or $(iaas),default).yaml \
			-v TKR_VERSION=${tkr} -v GLOBAL_NAMESPACE=$(or $(namespace),"tkg-system") ;\

.PHONY: generate-package-secret
generate-package-secret: ## Generate the default package values secret. Usage: make generate-package-secret PACKAGE=pinniped tkr=v1.23.3---vmware.1-tkg.1 iaas=vsphere
	@if [ -z "$(PACKAGE)" ]; then \
		echo "PACKAGE argument required"; \
		exit 1 ;\
	fi

	@if [ $(PACKAGE) == 'pinniped' ]; then \
	  ./pinniped-components/tanzu-auth-controller-manager/hack/generate-package-secret.sh -v tkr=${tkr} -v infrastructure_provider=${iaas} ;\
	elif [ $(PACKAGE) == 'capabilities' ]; then \
	  ./pkg/v1/sdk/capabilities/hack/generate-package-secret.sh -v tkr=${tkr} --data-value-yaml 'rbac.podSecurityPolicyNames=[${psp}]';\
	else \
	  echo "invalid PACKAGE: $(PACKAGE)" ;\
	  exit 1 ;\
	fi

## --------------------------------------
##@ Provider templates/overlays
## --------------------------------------

.PHONY: clustergen
clustergen: ## Generate diff between 'before' and 'after' of cluster configuration outputs using clustergen
	CLUSTERGEN_BASE=${CLUSTERGEN_BASE} make -C providers -f Makefile cluster-generation-diffs

.PHONY: generate-embedproviders
generate-embedproviders: ## Generate provider bundle to be embedded for local testing
	make -C providers -f Makefile generate-provider-bundle-zip
	cp -f ${PROVIDER_BUNDLE_ZIP} $(TKG_PROVIDER_BUNDLE_ZIP)

## --------------------------------------
##@ TKG integration tests
## --------------------------------------

GINKGO_NODES  ?= 1
GINKGO_NOCOLOR ?= false

.PHONY: e2e-tkgctl-docker
e2e-tkgctl-docker: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests for Docker clusters
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders tkg/test/tkgctl/docker

.PHONY: e2e-tkgctl-azure
e2e-tkgctl-azure: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests for Azure clusters
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders tkg/test/tkgctl/azure

.PHONY: e2e-tkgctl-aws
e2e-tkgctl-aws: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests for AWS clusters
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders tkg/test/tkgctl/aws

.PHONY: e2e-tkgctl-vc67
e2e-tkgctl-vc67: $(GINKGO) generate-embedproviders ## Run ginkgo tkgctl E2E tests for Vsphere clusters
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders tkg/test/tkgctl/vsphere67

.PHONY: e2e-packageclient-docker
e2e-packageclient-docker: $(GINKGO) generate-embedproviders ## Run ginkgo packageclient E2E tests for TKG client library
	$(GINKGO) -v -trace -nodes=$(GINKGO_NODES) --noColor=$(GINKGO_NOCOLOR) $(GINKGO_ARGS) -tags embedproviders cmd/cli/plugin/package/test

## --------------------------------------
##@ Docker build
## --------------------------------------

# These are the components in this repo that need to have a docker image built.
# This variable refers to directory paths that contain a Makefile with `docker-build`, `docker-publish` and
# `kbld-image-replace` targets that can build and push a docker image for that component.
COMPONENTS ?=  \
  pkg/v2/tkr/controller/tkr-source \
  pkg/v2/tkr/controller/tkr-status \
  featuregates \
  addons \
  cliplugins \
  pkg/v2/tkr/webhook/infra-machine \
  pkg/v1/sdk/capabilities \
  pkg/v2/tkr/webhook/tkr-conversion \
  pkg/v2/tkr/webhook/cluster/tkr-resolver \
  pinniped-components/tanzu-auth-controller-manager \
  pkg/v2/object-propagation

.PHONY: docker-build
docker-build: TARGET=docker-build
docker-build: $(COMPONENTS) ## Build Docker images

.PHONY: docker-publish
docker-publish: TARGET=docker-publish
docker-publish: $(COMPONENTS) ## Push Docker images

.PHONY: kbld-image-replace
kbld-image-replace: TARGET=kbld-image-replace
kbld-image-replace: $(COMPONENTS) ## Resolve Docker images

.PHONY: $(COMPONENTS)
$(COMPONENTS):
	$(MAKE) -C $@ $(TARGET)

.PHONY: docker-all
docker-all: docker-build docker-publish kbld-image-replace ## Ship Docker images

## --------------------------------------
##@ Packages
## --------------------------------------

.PHONY: create-package
create-package: ## Stub out new package directories and manifests. Usage: make create-package PACKAGE_NAME=foobar
	@hack/packages/scripts/create-package.sh $(PACKAGE_NAME)

.PHONY: prep-package-tools
prep-package-tools:
	cd hack/packages/package-tools && $(GO) mod tidy -compat=${GOVERSION}

.PHONY: package-bundle
package-bundle: tools prep-package-tools ## Build one specific tar bundle package, needs PACKAGE_NAME VERSION
	cd hack/packages/package-tools && $(GO) run main.go package-bundle generate $(PACKAGE_NAME) --thick --version=$(PACKAGE_VERSION) --sub-version=$(PACKAGE_SUB_VERSION) --registry=$(OCI_REGISTRY)

.PHONY: package-bundle-thin
package-bundle-thin: tools prep-package-tools ## Build one specific tar bundle package, needs PACKAGE_NAME VERSION
	cd hack/packages/package-tools && $(GO) run main.go package-bundle generate $(PACKAGE_NAME) --repository=$(PACKAGE_REPOSITORY) --version=$(PACKAGE_VERSION) --sub-version=$(PACKAGE_SUB_VERSION)

.PHONY: package-bundles
package-bundles: tools prep-package-tools ## Build tar bundles for multiple packages
	cd hack/packages/package-tools && $(GO) run main.go package-bundle generate --all --thick --repository=$(PACKAGE_REPOSITORY) --version=$(PACKAGE_VERSION) --sub-version=$(PACKAGE_SUB_VERSION) --registry=$(OCI_REGISTRY)

.PHONY: package-bundles-thin
package-bundles-thin: tools prep-package-tools ## Build tar bundles for multiple packages
	cd hack/packages/package-tools && $(GO) run main.go package-bundle generate --all --repository=$(PACKAGE_REPOSITORY) --version=$(PACKAGE_VERSION) --sub-version=$(PACKAGE_SUB_VERSION)

.PHONY: package-repo-bundle
package-repo-bundle: tools prep-package-tools ## Build tar bundles for package repo with given package-values.yaml file
	cd hack/packages/package-tools && $(GO) run main.go repo-bundle generate --repository=$(PACKAGE_REPOSITORY) --registry=$(OCI_REGISTRY) --package-values-file=$(PACKAGE_VALUES_FILE) --version=$(REPO_BUNDLE_VERSION) --sub-version=$(REPO_BUNDLE_SUB_VERSION)

.PHONY: push-package-bundles
push-package-bundles: tools prep-package-tools ## Push specified package bundle(s) in a package repository.
## Specified package bundles must be set to the PACKAGE_BUNDLES environment variable as comma-separated values
## and must not contain spaces. Example: PACKAGE_BUNDLES=featuregates,core-management-plugins
	cd hack/packages/package-tools && $(GO) run main.go package-bundle push $(PACKAGE_BUNDLES) --registry=$(OCI_REGISTRY) --version=$(BUILD_VERSION) --sub-version=$(PACKAGE_SUB_VERSION)

.PHONY: push-all-package-bundles
push-all-package-bundles: tools prep-package-tools ## Push all package bundles in a package repository
	cd hack/packages/package-tools && $(GO) run main.go package-bundle push --repository=$(PACKAGE_REPOSITORY) --registry=$(OCI_REGISTRY) --version=$(BUILD_VERSION) --sub-version=$(PACKAGE_SUB_VERSION) --all

.PHONY: push-package-repo-bundle
push-package-repo-bundle: tools prep-package-tools ## Push package repo bundles
	cd hack/packages/package-tools && $(GO) run main.go repo-bundle push --repository=$(PACKAGE_REPOSITORY) --registry=$(OCI_REGISTRY) --version=$(REPO_BUNDLE_VERSION) --sub-version=$(REPO_BUNDLE_SUB_VERSION)

.PHONY: package-vendir-sync
package-vendir-sync: tools ## Performs a `vendir sync` for each package in a repository
	cd hack/packages/package-tools && $(GO) run main.go vendir sync --repository=$(PACKAGE_REPOSITORY)

.PHONY: local-registry
local-registry: clean-registry ## Starts up a local docker registry. Local docker registry is used for pushing the package bundle to get the sha256, for using it later when producing repo bundle
	docker run -d -p 5001:5000 --name registry mirror.gcr.io/library/registry:2

.PHONY: clean-registry
clean-registry: ## Stops and removes local docker registry
	docker container stop registry && docker container rm -v registry || true

.PHONY: trivy-scan
trivy-scan: ## Trivy scan images used in packages
	make -C $(TOOLS_DIR) trivy
	$(PACKAGES_SCRIPTS_DIR)/package-utils.sh trivy_scan

.PHONY: package-push-bundles-repo ## Performs build and publishes packages and repo bundles
package-push-bundles-repo: package-bundles push-all-package-bundles package-repo-bundle push-package-repo-bundle
