// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package kappclient provides CRUD functionality for kapp-controller related resources
package kappclient

import (
	corev1 "k8s.io/api/core/v1"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	secretgen "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

//go:generate counterfeiter -o ../fakes -generate

// Client is the kapp client interface
//
//counterfeiter:generate -o ../fakes/kappclient.go --fake-name KappClient . Client
type Client interface {
	CreatePackageInstall(packageInstall *kappipkg.PackageInstall, pkgPluginResourceCreationStatus *packagedatamodel.PkgPluginResourceCreationStatus) error
	CreatePackageRepository(repository *kappipkg.PackageRepository) error
	DeletePackageRepository(repository *kappipkg.PackageRepository) error
	GetAppCR(appName string, namespace string) (*kappctrl.App, error)
	GetClient() crtclient.Client
	GetPackageInstall(packageInstallName string, namespace string) (*kappipkg.PackageInstall, error)
	GetPackageMetadataByName(packageName string, namespace string) (*kapppkg.PackageMetadata, error)
	GetPackageRepository(repositoryName, namespace string) (*kappipkg.PackageRepository, error)
	GetPackage(packageName string, namespace string) (*kapppkg.Package, error)
	GetSecretExport(secretName, namespace string) (*secretgen.SecretExport, error)
	GetSecretValue(secretName, namespace string) ([]byte, error)
	ListPackageInstalls(namespace string) (*kappipkg.PackageInstallList, error)
	ListPackageMetadata(namespace string) (*kapppkg.PackageMetadataList, error)
	ListPackages(packageName string, namespace string) (*kapppkg.PackageList, error)
	ListPackageRepositories(namespace string) (*kappipkg.PackageRepositoryList, error)
	ListRegistrySecrets(namespace string) (*corev1.SecretList, error)
	ListSecretExports(namespace string) (*secretgen.SecretExportList, error)
	UpdatePackageInstall(packageInstall *kappipkg.PackageInstall, pkgPluginResourceCreationStatus *packagedatamodel.PkgPluginResourceCreationStatus) error
	UpdatePackageRepository(repository *kappipkg.PackageRepository) error
}
