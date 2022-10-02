// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package packageclient provides functionality for package plugin
package packageclient

import (
	corev1 "k8s.io/api/core/v1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	secretgen "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

//go:generate counterfeiter -o ../fakes -generate

// PackageClient is the TKG package client interface
//
//counterfeiter:generate -o ../fakes/packageclient.go --fake-name PackageClient . PackageClient
type PackageClient interface {
	AddRegistrySecret(o *packagedatamodel.RegistrySecretOptions) error
	AddRepository(o *packagedatamodel.RepositoryOptions, packageProgress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType)
	AddRepositorySync(o *packagedatamodel.RepositoryOptions, operationType packagedatamodel.OperationType) error
	DeleteRegistrySecret(o *packagedatamodel.RegistrySecretOptions) (bool, error)
	DeleteRepository(o *packagedatamodel.RepositoryOptions, packageProgress *packagedatamodel.PackageProgress)
	DeleteRepositorySync(o *packagedatamodel.RepositoryOptions) error
	GetPackageInstall(o *packagedatamodel.PackageOptions) (*kappipkg.PackageInstall, error)
	GetPackage(o *packagedatamodel.PackageOptions) (*kapppkg.PackageMetadata, *kapppkg.Package, error)
	GetRepository(o *packagedatamodel.RepositoryOptions) (*kappipkg.PackageRepository, error)
	GetSecretExport(o *packagedatamodel.RegistrySecretOptions) (*secretgen.SecretExport, error)
	InstallPackage(o *packagedatamodel.PackageOptions, packageProgress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType)
	InstallPackageSync(o *packagedatamodel.PackageOptions, operationType packagedatamodel.OperationType) error
	ListPackageInstalls(o *packagedatamodel.PackageOptions) (*kappipkg.PackageInstallList, error)
	ListPackageMetadata(o *packagedatamodel.PackageAvailableOptions) (*kapppkg.PackageMetadataList, error)
	ListPackages(o *packagedatamodel.PackageAvailableOptions) (*kapppkg.PackageList, error)
	ListRegistrySecrets(o *packagedatamodel.RegistrySecretOptions) (*corev1.SecretList, error)
	ListSecretExports(o *packagedatamodel.RegistrySecretOptions) (*secretgen.SecretExportList, error)
	ListRepositories(o *packagedatamodel.RepositoryOptions) (*kappipkg.PackageRepositoryList, error)
	UninstallPackage(o *packagedatamodel.PackageOptions, packageProgress *packagedatamodel.PackageProgress)
	UninstallPackageSync(o *packagedatamodel.PackageOptions) error
	UpdateRegistrySecret(o *packagedatamodel.RegistrySecretOptions) error
	UpdatePackage(o *packagedatamodel.PackageOptions, packageProgress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType)
	UpdatePackageSync(o *packagedatamodel.PackageOptions, operationType packagedatamodel.OperationType) error
	UpdateRepository(o *packagedatamodel.RepositoryOptions, progress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType)
	UpdateRepositorySync(o *packagedatamodel.RepositoryOptions, operationType packagedatamodel.OperationType) error
}

// CrtClient clientset interface
//
//counterfeiter:generate -o ../fakes/crtclient.go --fake-name CrtClient . CrtClient
type CrtClient interface {
	crtclient.Client
}
