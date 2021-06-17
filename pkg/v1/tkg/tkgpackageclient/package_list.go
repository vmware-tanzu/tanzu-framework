// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"github.com/pkg/errors"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/installpackage/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/packages/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) ListInstalledPackages(o *tkgpackagedatamodel.PackageListOptions) (*kappipkg.InstalledPackageList, error) {
	installedPackageList, err := p.kappClient.ListInstalledPackages(o.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list installed packages")
	}
	return installedPackageList, nil
}

func (p *pkgClient) ListPackages() (*kapppkg.PackageList, error) {
	packageList, err := p.kappClient.ListPackages()
	if err != nil {
		return nil, errors.Wrap(err, "failed to list packages")
	}
	return packageList, nil
}

func (p *pkgClient) ListPackageVersions(o *tkgpackagedatamodel.PackageListOptions) (*kapppkg.PackageVersionList, error) {
	packageVersionList, err := p.kappClient.ListPackageVersions(o.PackageName, o.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list package versions")
	}
	return packageVersionList, nil
}
