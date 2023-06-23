# Copyright 2021 VMware, Inc. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0

create-kind-cluster:
	kind create cluster --name ${KIND_CLUSTER_NAME} --image "kindest/node:${KUBE_VERSION}"

build-local-readiness: manifests package-vendir-sync
	COMPONENTS=readiness/controller.readiness-controller-manager.readiness make docker-build-all
	docker tag projects.registry.vmware.com/tanzu_framework/readiness-controller-manager:latest readiness-controller-manager:latest 

deploy-local-readiness: build-local-readiness	
	kind load docker-image readiness-controller-manager:latest --name ${KIND_CLUSTER_NAME}
	ytt -f packages/readiness/bundle/config | kubectl apply -f-

e2e-readiness:
	cd readiness/e2e && go test -timeout 10m -v github.com/vmware-tanzu/tanzu-framework/readiness/e2e



