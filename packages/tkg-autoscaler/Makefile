# Copyright 2022 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

include ../../common.mk

configure-package: ## Configure package before creating the package
	sed -e 's/\IMAGE/${AUTOSCALER_IMAGE}/' values.template.yaml > bundle/config/zz_generated_values.yaml

reset-package: ## Reset configured package
	rm bundle/config/zz_generated_values.yaml | true
