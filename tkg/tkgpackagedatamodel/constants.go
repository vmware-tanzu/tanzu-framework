// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint
package tkgpackagedatamodel

import "time"

const (
	ClusterRoleBindingName              = "%s-%s-cluster-rolebinding"
	ClusterRoleName                     = "%s-%s-cluster-role"
	DataPackagingAPIName                = "data.packaging.carvel.dev"
	DefaultAPIVersion                   = "install.package.carvel.dev/v1alpha1"
	DefaultPollInterval                 = 1 * time.Second
	DefaultPollTimeout                  = 15 * time.Minute
	DefaultRepositoryImageTag           = "latest"
	DefaultRepositoryImageTagConstraint = ">0.0.0"
	ErrPackageNotInstalled              = "package install does not exist in the namespace"
	ErrRepoNotExists                    = "package repository does not exist in the namespace"
	ErrPackageAlreadyExists             = "package install already exists in the namespace"
	ErrRepoAlreadyExists                = "package repository already exists in the namespace"
	KindClusterRole                     = "ClusterRole"
	KindClusterRoleBinding              = "ClusterRoleBinding"
	KindNamespace                       = "Namespace"
	KindPackageInstall                  = "PackageInstall"
	KindPackageRepository               = "PackageRepository"
	KindSecret                          = "Secret"
	KindSecretExport                    = "SecretExport"
	KindServiceAccount                  = "ServiceAccount"
	PackagingAPINotAvailable            = "package plugin can not be used as '%s/%s' API is not available in the cluster"
	PackagingAPIName                    = "packaging.carvel.dev"
	PackagingAPIVersion                 = "v1alpha1"
	SecretGenAPINotAvailable            = "secret plugin can not be used as '%s/%s' API is not available in the cluster"
	SecretGenAPIName                    = "secretgen.carvel.dev"
	SecretGenAPIVersion                 = "v1alpha1"
	SecretName                          = "%s-%s-values"
	ServiceAccountName                  = "%s-%s-sa"
	ShortDescriptionMaxLength           = 20
	TanzuPkgPluginAnnotation            = "tkg.tanzu.vmware.com/tanzu-package"
	TanzuPkgPluginPrefix                = "tanzu-package"
	TanzuPkgPluginResource              = "%s-%s"
	YamlSeparator                       = "---"
)
