# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL:=help

# Set language for consistent sorting across operating systems
export LC_ALL := C

GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
PROJECT_ROOT = $(shell pwd)

CLI_REPO ?= ${PROJECT_ROOT}/../../..
TOOLS_DIR ?= ${PROJECT_ROOT}/../../../hack/tools
TOOLS_BIN_DIR ?= ${TOOLS_DIR}/bin
GOBINDATA := $(TOOLS_BIN_DIR)/gobindata
VENDIR := $(TOOLS_BIN_DIR)/vendir
VENDIR_VERSION = v0.17.0
CLUSTERGEN_OUTPUT_DIR ?= expected
CLUSTERGEN_BASE ?= origin/main

FILES_TO_IGNORE='.env\|.idea\|.git\|.github\|hack\|.yamllint\|.gitlab\|Makefile\|provider-bundle\|tests\|README.md\|client\|go.mod\|go.sum'

help:  ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

verify-build:  ## verify-build verifies the changes with generated files

all: lint checksum generate-bindata verify ## run all target

ci: checksum generate-bindata verify ## run ci target

## --------------------------------------
## Linting
## --------------------------------------

.PHONY: lint
lint: ## Run YAML linter
	${CLI_REPO}/hack/yamllint.sh

## --------------------------------------
## Generate provider
## --------------------------------------

.PHONY: clean-providers
clean-providers: ## cleans provider-bundle
	# cleanup old provider bundles
	@rm -rf provider-bundle || true
	@mkdir -p provider-bundle/tmp

.PHONY: generate-provider-bundle
generate-provider-bundle: clean-providers ## generates provider zip bundle
	# using temporary directory to extract all files which needs to bundle,
	# updating the file modification time, file permissions and sorting files to get
	# consistent zip file across all platform as this zip bytes are used to generate
	# zz_generated.bindata.go file
	find . -type f | grep -v ${FILES_TO_IGNORE} | xargs tar cf - -T - | tar -C provider-bundle/tmp -xv
	cd provider-bundle/tmp && for i in `find . -type f`; do touch -t 202006091513 $$i; chmod 644 $$i; done && find . -type f | sort | zip -qXo ../providers.zip -@ && cd .. && rm -rf tmp

rename-test-outputs: ## show commands to rename the expected output files to match up with testcases renamed by generate-testcases
	# Assumes that the topmost commit contains the testcases rename
	git show --summary | grep '^ rename' | perl -pe 's@rename (.*){(\d+.case) => (\d+.case).*@git mv $$1/expected/$$2.output $$1/expected/$$3.output@'

.PHONY: generate-bindata
generate-bindata: generate-provider-bundle ## Generate go-bindata files
	$(GOBINDATA) -mode=420 -modtime=1 -nometadata -o=client/manifest/zz_generated.bindata.go -pkg=manifest provider-bundle/providers.zip

.PHONY: checksum
checksum: ## Generate checksum of provider files
	find . -type f | grep -v ${FILES_TO_IGNORE} | sort | xargs shasum -a 256 | shasum -a 256 | cut -d" " -f1 > providers.sha256sum

.PHONY: verify
verify: ## verify that checksum and generated bindata file match
	git diff --exit-code providers.sha256sum client/manifest/zz_generated.bindata.go || (echo "Run make vendir-sync and ensure diff is added to your changeset")

.PHONY: update-checksum-bindata ## update providers checksum and bindata file
update-checksum-bindata: checksum generate-bindata

.PHONY: build-cli ## build tkg cli with existing provider changes
build-cli:
	CLI_REPO=${CLI_REPO} ${CLI_REPO}/hack/clustergen/rebuild-cli.sh ${PWD}

## --------------------------------------
## Cluster generation tests
## --------------------------------------

.PHONY: generate-testcases
generate-testcases: ## Regenerates test cases based on parameter model files. Requires PICT to be available.
	cd tests/clustergen && ./gen_testcases.sh

.PHONY: cluster-generation-tests
cluster-generation-tests: build-cli ## Run cluster compliance test against cluster generation dataset
	cd tests/clustergen && TKG=${CLI_REPO}/bin/tkg-${GOOS}-${GOARCH} ./run_tests.sh ${CLUSTERGEN_OUTPUT_DIR}

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
vendir-sync: ## Run "vendir sync"
	rm -rf ./ytt/vendir/cni/_ytt_lib/addons/calico/templates/* && \
	rm -rf ./ytt/vendir/cni/_ytt_lib/addons/antrea/templates/* && \
	rm -rf ./ytt/vendir/vsphere_cpi/_ytt_lib/addons/vsphere_cpi/templates/* && \
	cp -r ../../../addons/calico/templates ./ytt/vendir/cni/_ytt_lib/addons/calico/ && \
	cp -r ../../../addons/antrea/templates ./ytt/vendir/cni/_ytt_lib/addons/antrea/ && \
	cp -r ../../../addons/vsphere_cpi/templates ./ytt/vendir/vsphere_cpi/_ytt_lib/addons/vsphere_cpi/

# end of Makefile
