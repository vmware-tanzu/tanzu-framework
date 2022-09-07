# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../common.mk

test: fmt vet ## Run tests
	${GO} test ./... -coverprofile cover.out

vet: ## Run go vet against code
	${GO} vet ./...

COUNTERFEITER := $(TOOLS_BIN_DIR)/counterfeiter
$(COUNTERFEITER): $(TOOLS_BIN_DIR)
	$(GO) build -tags=tools -o $@ github.com/maxbrunsfeld/counterfeiter/v6

generate-fakes: $(COUNTERFEITER) ## Generate fakes for unit tests
	PATH=$(TOOLS_BIN_DIR):"$(PATH)" $(GO) generate ./...
	$(MAKE) fmt

