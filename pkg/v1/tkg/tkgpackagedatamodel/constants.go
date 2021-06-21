// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint
package tkgpackagedatamodel

import "time"

const (
	DefaultAPIVersion         = "install.package.carvel.dev/v1alpha1"
	TanzuPkgPluginAnnotation  = "tkg.tanzu.vmware.com/tanzu-package"
	TanzuPkgPluginPrefix      = "tanzu-package-"
	ClusterRoleBindingName    = "%s-%s-cluster-rolebinding"
	ClusterRoleName           = "%s-%s-cluster-role"
	DefaultPollInterval       = 1 * time.Second
	DefaultPollTimeout        = 5 * time.Minute
	KindClusterRole           = "ClusterRole"
	KindClusterRoleBinding    = "ClusterRoleBinding"
	KindInstalledPackage      = "InstalledPackage"
	KindNamespace             = "Namespace"
	KindSecret                = "Secret"
	KindServiceAccount        = "ServiceAccount"
	KindPackageRepository     = "PackageRepository"
	PackageRepositoryKind     = "PackageRepository"
	SecretName                = "%s-%s-values"
	ServiceAccountName        = "%s-%s-sa"
	ShortDescriptionMaxLength = 20
)
