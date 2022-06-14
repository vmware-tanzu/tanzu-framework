#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

if [ -n "$1" ]; then
  export KUBECONFIG=$1;
fi

# edit kapp-controller deployment
kubectl patch --type=json deployment kapp-controller -n tkg-system -p '''[{"op":"replace","path":"/spec/template/spec/containers/0/args/0","value": "-packaging-global-namespace=tkg-system"}]'''
sleep 90

# install aws cluster-class package
tanzu package available list
tanzu package install aws-cc-pi --package-name tkg-clusterclass-aws.tanzu.vmware.com --version 0.21.0 --values-file ../../config/cc/aws/aws_cc_package_config.yaml

# install addons-manager package
export ADDONS_MANAGER_SECRET_NAME=$(kubectl get secret -l tkg.tanzu.vmware.com/addon-name=tanzu-addons-manager -n tkg-system -o=jsonpath='{.items[0].metadata.name}')
kubectl patch secret $ADDONS_MANAGER_SECRET_NAME -n tkg-system -p '{"metadata":{"annotations":{"tkg.tanzu.vmware.com/addon-paused": ""}}}' --type=merge
kubectl delete pkgi tanzu-addons-manager -n tkg-system
kubectl apply -f https://gist.githubusercontent.com/saimanoj01/24c3731817e46b69976025176fced529/raw/488fbdb607715dc11864a653e2295b573030fa56/addons-manager-package.yaml
sleep 30
kubectl apply -f https://gist.githubusercontent.com/saimanoj01/24c3731817e46b69976025176fced529/raw/488fbdb607715dc11864a653e2295b573030fa56/addons-manager-package.yaml

# install TKR related components
kubectl scale deployment/tkr-controller-manager -n tkr-system --replicas=0
kubectl delete tkr --all
kubectl apply -f https://gist.githubusercontent.com/prkalle/61e58f6bb11e8f5599ce8370648394db/raw/19d7f7699952bec76eeb5f753e5737832c9b5e38/tkr-manifests.yaml
sleep 30
kubectl apply -f https://gist.githubusercontent.com/prkalle/61e58f6bb11e8f5599ce8370648394db/raw/19d7f7699952bec76eeb5f753e5737832c9b5e38/tkr-manifests.yaml
kubectl apply -f https://gist.githubusercontent.com/prkalle/4a7d6a19073847816b77b2a4842d4df7/raw/70dbda3db17e0577a4f31895fce16296ecc6ee54/tkr-os-images.yaml

tanzu config set features.cluster.auto-apply-generated-clusterclass-based-configuration true
tanzu config set features.global.package-based-lcm true