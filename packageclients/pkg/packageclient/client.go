// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient

import (
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/kappclient"
)

type pkgClient struct {
	kappClient kappclient.Client
}

// NewPackageClient instantiates pkgClient  pkgClient with kubeconfig
func NewPackageClient(kubeconfigPath string) (PackageClient, error) {
	return NewPackageClientForContext(kubeconfigPath, "")
}

// NewPackageClientForContext instantiates pkgClient with kubeconfig and kubecontext
func NewPackageClientForContext(kubeconfigPath, kubeContext string) (PackageClient, error) {
	var err error
	client := &pkgClient{}

	if client.kappClient, err = kappclient.NewKappClientForContext(kubeconfigPath, kubeContext); err != nil {
		return nil, err
	}

	return client, nil
}

func NewPackageClientWithKappClient(kappClient kappclient.Client) (PackageClient, error) {
	return &pkgClient{
		kappClient: kappClient,
	}, nil
}
