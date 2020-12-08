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
  -c, --controlplane-machine-count int32   The number of control plane nodes to scale to
  -h, --help                               help for scale
  -n, --namespace string                   The namespace where the workload cluster was created. Assumes 'default' if not specified.
  -w, --worker-machine-count int32         The number of worker nodes to scale to
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