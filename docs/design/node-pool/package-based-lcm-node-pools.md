# Node Pools configuration for ClusterClass-based clusters

This doc provides an overview of differences in behavior when using the node pool API against a cluster using a clusterclass definition. The one thing that all of these APIs have in common when operating against a clusterclass based cluster is the need to update the cluster, specifically the cluster topology.

## Delete

Delete is the most straight forward change. The cluster resource is updated to remove the definition of the machine deployment with the name of the node pool that is passed to the delete API. The API does still prohibit the deletion of the final node pool in the cluster.

```bash
tanzu cluster node-pool delete tkg-wc-vsphere -n md–0
```

## List

List retrieves the MachineDeployment resources as in the non clusterclass based case and matches these resources against the machine deployments in the cluster topology. This allows the list API to return the same MachineDeployment resources with the names updated to match the machine deployments defined in the cluster topology.

```bash
tanzu cluster node-pool list tkg-wc-vsphere
  NAME  NAMESPACE  PHASE      REPLICAS  READY  UPDATED  UNAVAILABLE
  md-0  default    Ready      1         1      1        0
```

## Properties on Node Pools in ClusterClass based Clusters

Properties that can be updated on node pools in clusterclass based clusters can be broken down into two types. Direct properties of the machine deployment topology resources and variables overrides. Direct properties, like replica count, modify the machine deployment topolgy direct. Variable overrides rely on the definition of variables at the clusterclass level. When deploying a cluster based on a clusterclass, the cluster definition must specify a number of these variables, which apply to all machine deployments by default. The machine deployment topology allows these variables to be overridden on a machine deployment basis.

The definitive source for properties that can be modified are the clusterclasses themselves. The TKGm clusterclasses all define a variable `worker` that contains worker node vm properties. This includes `instanceType` and `vmSize` for AWS and Azure respectively. The vSphere version includes properties for `diskGiB`, `memoryMiB`, and `numCPU`. TKGs defines top level variables for `vmClass` and `storageClass` as well as an array for `nodePoolVolumes` for customzing worker nodes. TKGm vSphere also defines a vcenter variable for updating properties related to the vcenter server.

All TKG clusterclasses define a `nodePoolLabels` variable that specifies the node labels on the worker nodes.

### Special Note

The node pool API provided by the Tanzu CLI requires the above variables to be defined on the clusterclass to support creating and updating node pools. If you use a custom clusterclass these variables are required as top level variable definitions in the clusterclass. It is still possible to get and delete node pools on custom clusterclass based clusters using the Tanzu CLI. There is a special case where create and update will work for custom clusterclasses. Specifically if the node pool definition passed to the Tanzu cli specifies at most the `replicas`, `az`, `name`, `workerClass`, and `tkrResolver`. In this case the `workerClass` must match a `workerClass` definition in the custom clusterclass.

## Set (Update)

When the Set API is called using the name of an existing node pool the update will only modify the replica count and labels of the named node pool, if provided. Any other provided options will be ignored. Specifically the cluster's topology will be update with a new replica count and a nodePoolLabels variable override if these values are provided.

## Set (Create)

Creating node pools can follow one of two paths. A user can pass a base machine deployment that is used as the starting point for the new node pool. A deep copy of the named base machine deployment acts as a basis on which the user passed properties will be layered. In general this means all of the variable overrides of the copied machine deployment will be updated with new values provided by the user. This is different from the case where a base machine deployment is not provided. In this case, variable overrides will be copied from the global variable definitions defined in the cluster topology and then the user provided values will be layered on those. Additionally when a user does not provide a base machine deployment, they must provide a workerClass and tkrResolver definition. TKG ClusterClasses have a `tkg-worker` class defined by default, and should be used unless a user have added their own worker class definition. The tkrResolver is the value of the tkr-resolver annotation.

Using a base machine deployment to create a new node pool.

```bash
tanzu cluster node-pool set tkg-wc-vsphere -f /path/to/node-pool.yml –-base-machine-deployment md-0
```

node-pool.yml

```yaml
name: np-1
replicas: 1
nodeMachineType: t3.large
```

Without using a base machine deployment to create a new node pool.

```bash
tanzu cluster node-pool set tkg-wc-vsphere -f /path/to/node-pool.yml
```

node-pool.yml

```yaml
name: np-1
replicas: 1
nodeMachineType: t3.large
workerClass: tkg-worker
tkrResolver: os-name=ubuntu,os-arch=amd64
```
