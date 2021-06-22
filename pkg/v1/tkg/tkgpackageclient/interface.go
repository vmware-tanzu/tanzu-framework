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
	GetPackage(pkgName, pkgVersion, namespace string) (*kapppkg.PackageMetadata, *kapppkg.Package, error)
	GetRepository(o *tkgpackagedatamodel.RepositoryGetOptions) (*kappipkg.PackageRepository, error)
	GetPackageInstall(o *tkgpackagedatamodel.PackageGetOptions) (*kappipkg.PackageInstall, error)
	InstallPackage(o *tkgpackagedatamodel.PackageOptions) error
	UninstallPackage(o *tkgpackagedatamodel.PackageUninstallOptions) (bool, error)
	UpdatePackage(o *tkgpackagedatamodel.PackageOptions) error
	AddRepository(o *tkgpackagedatamodel.RepositoryOptions) error
	DeleteRepository(o *tkgpackagedatamodel.RepositoryDeleteOptions) (bool, error)
	ListPackageInstalls(o *tkgpackagedatamodel.PackageListOptions) (*kappipkg.PackageInstallList, error)
	ListPackages(o *tkgpackagedatamodel.PackageListOptions) (*kapppkg.PackageList, error)
	ListPackageMetadata(o *tkgpackagedatamodel.PackageListOptions) (*kapppkg.PackageMetadataList, error)
	ListRepositories(o *tkgpackagedatamodel.RepositoryListOptions) (*kappipkg.PackageRepositoryList, error)
	UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) error
}
