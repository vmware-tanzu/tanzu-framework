# Tanzu Diagnostics Plugin

The Tanzu Diagnostics Plugin allows operators to collect cluster diagnostics data.
The plugin uses the  [Crashd API](https://github.com/vmware-tanzu/crash-diagnostics) to run internal script that automates the collection of 
diagnostics information for cluster troubleshooting.

## Examples
### Collecting diagnostics
By default, the diagnostics plugin will collect information from  the currently selected managed workload cluster
and its associated management cluster. 
For instance, the following will collect logs, API objects, and other API server info
from for the currently selected workload cluster and the management cluster:

```bash
tanzu diagnostics collect
```
The previous command will collect:
* Diagnostics for the currently selected workload cluster
* Diagnostics for the management cluster
* Diagnostics for the bootstrap cluster (if available)


### Skipping management cluster
Diagnostics data for the management or the bootstrap cluster can be skipped:
```bash
tanzu diagnostics collect --skip-bootstrap-cluster --skip-management-cluster
```
The command above will collect diagnostics only for the default managed workload server.

### Collecting diagnostics for specific workload servers
To collect diagnostics for specific a specific managed workload cluster, specify the name of the cluster as follows:
```bash
tanzu diagnostics collect --workload-cluster-name=wc-cluster --workload-cluster-namespace=my-clusters
```
The previous snippet will extract diagnostics for workload cluster `wc-cluster`
running in managed cluster namespace `my-clusters` managed by a management cluster.

### Collecting infrastructure diagnostics
The diagnostics plugin can collect infrastructure information from the machines running the cluster.
This is done by specifying the SSH user and private key used during the creation of the cluster:

```bash
tanzu diagnostics collect --workload-ssh-user=myssh --workload-ssh-pk=/path/to/private_key
```

## `diagnostics collect` arguments
The followings are optional arguments that can be used to override default values when collecting
diagnostics information from clusters.

#### Bootstrap cluster
Arguments to collect bootstrap cluster diagnostics:
* `--skip-bootstrap-cluster` - if present, bootstrap cluster diagnostics are ignored
* `--bootstrap-cluster-name` - provides a specific bootstrap cluster name to diagnose

#### Managed/standalone workload cluster
For managed cluster collection only
* `--workload-cluster-name` - specifies the name of the managed cluster for which to collect diagnostics
* `--workload-cluster-namespace` - The namespace associated with managed cluster resources

For managed/standalone collection
* `--workload-cluster-infra` - overrides the infrastructure type for the managed cluster (i.e. aws, azure, vsphere, etc)
* `--workload-cluster-kubeconfig` - if present, overrides the kubeconfig for the managed workload cluster
* `--workload-cluster-context` - if present, overrides the name of the context name of the workload cluster
* `--workload-ssh-user` - specifies SSH user to log unto workload cluster machines. If presents, used to collect machine diagnostics.
* `--workload-ssh-pk` - specifies SSH private key path to log unto machines. If presents, used to collect machine diagnostics.

#### Management cluster
Arguments for accessing diagnostics for management cluster
* `--skip-management-cluster` - if present, management cluster diagnostics are skipped
* `--management-kubeconfig` - overrides the path of the management server kubeconfig file
* `--management-context` - overrides the name of the cluster context
* `--management-ssh-user` - specifies SSH user to log unto management machines. If present, used to collect machine diagnostics.
* `--management-ssh-pk` - specifies SSH private key path used to log unto management machines. If present, used to collect machine diagnostics.