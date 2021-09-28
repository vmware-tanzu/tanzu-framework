// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgpackageclient provides functionality for package plugin
package tkgpackageclient

import (
	corev1 "k8s.io/api/core/v1"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	secretgen "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

// TKGPackageClient is the TKG package client interface
type TKGPackageClient interface {
	AddImagePullSecret(o *tkgpackagedatamodel.ImagePullSecretOptions) error
	AddRepository(o *tkgpackagedatamodel.RepositoryOptions) error
	DeleteImagePullSecret(o *tkgpackagedatamodel.ImagePullSecretOptions) (bool, error)
	DeleteRepository(o *tkgpackagedatamodel.RepositoryOptions) (bool, error)
	GetPackageInstall(o *tkgpackagedatamodel.PackageOptions) (*kappipkg.PackageInstall, error)
	GetPackage(o *tkgpackagedatamodel.PackageOptions) (*kapppkg.PackageMetadata, *kapppkg.Package, error)
	GetRepository(o *tkgpackagedatamodel.RepositoryOptions) (*kappipkg.PackageRepository, error)
	InstallPackage(o *tkgpackagedatamodel.PackageOptions, packageProgress *tkgpackagedatamodel.PackageProgress, update bool)
	ListImagePullSecrets(o *tkgpackagedatamodel.ImagePullSecretOptions) (*corev1.SecretList, error)
	ListPackageInstalls(o *tkgpackagedatamodel.PackageOptions) (*kappipkg.PackageInstallList, error)
	ListPackageMetadata(o *tkgpackagedatamodel.PackageAvailableOptions) (*kapppkg.PackageMetadataList, error)
	ListPackages(o *tkgpackagedatamodel.PackageAvailableOptions) (*kapppkg.PackageList, error)
	ListRepositories(o *tkgpackagedatamodel.RepositoryOptions) (*kappipkg.PackageRepositoryList, error)
	ListSecretExports(o *tkgpackagedatamodel.ImagePullSecretOptions) (*secretgen.SecretExportList, error)
	UninstallPackage(o *tkgpackagedatamodel.PackageOptions, packageProgress *tkgpackagedatamodel.PackageProgress)
	UpdateImagePullSecret(o *tkgpackagedatamodel.ImagePullSecretOptions) error
	UpdatePackage(o *tkgpackagedatamodel.PackageOptions, packageProgress *tkgpackagedatamodel.PackageProgress)
	UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) error
}
