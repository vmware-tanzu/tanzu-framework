// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"github.com/vmware-tanzu/tanzu-framework/tkg/kappclient"
)

type pkgClient struct {
	kappClient kappclient.Client
}

// NewTKGPackageClient instantiates pkgClient  pkgClient with kubeconfig
func NewTKGPackageClient(kubeconfigPath string) (TKGPackageClient, error) {
	return NewTKGPackageClientForContext(kubeconfigPath, "")
}

// NewTKGPackageClientForContext instantiates pkgClient with kubeconfig and kubecontext
func NewTKGPackageClientForContext(kubeconfigPath, kubeContext string) (TKGPackageClient, error) {
	var err error
	client := &pkgClient{}

	if client.kappClient, err = kappclient.NewKappClientForContext(kubeconfigPath, kubeContext); err != nil {
		return nil, err
	}

	return client, nil
}

func NewTKGPackageClientWithKappClient(kappClient kappclient.Client) (TKGPackageClient, error) {
	return &pkgClient{
		kappClient: kappClient,
	}, nil
}
