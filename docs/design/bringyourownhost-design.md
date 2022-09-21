# Bring-Your-Own-Host Design

## Abstract

The Bring Your Own Host aka BYOH aims at enabling Kubernetes cluster life cycle management on user provided hosts (bare metal hosts in most cases).

Currently some infrastructure providers are supported in tanzu-framework(e.g. vSphere, AWS, Azure). This is to enhance current tanzu-framework to support BYOH as infrastructure provider, so that users can use tanzu CLI to create BYOH workload cluster.

## Background

Consider the multiple provider work covers multiple components in Tanzu product family and from BYOH team's point of view we are searching to enable customer to be able to create BYOH workload cluster with Tanzu CLI as a start point. We shape the project into 3 phases and firstly we intend to implement the "crawl phase":

Crawl Phase

1. Document on patch the existing Management cluster with BYOH provider

2.Enhance the ```tanzu cluster create``` CLI so customer can use it to create BYOH cluster.

Walk Phase

1. Implement the tanzu management-cluster provider CLI so customer can use it to patch management cluster with new provider

2. Support patching BYOH provider

3. Support install BYOH provider when create management cluster with ```tanzu management-cluster create``` CLI.

Run Phase

1. Support lifecycle management of AWS/Azure/Any other provider with ```tanzu management-cluster provider``` CLI
  
2. Support create workload cluster with selected provider with multiple providers installed management cluster.
  
3. Support creates management cluster directly on BYOH hosts.

## Goals

* Define the configurations used for BYOH to create BYOH workload cluster.
* Use the existing ```tanzu cluster create``` command to create BYOH workload cluster using the configuration file.
* Use the existing ```tanzu cluster list/delete/scale``` command to list/delete/update BYOH workload cluster.

## High-Level Design

This is to describe what will be targeted for the first phase.

### Management cluster

To deploy a management cluster with BYOH integrated, firstly you create a command vSphere management cluster.
Then you patch the management cluster to have the 'byoh' provider included.

1. Using the ```tanzu management-cluster create``` command to create a common management cluster.

2. Install the "clusterctl" binary and use it to patch the management cluster created by step 1.
   clusterctl, which can be downloaded from the latest [release][releases] of Cluster API (CAPI) on GitHub.
   To learn more about cluster API in more depth, check out the the [Cluster API book][cluster-api-book].
  
3. Configuring and installing BringYourOwnHost provider in a management cluster  
  To initialize Cluster API Provider BringYourOwnHost, clusterctl requires the following settings, which should be set in `~/.cluster-api/clusterctl.yaml` as the following:

``` yaml
providers:
  - name: byoh
    url: https://github.com/vmware-tanzu/cluster-api-provider-bringyourownhost/releases/latest/infrastructure-components.yaml
    type: InfrastructureProvider  
```

running `clusterctl config repositories`.

You should be able to see the new BYOH provider there.

```shell
clusterctl config repositories
NAME           TYPE                     URL                                                                                          FILE
cluster-api    CoreProvider             https://github.com/kubernetes-sigs/cluster-api/releases/latest/                              core-components.yaml
...
byoh           InfrastructureProvider   https://github.com/vmware-tanzu/cluster-api-provider-bringyourownhost/releases/latest/       infrastructure-components.yaml
...
vsphere        InfrastructureProvider   https://github.com/kubernetes-sigs/cluster-api-provider-vsphere/releases/latest/             infrastructure-components.yaml
```

Install the BYOH provider

```shell
clusterctl init --infrastructure byoh
```

### Workload cluster

To deploy a Kubernetes workload cluster, you create a configuration file that specifies the different options with which to deploy the cluster. You then run the tanzu cluster create command, specifying the configuration file in the ```--file``` option.  
When you deploy a Kubernetes cluster, most of the configuration for the cluster is the same as the configuration of the management cluster that you use to deploy it.  

1. Specify the type of infrastructure provider to "BringYourOwnHost" with ```INFRASTRUCTURE_PROVIDER``` variable.
  ```INFRASTRUCTURE_PROVIDER: byoh```
  
