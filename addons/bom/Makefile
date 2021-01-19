# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := help

ROOT_DIR	:= $(dir $(shell pwd))

OBJECTS ?= addons-manager vsphere_cpi kapp-controller pinniped antrea calico metrics-server vsphere_csi# add more add-ons here, folder name

IMG_TAG ?= latest

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Build templates
## --------------------------------------

.PHONY: generate-addon-bom
generate-addon-bom: init-boms $(OBJECTS) move-bom-to-artifacts## generate add-on BOM

.PHONY: init-boms
init-boms:
	$(shell echo "addons:" > bom.yaml)

$(OBJECTS): 
	./generate_values.sh $@ $(ROOT_DIR) $(IMG_TAG)

.PHONY: move-bom-to-artifacts
move-bom-to-artifacts:
	$(shell mv ./bom.yaml ../artifacts/bom.yaml)