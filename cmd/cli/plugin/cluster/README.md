# Cluster

Manage cluster lifecycle operations.

## Usage

```
>>> tanzu cluster create --help
Create a cluster

Usage:
  tanzu cluster create CLUSTER_NAME [flags]

Flags:
  -d, --dry-run       Does not create cluster but show the deployment YAML instead
  -f, --file string   Cluster configuration file from which to create a Cluster
  -h, --help          help for create
```
```
>>> tanzu cluster list --help
List clusters

Usage:
  tanzu cluster list [flags]

Flags:
  -h, --help                         help for list
      --include-management-cluster   Show active management cluster information as well
  -n, --namespace string             The namespace from which to list workload clusters. If not provided clusters from all namespaces will be returned
  -o, --output string                Output format. Supported formats: json|yaml
```
```
>>> tanzu cluster delete --help
Delete a cluster

Usage:
  tanzu cluster delete CLUSTER_NAME [flags]

Flags:
  -h, --help               help for delete
  -n, --namespace string   The namespace where the workload cluster was created. Assumes 'default' if not specified.
  -y, --yes                Delete workload cluster without asking for confirmation
```
```
>>> tanzu cluster scale --help
Scale a cluster

Usage:
  tanzu cluster scale CLUSTER_NAME [flags]

Flags:
  -c, --controlplane-machine-count int32   The number of control plane nodes to scale to. Assumes unchanged if not specified
  -h, --help                               help for scale
  -n, --namespace string                   The namespace where the workload cluster was created. Assumes 'default' if not specified.
  -w, --worker-machine-count int32         The number of worker nodes to scale to. Assumes unchanged if not specified
```
```
>>> tanzu cluster upgrade --help
Upgrade a cluster

Usage:
  tanzu cluster upgrade CLUSTER_NAME [flags]

Flags:
  -h, --help                        help for upgrade
  -k, --kubernetes-version string   The kubernetes version to upgrade to
  -n, --namespace string            The namespace where the workload cluster was created. Assumes 'default' if not specified
  -t, --timeout duration            Time duration to wait for an operation before timeout. Timeout duration in hours(h)/minutes(m)/seconds(s) units or as some combination of them (e.g. 2h, 30m, 2h30m10s) (default 30m0s)
  -y, --yes                         Upgrade workload cluster without asking for confirmation
```
```
>>> tanzu cluster machinehealthcheck --help
Get,set, or delete a MachineHealthCheck object for a Tanzu Kubernetes cluster

Usage:
  tanzu cluster machinehealthcheck [command]

Available Commands:
  delete      Delete machinehealthcheck of a cluster
  get         Get MachineHealthCheck object
  set         Create or update a MachineHealthCheck for a cluster
```
```
>>> tanzu cluster machinehealthcheck get --help
Get a MachineHealthCheck object for the given cluster

Usage:
  tanzu cluster machinehealthcheck get CLUSTER_NAME [flags]

Flags:
  -h, --help               help for get
  -m, --mhc-name string    Name of the MachineHealthCheck object
  -n, --namespace string   The namespace where the MachineHealthCheck object was created.
```
```
>>> tanzu cluster machinehealthcheck set --help
Create or update a MachineHealthCheck object for a cluster

Usage:
  tanzu cluster machinehealthcheck set CLUSTER_NAME [flags]

Flags:
  -h, --help                      help for set
  --match-labels string           Label selector to match machines whose health will be exercised
  -m, --mhc-name string               Name of the MachineHealthCheck object
  -n, --namespace string              Namespace of the cluster
  --node-startup-timeout string   Any machine being created that takes longer than this duration to join the cluster is considered to have failed and will be remediated
  --unhealthy-conditions string   A list of the conditions that determine whether a node is considered unhealthy. Available condition types: [Ready, MemoryPressure,DiskPressure,PIDPressure, NetworkUnavailable], Available condition status: [True, False, Unknown]heck object was created.
```
```
>>> tanzu cluster machinehealthcheck delete --help
Delete a MachineHealthCheck object for the given cluster

Usage:
  tanzu cluster machinehealthcheck delete CLUSTER_NAME [flags]

Flags:
  -h, --help               help for delete
  -m, --mhc-name string        Name of the MachineHealthCheck object
  -n, --namespace string   The namespace where the MachineHealthCheck object was created, default to the cluster's namespace
  -y, --yes                Delete the MachineHealthCheck object without asking for confirmation
```

```
>>> tanzu cluster credentials        
Update Credentials for Cluster

Usage:
  tanzu cluster credentials [command]

Available Commands:
  update      Update credentials for cluster

Flags:
  -h, --help   help for credentials

Use "cluster credentials [command] --help" for more information about a command.
```

```
>>> tanzu cluster credentials update --help
Update credentials for cluster

Usage:
  tanzu cluster credentials update CLUSTER_NAME [flags]

Flags:
  -h, --help                      help for update
  -n, --namespace string          The namespace of cluster whose credentials have to be updated
      --vsphere-password string   Password for vSphere provider
      --vsphere-user string       Username for vSphere provider
```

```
>>> tanzu cluster get --help
Getting clusters details

Usage:
  tanzu cluster get CLUSTER_NAME [flags]

Flags:
  --disable-grouping             Disable grouping machines when ready condition has the same Status, Severity and Reason
  --disable-no-echo              Disable hiding of a MachineInfrastructure and BootstrapConfig when ready condition is true or it has the Status, Severity and Reason of the machine's object
-h, --help                         help for get
-n, --namespace string             The namespace from which to get workload clusters. If not provided clusters from all namespaces will be returned
  --show-all-conditions string    list of comma separated kind or kind/name for which we should show all the object's conditions (all to show conditions for all the objects)
```