#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

set -x

wget https://github.com/mikefarah/yq/releases/download/v4.24.5/yq_linux_amd64
mv yq_linux_amd64 /usr/local/bin/yq
chmod +x /usr/local/bin/yq

yq -i e '.components.tanzu-framework-management-packages[0].version |= "v0.18.0-dev-13-g233a6405"' $HOME/.config/tanzu/tkg/bom/tkg-bom-v1.6.0-zshippable.yaml
yq -i e '.components.tanzu-framework-management-packages[0].images.tanzuFrameworkManagementPackageRepositoryImage.imagePath |= "management"' $HOME/.config/tanzu/tkg/bom/tkg-bom-v1.6.0-zshippable.yaml
yq -i e '.components.tanzu-framework-management-packages[0].images.tanzuFrameworkManagementPackageRepositoryImage.imageRepository |= "gcr.io/eminent-nation-87317/tkg/test16"' $HOME/.config/tanzu/tkg/bom/tkg-bom-v1.6.0-zshippable.yaml
yq -i e '.components.tanzu-framework-management-packages[0].images.tanzuFrameworkManagementPackageRepositoryImage.tag |= "v0.21.0"' $HOME/.config/tanzu/tkg/bom/tkg-bom-v1.6.0-zshippable.yaml
