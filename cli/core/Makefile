# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../common.mk

# Active module mode, as we use go modules to manage dependencies
export GO111MODULE := on

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
ifneq ($(strip $(TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.DefaultAllowedPluginRepositories=$(TANZU_PLUGINS_ALLOWED_IMAGE_REPOSITORIES)'
endif

ifneq ($(strip $(ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common.IsContextAwareDiscoveryEnabled=$(ENABLE_CONTEXT_AWARE_PLUGIN_DISCOVERY)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.DefaultStandaloneDiscoveryImagePath=$(DEFAULT_STANDALONE_DISCOVERY_IMAGE_PATH)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.DefaultStandaloneDiscoveryImageTag=$(DEFAULT_STANDALONE_DISCOVERY_IMAGE_TAG)'
endif
ifneq ($(strip $(DEFAULT_STANDALONE_DISCOVERY_LOCAL_PATH)),)
LD_FLAGS += -X 'github.com/vmware-tanzu/tanzu-framework/pkg/v1/config.DefaultStandaloneDiscoveryLocalPath=$(DEFAULT_STANDALONE_DISCOVERY_LOCAL_PATH)'
endif


BUILD_TAGS ?=



.PHONY: all
all: build-tanzu-core install-tanzu-core test ## Tests the library

.PHONY: build-tanzu-core
build-tanzu-core: ##Build the tanzu core image
	$(GO) build -tags "${BUILD_TAGS}" --ldflags "$(LD_FLAGS)"  ./cmd/cli/tanzu/main.go -o $(ARTIFACTS_DIR)/$(GOHOSTOS)/$(GOHOSTARCH)/cli/core/$(BUILD_VERSION)/tanz-core-$(GOHOSTOS)_$(GOHOSTARCH)



## --------------------------------------
## Testing
## --------------------------------------

.PHONY: test
test: fmt vet ## Run Tests
	${GO} test ./... -timeout 60m -race -coverprofile coverage.txt -v 2

.PHONY: vet
vet: ## Vet codebase
	go vet ./...

