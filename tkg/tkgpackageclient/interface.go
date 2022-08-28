// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgpackageclient provides functionality for package plugin
package tkgpackageclient

import (
	corev1 "k8s.io/api/core/v1"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	secretgen "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

//go:generate counterfeiter -o ../fakes/tkgpackageclient.go --fake-name TKGPackageClient . TKGPackageClient

// TKGPackageClient is the TKG package client interface
type TKGPackageClient interface {
	AddRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) error
	AddRepository(o *tkgpackagedatamodel.RepositoryOptions, packageProgress *tkgpackagedatamodel.PackageProgress, operationType tkgpackagedatamodel.OperationType)
	AddRepositorySync(o *tkgpackagedatamodel.RepositoryOptions, operationType tkgpackagedatamodel.OperationType) error
	DeleteRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) (bool, error)
	DeleteRepository(o *tkgpackagedatamodel.RepositoryOptions, packageProgress *tkgpackagedatamodel.PackageProgress)
	DeleteRepositorySync(o *tkgpackagedatamodel.RepositoryOptions) error
	GetPackageInstall(o *tkgpackagedatamodel.PackageOptions) (*kappipkg.PackageInstall, error)
	GetPackage(o *tkgpackagedatamodel.PackageOptions) (*kapppkg.PackageMetadata, *kapppkg.Package, error)
	GetRepository(o *tkgpackagedatamodel.RepositoryOptions) (*kappipkg.PackageRepository, error)
	GetSecretExport(o *tkgpackagedatamodel.RegistrySecretOptions) (*secretgen.SecretExport, error)
	InstallPackage(o *tkgpackagedatamodel.PackageOptions, packageProgress *tkgpackagedatamodel.PackageProgress, operationType tkgpackagedatamodel.OperationType)
	InstallPackageSync(o *tkgpackagedatamodel.PackageOptions, operationType tkgpackagedatamodel.OperationType) error
	ListPackageInstalls(o *tkgpackagedatamodel.PackageOptions) (*kappipkg.PackageInstallList, error)
	ListPackageMetadata(o *tkgpackagedatamodel.PackageAvailableOptions) (*kapppkg.PackageMetadataList, error)
	ListPackages(o *tkgpackagedatamodel.PackageAvailableOptions) (*kapppkg.PackageList, error)
	ListRegistrySecrets(o *tkgpackagedatamodel.RegistrySecretOptions) (*corev1.SecretList, error)
	ListSecretExports(o *tkgpackagedatamodel.RegistrySecretOptions) (*secretgen.SecretExportList, error)
	ListRepositories(o *tkgpackagedatamodel.RepositoryOptions) (*kappipkg.PackageRepositoryList, error)
	UninstallPackage(o *tkgpackagedatamodel.PackageOptions, packageProgress *tkgpackagedatamodel.PackageProgress)
	UninstallPackageSync(o *tkgpackagedatamodel.PackageOptions) error
	UpdateRegistrySecret(o *tkgpackagedatamodel.RegistrySecretOptions) error
	UpdatePackage(o *tkgpackagedatamodel.PackageOptions, packageProgress *tkgpackagedatamodel.PackageProgress, operationType tkgpackagedatamodel.OperationType)
	UpdatePackageSync(o *tkgpackagedatamodel.PackageOptions, operationType tkgpackagedatamodel.OperationType) error
	UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions, progress *tkgpackagedatamodel.PackageProgress, operationType tkgpackagedatamodel.OperationType)
	UpdateRepositorySync(o *tkgpackagedatamodel.RepositoryOptions, operationType tkgpackagedatamodel.OperationType) error
}
