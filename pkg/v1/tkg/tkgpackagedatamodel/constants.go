// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint
package tkgpackagedatamodel

import "time"

const (
	ClusterRoleBindingName              = "%s-%s-cluster-rolebinding"
	ClusterRoleName                     = "%s-%s-cluster-role"
	DefaultAPIVersion                   = "install.package.carvel.dev/v1alpha1"
	DefaultPollInterval                 = 1 * time.Second
	DefaultPollTimeout                  = 5 * time.Minute
	DefaultRepositoryImageTag           = "latest"
	DefaultRepositoryImageTagConstraint = ">0.0.0"
	ErrPackageNotInstalled              = "package install does not exist in the namespace"
	ErrRepoNotExists                    = "package repository does not exist in the namespace"
	KindClusterRole                     = "ClusterRole"
	KindClusterRoleBinding              = "ClusterRoleBinding"
	KindNamespace                       = "Namespace"
	KindPackageInstall                  = "PackageInstall"
	KindPackageRepository               = "PackageRepository"
	KindSecret                          = "Secret"
	KindSecretExport                    = "SecretExport"
	KindServiceAccount                  = "ServiceAccount"
	SecretName                          = "%s-%s-values"
	ServiceAccountName                  = "%s-%s-sa"
	ShortDescriptionMaxLength           = 20
	TanzuPkgPluginAnnotation            = "tkg.tanzu.vmware.com/tanzu-package"
	TanzuPkgPluginPrefix                = "tanzu-package"
	TanzuPkgPluginResource              = "%s-%s"
	YamlSeparator                       = "---"
)
