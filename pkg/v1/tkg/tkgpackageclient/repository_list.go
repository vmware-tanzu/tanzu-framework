// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"github.com/pkg/errors"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/installpackage/v1alpha1"
)

func (p *pkgClient) ListRepositories() (*kappipkg.PackageRepositoryList, error) {
	packageRepositoryList, err := p.kappClient.ListPackageRepositories()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list existing package repositories in the cluster")
	}

	return packageRepositoryList, nil
}
