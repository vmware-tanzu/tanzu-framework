// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgpackageclient provides functionality for package plugin
package tkgpackageclient

import (
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/installpackage/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/packages/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

// TKGPackageClient is the TKG package client interface
type TKGPackageClient interface {
	InstallPackage(o *tkgpackagedatamodel.PackageOptions) error
	UnInstallPackage(o *tkgpackagedatamodel.PackageUninstallOptions) error
	AddRepository(o *tkgpackagedatamodel.RepositoryOptions) error
	DeleteRepository(o *tkgpackagedatamodel.RepositoryDeleteOptions) (bool, error)
	ListInstalledPackages(o *tkgpackagedatamodel.PackageListOptions) (*kappipkg.InstalledPackageList, error)
	ListPackageVersions(o *tkgpackagedatamodel.PackageListOptions) (*kapppkg.PackageVersionList, error)
	ListPackages() (*kapppkg.PackageList, error)
	ListRepositories() (*kappipkg.PackageRepositoryList, error)
	UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) error
}
