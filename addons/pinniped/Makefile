.DEFAULT_GOAL := help

ROOT_DIR	:= $(dir $(shell pwd))
IMG_CATEGORY := authentication
IMG_CLUSTER_TYPE := management workload
ADDON_NAME := pinniped-addon

IMG_NAME ?= pinniped-addon-templates
IMG_TAG ?= latest

## --------------------------------------
## Help
## --------------------------------------

help: ## Display this help
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

## --------------------------------------
## Build templates
## --------------------------------------

# TODO: Using the following copy and remove is just a hack to make the build work temporarily.
# We need a generic solution to handle this.
.PHONY: template-image
template-image: ## Build template image
	cp -R $(ROOT_DIR)/ytt-common-libs ./
	docker build -t $(IMG_NAME):$(IMG_TAG) -f Dockerfile.templates .
	rm -rf ./ytt-common-libs
