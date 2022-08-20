# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../common.mk
# set root for manifests
RELATIVE_ROOT=.

.PHONY: manifests
manifests: ## Generate manifests e.g. CRD, RBAC etc.
	$(MAKE) generate-manifests CONTROLLER_GEN_SRC=./...