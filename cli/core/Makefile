# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

ROOT_DIR := $(shell git rev-parse --show-toplevel)
MODULE_ROOT_DIR := $(ROOT_DIR)/cli/core

# Golang specific variables
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOHOSTOS ?= $(shell go env GOHOSTOS)
GOHOSTARCH ?= $(shell go env GOHOSTARCH)
# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
GO := go

NUL = /dev/null
ifeq ($(GOHOSTOS),windows)
	NUL = NUL
endif

BUILD_SHA ?= $$(git describe --match=$(git rev-parse --short HEAD) --always --dirty)
BUILD_DATE ?= $$(date -u +"%Y-%m-%d")
BUILD_VERSION ?= $(shell git describe --tags --abbrev=0 2>$(NUL))

ifeq ($(strip $(BUILD_VERSION)),)
BUILD_VERSION = dev
endif

# Directories
TOOLS_DIR := $(abspath $(MODULE_ROOT_DIR)/hack/tools)
TOOLS_BIN_DIR := $(TOOLS_DIR)/bin

# Add tooling binaries here and in hack/tools/Makefile
GOIMPORTS        := $(TOOLS_BIN_DIR)/goimports
TOOLING_BINARIES := $(GOIMPORTS)

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help (default)
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-28s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m\033[32m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Add supported OS-ARCHITECTURE combinations here
ENVS ?= linux-amd64 windows-amd64 darwin-amd64

LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/buildinfo.Date=$(BUILD_DATE)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/buildinfo.SHA=$(BUILD_SHA)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/buildinfo.Version=$(BUILD_VERSION)'

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


LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/buildinfo.IsOfficialBuild=$(IS_OFFICIAL_BUILD)'
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/tkg/buildinfo.IsOfficialBuild=$(IS_OFFICIAL_BUILD)'

ifneq ($(strip $(TANZU_CORE_BUCKET)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.CoreBucketName=$(TANZU_CORE_BUCKET)'
endif

ifneq ($(strip $(TKG_DEFAULT_IMAGE_REPOSITORY)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryRepository=$(TKG_DEFAULT_IMAGE_REPOSITORY)'
endif
ifneq ($(strip $(TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultAllowedPluginRepositories=$(TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES)'
endif

ifneq ($(strip $(ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.IsContextAwareDiscoveryEnabled=$(ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY)'
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
RELEASE_JOBS := $(addprefix release-,${ENVS})

.PHONY: all
all: build-tanzu-core-cli-local install-tanzu-core-cli-local ## Build and install Tanzu Core CLI

.PHONY: build-tanzu-core-cli-local
build-tanzu-core-cli-local: ##Build the Tanzu Core CLI image for local platform
	@mkdir -p $(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli/core/$(BUILD_VERSION)
	@if [ "$(GOHOSTOS)" = "windows" ]; then \
		$(GO) build -tags "${BUILD_TAGS}" --ldflags "$(LD_FLAGS)"  -o "$(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli/core/$(BUILD_VERSION)/tanzu-core-$(GOHOSTOS)_$(GOHOSTARCH).exe" ./cmd/tanzu/main.go ;\
	else \
		$(GO) build -tags "${BUILD_TAGS}" --ldflags "$(LD_FLAGS)"  -o "$(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli/core/$(BUILD_VERSION)/tanzu-core-$(GOHOSTOS)_$(GOHOSTARCH)" ./cmd/tanzu/main.go ;\
	fi

.PHONY: build-tanzu-core-cli-%
build-tanzu-core-cli-%: ##Build the Tanzu Core CLI image for a platform
	$(eval ARCH = $(word 3,$(subst -, ,$*)))
	$(eval OS = $(word 2,$(subst -, ,$*)))
	$(eval DISCOVERY_TYPE = $(word 1,$(subst -, ,$*)))

	@if [ "$(filter $(OS)-$(ARCH),$(ENVS))" = "" ]; then\
		printf "\n\n======================================\n";\
		printf "! $(OS)-$(ARCH) is not an officially supported platform!\n";\
		printf "! Make sure to perform a full build to make sure expected plugins are available!\n";\
		printf "======================================\n\n";\
	fi
	@mkdir -p $(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli/core/$(BUILD_VERSION)
	@if [ "$(OS)" = "windows" ]; then \
		GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -tags "${BUILD_TAGS}" --ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=${DISCOVERY_TYPE}'"  -o "$(ARTIFACTS_DIR)/$(OS)/$(ARCH)/cli/core/$(BUILD_VERSION)/tanzu-core-$(OS)_$(ARCH).exe" ./cmd/tanzu/main.go;\
	else \
		GOOS=$(OS) GOARCH=$(ARCH) $(GO) build -tags "${BUILD_TAGS}" --ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=${DISCOVERY_TYPE}'"  -o "$(ARTIFACTS_DIR)/$(OS)/$(ARCH)/cli/core/$(BUILD_VERSION)/tanzu-core-$(OS)_$(ARCH)" ./cmd/tanzu/main.go;\
	fi

.PHONY: install-tanzu-core-cli ## Install Tanzu Core CLI with local discovery
install-tanzu-core-cli: install-tanzu-core-cli-local ## Install Tanzu CLI with local discovery

.PHONY: install-tanzu-core-cli-%
install-tanzu-core-cli-%: ## Install Tanzu Core CLI
	$(eval DISCOVERY_TYPE = $(word 1,$(subst -, ,$*)))
	$(GO) install -ldflags "$(LD_FLAGS) -X 'github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/config.DefaultStandaloneDiscoveryType=${DISCOVERY_TYPE}'" ./cmd/tanzu


.PHONY: release
release: ${RELEASE_JOBS} ## Create release binaries

.PHONY: release-%
release-%: ## Create release for a platform
	$(eval ARCH = $(word 2,$(subst -, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$*)))

	$(MAKE) build-tanzu-core-cli-oci-$(OS)-$(ARCH)

## --------------------------------------
## OS Packages
## --------------------------------------
.PHONY: apt-package
apt-package: ## Build a debian package to use with APT
	@if [ "$$(command -v docker)" == "" ]; then \
		echo "Docker required to build apt package" ;\
		exit 1 ;\
	fi

	@# To call this target, the VERSION variable must be set by the caller.  The version must match an existing release
	@# of the tanzu CLI on Github. E.g., VERSION=v0.26.0 make apt-package
	docker run --rm -e VERSION=$${VERSION} -v $(ROOT_DIR):$(ROOT_DIR) ubuntu $(MODULE_ROOT_DIR)/hack/apt/build_package.sh

.PHONY: rpm-package
rpm-package: ## Build an RPM package
	@if [ "$$(command -v docker)" == "" ]; then \
		echo "Docker required to build rpm package" ;\
		exit 1 ;\
	fi

	@# To call this target, the VERSION variable must be set by the caller.  The version must match an existing release
	@# of the tanzu CLI on Github. E.g., VERSION=v0.26.0 make rpm-package
	docker run --rm -e VERSION=$${VERSION} -v $(ROOT_DIR):$(ROOT_DIR) fedora $(MODULE_ROOT_DIR)/hack/rpm/build_package.sh

.PHONY: choco-package
choco-package: ## Build a Chocolatey package
	@if [ "$$(command -v docker)" = "" ]; then \
		echo "Docker required to build chocolatey package" ;\
		exit 1 ;\
	fi

	@# There are only AMD64 images to run chocolatey on docker
	@if [ "$(GOHOSTARCH)" != "amd64" ]; then \
		echo "Can only build chocolatey package on an amd64 machine at the moment" ;\
		exit 1 ;\
	fi

	@# To call this target, the VERSION variable must be set by the caller.  The version must match an existing release
	@# of the tanzu CLI on Github. E.g., VERSION=v0.26.0 make choco-package
	docker run --rm -e VERSION=$${VERSION} -v $(ROOT_DIR):$(ROOT_DIR) chocolatey/choco $(MODULE_ROOT_DIR)/hack/choco/build_package.sh

## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: fmt vet ## Run Tests
	${GO} test ./... -timeout 60m -race -coverprofile coverage.txt -v

.PHONY: vet
vet: ## Vet codebase
	${GO} vet ./...

.PHONY: fmt
fmt: $(GOIMPORTS) ## Run goimports
	$(GOIMPORTS) -w -local github.com/vmware-tanzu ./

.PHONY: generate-fakes
generate-fakes: ## Generate fakes for writing unit tests
	$(GO) generate ./...
	$(MAKE) fmt
## --------------------------------------
## Tooling Binaries
## --------------------------------------

tools: $(TOOLING_BINARIES) ## Build tooling binaries
.PHONY: $(TOOLING_BINARIES)
$(TOOLING_BINARIES):
	make -C $(TOOLS_DIR) $(@F)
