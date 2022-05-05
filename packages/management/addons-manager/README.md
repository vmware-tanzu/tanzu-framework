# addons-manager Package

Addons-manager generates addon secrets which contain the configurations for addons based on TKR/BOM, KubeadmControlPlane and CLI inputs provided during cluster creation.
It then creates these PackageInstall CRs on management and workload clusters which then get reconciled by a local kapp-controller.

## Components

Client Library
Addon Controller

## Usage Example

The following is a basic guide for getting started with addons-manager.

Management cluster

1. “tanzu management-cluster create” to create a management cluster.

2. tanzu-addons-controller-manager will be deployed when cluster is being created