2. Set a name for the cluster in the ```CLUSTER_NAME``` variable.
For example if you are deploying the BYOH cluster, set the name to "my-byoh-tkc"
  ```CLUSTER_NAME: my-byoh-tkc```  
  
3. If you are deploying the cluster to BYOH, specify a static virtual IP address or FQDN in the ```BYOH_CONTROL_PLANE_ENDPOINT``` variable.
   Specify a port in the "BYOH_CONTROL_PLANE_ENDPOINT_PORT" variable. (If not specified, by default the port is: 6443)
  ```BYOH_CONTROL_PLANE_ENDPOINT: 10.90.110.100```
  ```BYOH_CONTROL_PLANE_ENDPOINT_PORT: 6443```  
  
4. You can also deploy a BYOH workload cluster with Custom Control Plane and Worker Node Counts.
In this version, the dev and prod plans for Tanzu Kubernetes clusters deploy the following:  
The dev plan: one control plane node and one worker node.  
The prod plan: three control plane nodes and three worker nodes.  
  To deploy a Tanzu Kubernetes cluster with more control plane nodes than the dev and prod plans define by default, specify the ```CONTROL_PLANE_MACHINE_COUNT``` variable in the cluster configuration file. The number of control plane nodes that you specify in ```CONTROL_PLANE_MACHINE_COUNT``` must be uneven:  
  ```CONTROL_PLANE_MACHINE_COUNT: 5```  
  Specify the number of worker nodes for the cluster in the WORKER_MACHINE_COUNT variable. For example:  
  ```WORKER_MACHINE_COUNT: 5```
  
5. Finally, you get a configuration file below, then save the configuration file as "my-byoh-tkc.yaml".

   ```yaml
   CLUSTER_CIDR: 100.96.1.0/11
   CLUSTER_NAME: my-byoh-tkc
   CLUSTER_PLAN: prod
   INFRASTRUCTURE_PROVIDER: byoh
   SERVICE_CIDR: 10.192.168.0/24
   BYOH_CONTROL_PLANE_ENDPOINT: 10.90.110.100
   BYOH_CONTROL_PLANE_ENDPOINT_PORT: 6443
   WORKER_MACHINE_COUNT: 5
   CONTROL_PLANE_MACHINE_COUNT: 5
   ```  

6. Run the ```tanzu cluster create``` command, specifying the path to the configuration file in the ```--file``` option.  
  ```tanzu cluster create --file ./my-byoh-tkc.yaml```  

7. To see information about the cluster, run the ```tanzu cluster get``` command, specifying the cluster name.  
  ```tanzu cluster get my-byoh-tkc```

8. You can also scale a cluster horizontally with the Tanzu CLI. To horizontally scale a Tanzu Kubernetes cluster, use the ```tanzu cluster scale``` command. You change the number of control plane nodes by specifying the ```--controlplane-machine-count``` option. You change the number of worker nodes by specifying the ```--worker-machine-count``` option.
   To scale a cluster that you originally deployed with 3 control plane nodes and 5 worker nodes to 7 and 10 nodes respectively, run the following command:  
```tanzu cluster scale <cluster_name> --controlplane-machine-count 7 --worker-machine-count 10```

## Detailed Design

1. Add the “**infrastructure-byoh**” folder which contains the YTT templates to generate the YAML files, so that CAPI can provision BYOH workload clusters using these YAML files.
Update "providers/config.yaml" and create new YTT under folder " providers/infrastructure-byoh/v0.1.0/" to support BYOH as a new type of provider.

2. Add the configuration item to let users define control plane endpoint for BYOH workload cluster.
Update "providers/config_default.yaml" and tanzu-framework go code, so that new configurations can be read to create BYOH workload cluster.

3. Configurations to provision BYOH workload clusters should be validated in case of invalid configuration input.

<!-- References -->
[cluster-api-book]: https://cluster-api.sigs.k8s.io/
[releases]: https://github.com/kubernetes-sigs/cluster-api/releases
