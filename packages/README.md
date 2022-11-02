# Packages

This directory contains two types of packages. Management Packages and Standalone Packages

## Management Packages

This directory contains some carvel packages for management components. Any package that performs the role of managing and operating a workload cluster is a Management package. These management packages are purpose-built to be installed on the management cluster. More details about packages can be found [here](../docs/packages/definitions.md)

These management packages are associated with each other in terms of [Meta Package](../docs/packages/definitions.md#meta-package). A Meta Package contains PackageInstall CRs, Secrets and ServiceAccounts, RBAC rules that are needed to install other packages and when the meta package is installed, it installs the other packages.

Below is the hierarchy of management packages association:

- tkg (TKG Meta Package)
  - framework (Framework Meta Package)
    - addons-manager
    - tkr-service
    - featuregates
    - tanzu-auth
    - cliplugins
  - object-propagation
  - tkg-clusterclass-docker
  - tkg-clusterclass-aws
  - tkg-clusterclass-azure
  - tkg-clusterclass-vsphere
  - core-management-plugins
  - tkr-source-controller

As part of the Bootstrap and Management Cluster creation, the `management-cluster` plugin installs the `tkg` meta package which installs all child packages. As part of this package installation the `management-cluster` plugin creates a [values.yaml](./tkg/bundle/config/values.yaml) file based on the user input. The configuration is then propagated to child packages based on overlays written within the package.

There are few management packages created under this directory which are not used at the moment and not integrated with the `tkg` meta package. Below is the list of these packages:

- cluster-api
- cluster-api-bootstrap-kubeadm
- cluster-api-control-plane-kubeadm
- cluster-api-provider-aws
- cluster-api-provider-azure
- cluster-api-provider-docker
- cluster-api-provider-vsphere

## Standalone Packages

This type of carvel packages are individual packages that don't depend on other packages. Below is the list of standalone packages:

- standalone-plugins
- capabilities
- tkg-autoscaler
- tkg-storageclass
