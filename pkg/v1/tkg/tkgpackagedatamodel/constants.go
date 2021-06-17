// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint
package tkgpackagedatamodel

const (
	DefaultAPIVersion        = "install.package.carvel.dev/v1alpha1"
	TanzuPkgPluginAnnotation = "tkg.tanzu.vmware.com/tanzu-package"
	TanzuPkgPluginPrefix     = "tanzu-package-"
	ClusterRoleBindingName   = "%s-%s-cluster-rolebinding"
	ClusterRoleName          = "%s-%s-cluster-role"
	KindClusterRole          = "ClusterRole"
	KindClusterRoleBinding   = "ClusterRoleBinding"
	KindInstalledPackage     = "InstalledPackage"
	KindNamespace            = "Namespace"
	KindSecret               = "Secret"
	KindServiceAccount       = "ServiceAccount"
	KindPackageRepository    = "PackageRepository"
	PackageRepositoryKind    = "PackageRepository"
	SecretName               = "%s-%s-values"
	ServiceAccountName       = "%s-%s-sa"
)
