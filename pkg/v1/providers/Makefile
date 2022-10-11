# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL:=help

# Set language for consistent sorting across operating systems
build-provider-template-image:
	make -C "../../../providers" build-provider-template-image

save-provider-template-image: ## Save provider template image
	make -C "../../../providers" save-provider-template-image
	ln -fs ../../../providers/artifacts artifacts

push-provider-template-image: ## Push provider template image
	make -C "../../../providers" push-provider-template-image
