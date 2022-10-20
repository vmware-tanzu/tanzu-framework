// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

// ConfigFilePermissions defines the permissions of the config file
const (
	ConfigFilePermissions       = 0o600
	DefaultDirectoryPermissions = 0o700
)

// File name related constants
const (
	LocalProvidersFolderName  = "providers"
	LocalProvidersZipFileName = "providers.zip"
	LocalTanzuFileLock        = ".tanzu.lock"

	LocalProvidersConfigFileName = "config.yaml"
	LocalBOMsFolderName          = "bom"
	LocalCompatibilityFolderName = "compatibility"

	LocalProvidersChecksumFileName = "providers.sha256sum"
	OverrideFolder                 = "overrides"

	TKGKubeconfigDir    = ".kube-tkg"
	TKGKubeconfigFile   = "config"
	TKGKubeconfigTmpDir = "tmp"

	TKGConfigFileName               = "config.yaml"
	TKGDefaultClusterConfigFileName = "cluster-config.yaml"
	TKGCompatibilityFileName        = "tkg-compatibility.yaml"
	TKGConfigDefaultFileName        = "config_default.yaml"

	TKGClusterConfigFileDirForUI           = "clusterconfigs"
	TKGRegistryCertFile                    = "registry_certs"
	TKGRegistryTrustedRootCAFileForWindows = ".registry_trusted_root_certs_win"

	LogFolderName = "logs"

	TKGPackageValuesFile = "tkgpackagevalues.yaml"
)

// Config content
const (
	AKOPackageInstall = `#@ load("@ytt:data", "data")
#@ load("@ytt:yaml", "yaml")

---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
  name: load-balancer-and-ingress-service
  namespace: tkg-system
  annotations:
    kapp.k14s.io/disable-wait: ""
  labels:
    tkg.tanzu.vmware.com/package-type: "management"
spec:
  packageRef:
    refName: load-balancer-and-ingress-service.tanzu.vmware.com
    versionSelection:
      prereleases: {}
  serviceAccountName: load-balancer-and-ingress-service-package-sa
  values:
  - secretRef:
      name: #@ "{}-load-balancer-and-ingress-service-addon".format(data.values.CLUSTER_NAME)
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: load-balancer-and-ingress-service-package-sa
  namespace: tkg-system
  annotations:
    kapp.k14s.io/change-group: "load-balancer-and-ingress-service-packageinstall/serviceaccount-0"
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: load-balancer-and-ingress-service-package-role
  annotations:
    kapp.k14s.io/change-group: "load-balancer-and-ingress-service-packageinstall/serviceaccount-0"
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["*"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: load-balancer-and-ingress-service-package-cluster-rolebinding
  annotations:
    kapp.k14s.io/change-group: "load-balancer-and-ingress-service-packageinstall/serviceaccount"
    kapp.k14s.io/change-rule.0: "upsert after upserting load-balancer-and-ingress-service-packageinstall/serviceaccount-0"
    kapp.k14s.io/change-rule.1: "delete before deleting load-balancer-and-ingress-service-packageinstall/serviceaccount-0"
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: load-balancer-and-ingress-service-package-role
subjects:
  - kind: ServiceAccount
    name: load-balancer-and-ingress-service-package-sa
    namespace: tkg-system

`
)
