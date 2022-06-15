# Management Cluster Deployment using ClusterClass

Below is the sequence of management-cluster creation using clusterclass.

1. Deploy a bootstrap kind cluster
1. Deploy `kapp-controller` on bootstrap cluster
1. Deploy `providers` on bootstrap cluster
1. Deploy `TKG meta package` to bootstrap cluster
    1. As part of this package installation we will be creating configmap with label `run.tanzu.vmware.com/additional-compatible-tkrs` to make default TKR compatible.
1. Wait for packages to get deployed on management-cluster
1. Apply `CCluster.yaml` to create management-cluster using default clusterclass
1. Wait for CAPx controllers to reconcile the CCluster and deploy a management-cluster on specified infra
1. Deploy `kapp-controller` on the management-cluster
1. Addson-controller running on bootstrap cluster will deploy addons-components like CNI, CPI, CSI etc on the management-cluster
1. Deploy `providers` on management-cluster
1. Deploy `TKG meta package` to management-cluster
    * Hack: Using `_ADDITIONAL_MANAGEMENT_COMPONENT_CONFIGURATION_FILE` to deploy additional components that are not getting deployed as part of `TKG meta package` at the moment. This includes TKR package related resource. This is only for testing purpose and should be removed once TKG package related changes are in.
1. Wait for packages to get deployed on management-cluster
1. Move cluster-api resources from the bootstrap-cluster to management-cluster
1. Deploy `telemetry` job if telemetry is enabled
