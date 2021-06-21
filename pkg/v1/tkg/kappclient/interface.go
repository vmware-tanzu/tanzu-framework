// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kappclient provides CRUD functionality for kapp-controller related resources
package kappclient

import (
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
)

//go:generate counterfeiter -o ../fakes/kappclient.go --fake-name KappClient . Client

// Client is the kapp client interface
type Client interface {
	CreatePackageInstall(installedPackage *kappipkg.PackageInstall, isPkgPluginCreatedSvcAccount bool, isPkgPluginCreatedSecret bool) error
	CreatePackageRepository(repository *kappipkg.PackageRepository) error
	DeletePackageRepository(repository *kappipkg.PackageRepository) error
	GetPackageRepository(repositoryName string, namespace string) (*kappipkg.PackageRepository, error)
	GetAppCR(appName string, namespace string) (*kappctrl.App, error)
	GetClient() crtclient.Client
	GetPackageInstall(installedPackageName string, namespace string) (*kappipkg.PackageInstall, error)
	GetPackageMetadataByName(packageName string, namespace string) (*kapppkg.PackageMetadata, error)
	ListPackageInstalls(namespace string) (*kappipkg.PackageInstallList, error)
	ListPackages(packageName string, namespace string) (*kapppkg.PackageList, error)
	ListPackageMetadata(namespace string) (*kapppkg.PackageMetadataList, error)
	ListPackageRepositories(namespace string) (*kappipkg.PackageRepositoryList, error)
	UpdatePackageRepository(repository *kappipkg.PackageRepository) error
}
