#!/bin/bash

# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

if [ -n "$1" ]; then
  export KUBECONFIG=$1;
fi

# edit kapp-controller deployment
kubectl patch --type=json deployment kapp-controller -n tkg-system -p '''[{"op":"replace","path":"/spec/template/spec/containers/0/args/0","value": "-packaging-global-namespace=tkg-system"}]'''
sleep 30

# install aws cluster-class package
tanzu package install aws-cc-pi --package-name tkg-clusterclass-aws.tanzu.vmware.com --version 0.18.0-dev --values-file pkg/v1/tkg/test/config/cc/aws/aws_cc_package_config.yaml

# install addons-manager package
export ADDONS_MANAGER_SECRET_NAME=$(kubectl get secret -l tkg.tanzu.vmware.com/addon-name=tanzu-addons-manager -n tkg-system -o=jsonpath='{.items[0].metadata.name}')
kubectl patch secret $ADDONS_MANAGER_SECRET_NAME -n tkg-system -p '{"metadata":{"annotations":{"tkg.tanzu.vmware.com/addon-paused": ""}}}' --type=merge
kubectl delete pkgi tanzu-addons-manager -n tkg-system
kubectl apply -f https://gist.githubusercontent.com/vijaykatam/ca28f5233acf1e4fa48a330e7c03f3d9/raw/7b3a57404b0bfc0963fc7512f82da4e98b9183d2/addons-manager-package.yaml
sleep 30
kubectl apply -f https://gist.githubusercontent.com/vijaykatam/ca28f5233acf1e4fa48a330e7c03f3d9/raw/7b3a57404b0bfc0963fc7512f82da4e98b9183d2/addons-manager-package.yaml

# install TKR related components
kubectl scale deployment/tkr-controller-manager -n tkr-system --replicas=0
kubectl delete tkr --all
kubectl apply -f https://gist.githubusercontent.com/prkalle/61e58f6bb11e8f5599ce8370648394db/raw/19d7f7699952bec76eeb5f753e5737832c9b5e38/tkr-manifests.yaml
sleep 30
kubectl apply -f https://gist.githubusercontent.com/prkalle/61e58f6bb11e8f5599ce8370648394db/raw/19d7f7699952bec76eeb5f753e5737832c9b5e38/tkr-manifests.yaml
kubectl apply -f https://gist.githubusercontent.com/prkalle/4a7d6a19073847816b77b2a4842d4df7/raw/1f3f3f1a0caf3ee59fdeea263835aa1c2ca574d5/tkr-os-images.yaml

tanzu config set features.cluster.auto-apply-generated-clusterclass-based-configuration true
tanzu config set features.global.package-based-lcm true