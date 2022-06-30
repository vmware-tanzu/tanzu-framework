# Management Cluster Deployment using ClusterClass

Below is the sequence of steps involved in creating a management cluster using ClusterClass.

1. Deploy a bootstrap Kind cluster
1. Deploy `kapp-controller` on bootstrap cluster
1. Deploy CAPI provider and CAPx infra provider on bootstrap cluster
1. Deploy `tkg composite package` to bootstrap cluster
    1. As part of this package installation we will be creating configmap with label `run.tanzu.vmware.com/additional-compatible-tkrs` to make default TKR compatible.
    1. As part of this package, a set of sub packages will also be installed, among which is a package containing the cluster class and templates appropriate for the target CAPI infrastructure provider
1. Wait for packages to get deployed on management-cluster
1. Apply `CCluster.yaml` to create management-cluster using default clusterclass
1. Wait for CAPx controllers to reconcile the CCluster and deploy a management-cluster on specified infra
1. Deploy `kapp-controller` on the management-cluster
1. Addon-controller running on bootstrap cluster will deploy addons-components like CNI, CPI, CSI etc on the management-cluster
1. Deploy CAPI provider and CAPx infra provider on management-cluster
1. Deploy `tkg composite package` to management-cluster
    * Note: Use `_ADDITIONAL_MANAGEMENT_COMPONENT_CONFIGURATION_FILE` env var to specify any additional YAML that should be deployed. This should be used for development and testing only.
1. Wait for packages to get deployed on management-cluster
1. Move cluster-api resources from the bootstrap-cluster to management-cluster
1. Deploy `telemetry` job if telemetry is enabled

(See [here](../../packages/README.md) for more information regarding defining
packages)

## framework package

The [framework](../../../packages/framework) package is a meta package that
provides PackageInstall CRs and Service Accounts needed for installing the
following packages:

1. Featuregates
2. Addons-manager
3. TKR Service
4. Tanzu Auth Service
5. CLI Plugins

## tkg composite package

The [tkg](../../../packages/tkg) package is a composite package that installs
the following components via sub-packages:

1. The above-mentioned framework package
2. Core management plugins specific to the management cluster being deployed
3. TKR controllers
4. ClusterClass and associated templates
