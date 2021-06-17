// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kappclient provides CRUD functionality for kapp-controller related resources
package kappclient

import (
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/installpackage/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/packages/v1alpha1"
)

// Client is the kapp client interface
type Client interface {
	CreateInstalledPackage(installedPackage *kappipkg.InstalledPackage, isPkgPluginCreatedSvcAccount bool, isPkgPluginCreatedSecret bool) error
	CreatePackageRepository(repository *kappipkg.PackageRepository) error
	DeletePackageRepository(repository *kappipkg.PackageRepository) error
	GetPackageRepository(repositoryName string) (*kappipkg.PackageRepository, error)
	GetAppCR(appName string, namespace string) (*kappctrl.App, error)
	GetClient() crtclient.Client
	GetInstalledPackage(installedPackageName string, namespace string) (*kappipkg.InstalledPackage, error)
	GetPackageByName(packageName string, namespace string) (*kapppkg.Package, error)
	ListInstalledPackages(namespace string) (*kappipkg.InstalledPackageList, error)
	ListPackageVersions(packageName string, namespace string) (*kapppkg.PackageVersionList, error)
	ListPackages() (*kapppkg.PackageList, error)
	ListPackageRepositories() (*kappipkg.PackageRepositoryList, error)
	UpdatePackageRepository(repository *kappipkg.PackageRepository) error
}
