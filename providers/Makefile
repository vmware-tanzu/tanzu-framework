# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL:=help

# Set language for consistent sorting across operating systems
export LC_ALL := C

ROOT_DIR := $(shell git rev-parse --show-toplevel)

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
PROJECT_ROOT = $(shell pwd)

CLI_REPO ?= ${ROOT_DIR}
TOOLS_DIR ?= ${ROOT_DIR}/hack/tools
TOOLS_BIN_DIR ?= ${TOOLS_DIR}/bin
GOBINDATA := $(TOOLS_BIN_DIR)/gobindata
VENDIR := $(TOOLS_BIN_DIR)/vendir
VENDIR_VERSION = v0.17.0
CLUSTERGEN_OUTPUT_DIR ?= generated
CLUSTERGEN_BASE ?= origin/main
PROVIDER_DIR_NAME = ${ROOT_DIR}/providers
PROVIDER_BUNDLE_DIR = ${PROVIDER_DIR_NAME}/provider-bundle/providers
PROVIDER_BUNDLE_ZIP = ${PROVIDER_DIR_NAME}/client/manifest/providers.zip

# Image URL to use building/pushing provider image
PROVIDER_TEMPLATE_IMG_NAME ?= provider-templates
IMAGE_REPO ?= vmware.io
PROVIDER_TEMPLATE_IMG_TAG ?= latest


FILES_TO_IGNORE?=.env .idea .git .github hack .yamllint .gitlab Makefile provider-bundle tests README.md client go.mod go.sum providers.sha256sum Dockerfile.templates
ifeq ($(EXCLUDE_CLUSTER_CLASS_FILES), true)
FILES_TO_IGNORE+=yttcc yttcb cconly cc.yaml
endif
FILES_TO_IGNORE_REGEX=$(subst $(space) ,\|,$(FILES_TO_IGNORE))


help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

verify-build:  ## verify-build verifies the changes with generated files

all: lint ## run all target

.PHONY: build-cli ## build tkg cli with existing provider changes
build-cli:
	CLI_REPO=${CLI_REPO} ${CLI_REPO}/hack/clustergen/rebuild-cli.sh ${PWD}

## --------------------------------------
## Linting
## --------------------------------------

.PHONY: lint
lint: ## Run YAML linter
	yamllint ${PROVIDER_DIR_NAME} -c ${PROVIDER_DIR_NAME}/.yamllint

## --------------------------------------
## Generate provider templates
## --------------------------------------

.PHONY: clean-providers
clean-providers: ## cleans provider-bundle
	# cleanup old provider bundles
	@rm -rf client/manifest/providers.zip
	@rm -rf provider-bundle || true
	@mkdir -p ${PROVIDER_BUNDLE_DIR}

setup-provider-for-generation: clean-providers
	# using temporary directory to extract all files which needs to bundle
	find . -type f | grep -v "${FILES_TO_IGNORE_REGEX}" | xargs tar cf - -T - | tar -C ${PROVIDER_BUNDLE_DIR} -x

.PHONY: generate-provider-bundle-zip
generate-provider-bundle-zip: setup-provider-for-generation ## generates provider zip bundle
	cd ${PROVIDER_BUNDLE_DIR} && zip -r ${PROVIDER_BUNDLE_ZIP} ${PROVIDER_DIR_NAME} 

## --------------------------------------
## build save and push Provider templates
## --------------------------------------

build-provider-template-image: setup-provider-for-generation ## Build provider template image
	cd ${PROVIDER_BUNDLE_DIR} && docker build -t $(IMAGE_REPO)/${PROVIDER_TEMPLATE_IMG_NAME}:${PROVIDER_TEMPLATE_IMG_TAG} -f ../../hack/Dockerfile.templates ${PROVIDER_DIR_NAME} 

save-provider-template-image: ## Save provider template image
	mkdir -p "${PROVIDER_DIR_NAME}/artifacts"
	docker save $(IMAGE_REPO)/${PROVIDER_TEMPLATE_IMG_NAME}:${PROVIDER_TEMPLATE_IMG_TAG} | gzip -c > "${PROVIDER_DIR_NAME}/artifacts/${PROVIDER_TEMPLATE_IMG_NAME}-${PROVIDER_TEMPLATE_IMG_TAG}.tar.gz"

push-provider-template-image: ## Push provider template image
	docker push $(IMAGE_REPO)/${PROVIDER_TEMPLATE_IMG_NAME}:${PROVIDER_TEMPLATE_IMG_TAG}

## --------------------------------------
## Cluster generation tests
## --------------------------------------

rename-test-outputs: ## show commands to rename the expected output files to match up with testcases renamed by generate-testcases
	# Assumes that the topmost commit contains the testcases rename
	git show --summary | grep '^ rename' | perl -pe 's@rename (.*){(\d+.case) => (\d+.case).*@git mv $$1/expected/$$2.output $$1/expected/$$3.output@'

.PHONY: generate-testcases
generate-testcases: ## Regenerates test cases based on parameter model files. Requires PICT to be available.
	cd ${PROVIDER_DIR_NAME}/tests/clustergen && ./gen_testcases.sh

.PHONY: cluster-generation-tests
cluster-generation-tests: build-cli ## Run cluster compliance test against cluster generation dataset
	cd tests/clustergen && TKG=${CLI_REPO}/bin/tkg-${GOOS}-${GOARCH} ./run_tests.sh ${CLUSTERGEN_OUTPUT_DIR} ${CLUSTERGEN_CC_OUTPUT_DIR}

.PHONY: cluster-generation-diffs
cluster-generation-diffs: ## Run before/after cluster generation comparison tests
	CLI_REPO=${CLI_REPO} ${CLI_REPO}/hack/clustergen/cluster-check.sh ${CLUSTERGEN_BASE}

## --------------------------------------
## Install tools
## --------------------------------------

vendir: $(VENDIR) ## Install vendir
$(VENDIR):
	mkdir -p hack/tools/bin
	curl -LO https://github.com/vmware-tanzu/carvel-vendir/releases/download/$(VENDIR_VERSION)/vendir-$(GOOS)-$(GOARCH)
	mv vendir-$(GOOS)-$(GOARCH) $(@)
	chmod a+x $(@)

## --------------------------------------
## Run vendir sync
## --------------------------------------

.PHONY: vendir-sync
vendir-sync: vendir  ## Run vendir sync
	${TOOLS_BIN_DIR}/vendir sync --file ${PROVIDER_DIR_NAME}/vendir.yml
# end of Makefile
