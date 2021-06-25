// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgpackageclient provides functionality for package plugin
package tkgpackageclient

import (
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

// TKGPackageClient is the TKG package client interface
type TKGPackageClient interface {
	AddRepository(o *tkgpackagedatamodel.RepositoryOptions) error
	DeleteRepository(o *tkgpackagedatamodel.RepositoryOptions) (bool, error)
	GetPackageInstall(o *tkgpackagedatamodel.PackageGetOptions) (*kappipkg.PackageInstall, error)
	GetPackage(pkgName, pkgVersion, namespace string) (*kapppkg.PackageMetadata, *kapppkg.Package, error)
	GetRepository(o *tkgpackagedatamodel.RepositoryOptions) (*kappipkg.PackageRepository, error)
	InstallPackage(o *tkgpackagedatamodel.PackageInstalledOptions) error
	InstallPackageWithProgress(o *tkgpackagedatamodel.PackageInstalledOptions, packageProgress *tkgpackagedatamodel.PackageProgress)
	ListPackageInstalls(o *tkgpackagedatamodel.PackageListOptions) (*kappipkg.PackageInstallList, error)
	ListPackageMetadata(o *tkgpackagedatamodel.PackageListOptions) (*kapppkg.PackageMetadataList, error)
	ListPackages(o *tkgpackagedatamodel.PackageListOptions) (*kapppkg.PackageList, error)
	ListRepositories(o *tkgpackagedatamodel.RepositoryOptions) (*kappipkg.PackageRepositoryList, error)
	UninstallPackage(o *tkgpackagedatamodel.PackageUninstallOptions) (bool, error)
	UpdatePackageInstall(o *tkgpackagedatamodel.PackageInstalledOptions) error
	UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) error
}
