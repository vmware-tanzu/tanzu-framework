# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../common.mk

ROOT_DIR := $(shell git rev-parse --show-toplevel)
ARTIFACTS_DIR ?= $(ROOT_DIR)/artifacts
CLI_PLUGINS_DIR ?= $(ROOT_DIR)/cliplugins

IMG_BASEPATH := tanzu-cli-plugins
IMG_DEFAULT_TAG := latest

IMG_VERSION_OVERRIDE ?= $(IMG_DEFAULT_TAG)

ifeq ($(strip $(OCI_REGISTRY)),)
	IMG_REPOSITORY ?= $(IMG_BASEPATH)
else
	IMG_REPOSITORY ?= $(OCI_REGISTRY)/$(IMG_BASEPATH)
endif

.PHONY: docker-build-%
docker-build-%:
	$(eval PLUGIN_NAME = $(word 1,$(subst _, ,$*)))
	$(eval OSARCH = $(word 2,$(subst _, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$(OSARCH))))
	$(eval ARCH = $(word 2,$(subst -, ,$(OSARCH))))
	
	cd $(ARTIFACTS_DIR)/$(OS)/$(ARCH)/cli/$(PLUGIN_NAME)/$(BUILD_VERSION) && docker build -t $(IMG_REPOSITORY)/$(PLUGIN_NAME)-$(OS)-$(ARCH):$(IMG_VERSION_OVERRIDE) -f $(CLI_PLUGINS_DIR)/Dockerfile.templates .

.PHONY: docker-publish-%
docker-publish-%:
	$(eval PLUGIN_NAME = $(word 1,$(subst _, ,$*)))
	$(eval OSARCH = $(word 2,$(subst _, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$(OSARCH))))
	$(eval ARCH = $(word 2,$(subst -, ,$(OSARCH))))

	docker push $(IMG_REPOSITORY)/$(PLUGIN_NAME)-$(OS)-$(ARCH):$(IMG_VERSION_OVERRIDE)

.PHONY: kbld-image-replace-%
kbld-image-replace-%: ## Add newImage in kbld-config.yaml
	$(eval PLUGIN_NAME = $(word 1,$(subst _, ,$*)))
	$(eval OSARCH = $(word 2,$(subst _, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$(OSARCH))))
	$(eval ARCH = $(word 2,$(subst -, ,$(OSARCH))))
	$(eval PACKAGE_NAME = $(word 3,$(subst _, ,$*)))

	cd ../hack/packages/kbld-image-replace && \
	  go run main.go -kbld-config ../../../packages/$(PACKAGE_NAME)/kbld-config.yaml \
	  $(IMG_BASEPATH)/$(PLUGIN_NAME)-$(OS)-$(ARCH):$(IMG_DEFAULT_TAG) \
	  $(IMG_REPOSITORY)/$(PLUGIN_NAME)-$(OS)-$(ARCH):$(IMG_VERSION_OVERRIDE)

.PHONY: docker-image-name-%
docker-image-name-%:
	$(eval PLUGIN_NAME = $(word 1,$(subst _, ,$*)))
	$(eval OSARCH = $(word 2,$(subst _, ,$*)))
	$(eval OS = $(word 1,$(subst -, ,$(OSARCH))))
	$(eval ARCH = $(word 2,$(subst -, ,$(OSARCH))))

	@echo $(IMG_REPOSITORY)/$(PLUGIN_NAME)-$(OS)-$(ARCH):$(IMG_VERSION_OVERRIDE)

.PHONY: docker-build
docker-build: ## Build docker image
	@ for plugin in $(PLUGINS_WITHOUT_TARGET) ; do \
		for env in $(ENVS) ; do \
			$(MAKE) docker-build-$$plugin"_"$$env --no-print-directory ; \
		done \
	done

.PHONY: docker-publish
docker-publish: ## Publish docker image
	@ for plugin in $(PLUGINS_WITHOUT_TARGET) ; do \
		for env in $(ENVS) ; do \
			$(MAKE) docker-publish-$$plugin"_"$$env --no-print-directory ; \
		done \
	done

.PHONY: kbld-image-replace
kbld-image-replace: ## Add newImage in kbld-config.yaml
	@ for plugin in $(STANDALONE_PLUGINS_WITHOUT_TARGET) ; do \
		for env in $(ENVS) ; do \
			$(MAKE) kbld-image-replace-$$plugin"_"$$env"_"standalone-plugins --no-print-directory ; \
		done \
	done
	@ for plugin in $(CONTEXTAWARE_PLUGINS_WITHOUT_TARGET) ; do \
		for env in $(ENVS) ; do \
			$(MAKE) kbld-image-replace-$$plugin"_"$$env"_"core-management-plugins --no-print-directory ; \
		done \
	done

.PHONY: docker-image-names
docker-image-names: ## Return newline-separated names of all docker containers built in this package
	@ for plugin in $(PLUGINS_WITHOUT_TARGET) ; do \
		for env in $(ENVS) ; do \
			$(MAKE) docker-image-name-$$plugin"_"$$env --no-print-directory ; \
		done \
	done

.PHONY: docker-build-and-publish
docker-build-and-publish: docker-build docker-publish kbld-image-replace
