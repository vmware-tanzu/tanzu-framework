# AWS Spreading of Nodes on a Prod Cluster

tkgctl library supports spreading nodes of a prod cluster across AZs. This is supported when using an existing VPC or by having library create a VPC on your behalf.

## Functionality Changes

For `dev` and `prod` clusters, in the kickstart UI you no longer need to provide the CIDRs for the subnet(s) if you are choosing to create a new VPC. Library divides the subnet CIDRs for you based on the VPC CIDR you provide.

For `prod` clusters, TKG now requires that you provide a unique AZ for each of the 3 control plane nodes. There is no option to create a prod cluster in one or two AZs. This means that you must create a prod management cluster in a region that has at least 3 AZs. In addition, it is required that you specify which AZs you would like to place the nodes in.

## Prod Cluster Template Changes

The prod cluster plan now expects additional AZ and subnet configurations variables to support this feature. These variables can be populated in two different ways depending on if you are using an existing VPC or would like to create a new one as shown in the following examples:

### Configuration Variables for Existing VPC

```yaml
AWS_VPC_ID: "vpc_id"
AWS_PRIVATE_SUBNET_ID: "private_subnet0_id"
AWS_PRIVATE_SUBNET_ID_1: "private_subnet1_id"
AWS_PRIVATE_SUBNET_ID_2: "private_subnet2_id"
AWS_PUBLIC_SUBNET_ID: "public_subnet0_id"
AWS_PUBLIC_SUBNET_ID_1: "public_subnet1_id"
AWS_PUBLIC_SUBNET_ID_2: "public_subnet2_id"
AWS_NODE_AZ: "az0"
AWS_NODE_AZ_1: "az1"
AWS_NODE_AZ_2: "az2"
```

### Configuration Variables for New VPC

```yaml
AWS_VPC_CIDR: 10.0.0.0/16
AWS_NODE_AZ: "az0"
AWS_PUBLIC_NODE_CIDR: 10.0.0.0/24
AWS_PRIVATE_NODE_CIDR: 10.0.1.0/24
AWS_NODE_AZ_1: "az1"
AWS_PUBLIC_NODE_CIDR_1: 10.0.2.0/24
AWS_PRIVATE_NODE_CIDR_1: 10.0.3.0/24
AWS_NODE_AZ_2: "az2"
AWS_PUBLIC_NODE_CIDR_2: 10.0.4.0/24
AWS_PRIVATE_NODE_CIDR_2: 10.0.5.0/24
```

## Dev Cluster Template Changes

The configuration variables expected by dev cluster plan remain unchanged, as shown in the following examples to use for an existing VPC and new VPC:

### Configuration Variables for Existing VPC

```yaml
AWS_VPC_ID: "vpc_id"
AWS_PRIVATE_SUBNET_ID: "private_subnet0_id"
AWS_PUBLIC_SUBNET_ID: "public_subnet0_id"
AWS_NODE_AZ: "az0"
```

### Configuration Variables for New VPC

```yaml
AWS_VPC_CIDR: 10.0.0.0/16
AWS_NODE_AZ: "az0"
AWS_PUBLIC_NODE_CIDR: 10.0.0.0/24
AWS_PRIVATE_NODE_CIDR: 10.0.1.0/24
```

## Spreading Machine Deployments for Workload Clusters in AWS

**Goal:** Spread worker nodes / machine deployments of prod workload clusters across AZs

Spreading AWSMachines across multiple AZs within a MachineDeployment is not yet supported by CAPA natively. However, there is an issue open regarding it:
[https://github.com/kubernetes-sigs/cluster-api/issues/3358](https://github.com/kubernetes-sigs/cluster-api/issues/3358)

Currently, until CAPA supports this natively, the recommended pattern to achieve this is to create multiple MachineDeployments, each deploying AWSMachines to a different AZ.

So if a user specified a `--worker-node-count` of 9 for a prod workload cluster, TKG would create 3 MachineDeployments each with a `replica` of 3 and each with a unique `failureDomain` value (3 AZs are required since this is a prod cluster).

Workload clusters on plan DEV are unaffected by this change.

### Template changes required

- Set `failureDomain` property of `Machine` spec
- Conditionally add 2 additional MachineDeployments, AWSMachineTemplates, and KubeadmConfigTemplates if the plan is PROD
- Set the replica count for each `MachineDeployment` accordingly
- Update the default `MachineDeployment`, `AWSMachineTemplate`, and `KubeadmConfigTemplate` so the YTT matchers to include `metdata.name`

### CLI changes required

- Add three new config variables, `WORKER_MACHINE_COUNT_0`, `WORKER_MACHINE_COUNT_1`, and `WORKER_MACHINE_COUNT_2`
- Add logic to `validate.go` to divide the number of worker machines evenly across the 3 MachineDeployments if the plan is PROD and `WORKER_MACHINE_COUNT_0`, `WORKER_MACHINE_COUNT_1` and `WORKER_MACHINE_COUNT_2` are not provided.
- Update `client/scale.go` to distribute the new number of workers evenly in `ScaleCluster()` using the same logic in `validate.go`
- Handle the situation where PROD management clusters only have 1 worker node by distributing the 1 worker replica value across the 3 MDs like 1,0,0.

#### Distribute Nodes Logic

```python
num_workers_per_az = floor(worker-node-count / 3)
remainder = worker-node-count % 3
azs[0] = num_workers_per_az
azs[1] = num_workers_per_az
azs[2] = num_workers_per_az
for a in azs:
  if remainder > 0:
    a += 1
    remainder -= 1
```

## Template Changes

`providers/infrastructure-aws/ytt/aws-overlay.yaml`:

```yaml
#@ if data.values.CLUSTER_PLAN == "prod":
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: #@ "{}-md-1".format(data.values.CLUSTER_NAME)
spec:
  template:
    spec:
      instanceType: #@ data.values.NODE_MACHINE_TYPE
      iamInstanceProfile: "nodes.cluster-api-provider-aws.sigs.k8s.io"
      sshKeyName: #@ data.values.AWS_SSH_KEY_NAME
      ami:
        id: #@ getattr(bomData.ami, data.values.AWS_REGION).id
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: #@ "{}-md-2".format(data.values.CLUSTER_NAME)
spec:
  template:
    spec:
      instanceType: #@ data.values.NODE_MACHINE_TYPE
      iamInstanceProfile: "nodes.cluster-api-provider-aws.sigs.k8s.io"
      sshKeyName: #@ data.values.AWS_SSH_KEY_NAME
      ami:
        id: #@ getattr(bomData.ami, data.values.AWS_REGION).id
---
apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
kind: KubeadmConfigTemplate
metadata:
  name: #@ "{}-md-1".format(data.values.CLUSTER_NAME)
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          name: '{{ ds.meta_data.local_hostname }}'
          kubeletExtraArgs:
            cloud-provider: aws
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineDeployment
metadata:
  name: #@ "{}-md-1".format(data.values.CLUSTER_NAME)
spec:
  clusterName: #@ data.values.CLUSTER_NAME
  replicas: #@ data.values.WORKER_MACHINE_COUNT_1
  selector:
    matchLabels: null
  template:
    metadata:
      labels:
        node-pool: #@ "{}-worker-pool".format(data.values.CLUSTER_NAME)
    spec:
      clusterName: #@ data.values.CLUSTER_NAME
      version: #@ data.values.KUBERNETES_VERSION
      bootstrap:
        configRef:
          name: #@ "{}-md-1".format(data.values.CLUSTER_NAME)
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
          kind: KubeadmConfigTemplate
      infrastructureRef:
        name: #@ "{}-md-1".format(data.values.CLUSTER_NAME)
        apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
        kind: AWSMachineTemplate
        failureDomain: #@ data.values.AWS_NODE_AZ_1
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineDeployment
metadata:
  name: #@ "{}-md-2".format(data.values.CLUSTER_NAME)
spec:
  clusterName: #@ data.values.CLUSTER_NAME
  replicas: #@ data.values.WORKER_MACHINE_COUNT_2
  selector:
    matchLabels: null
  template:
    metadata:
      labels:
        node-pool: #@ "{}-worker-pool".format(data.values.CLUSTER_NAME)
    spec:
      clusterName: #@ data.values.CLUSTER_NAME
      version: #@ data.values.KUBERNETES_VERSION
      bootstrap:
        configRef:
          name: #@ "{}-md-2".format(data.values.CLUSTER_NAME)
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
          kind: KubeadmConfigTemplate
      infrastructureRef:
        name: #@ "{}-md-2".format(data.values.CLUSTER_NAME)
        apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
        kind: AWSMachineTemplate
        failureDomain: #@ data.values.AWS_NODE_AZ_2
---
apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
kind: KubeadmConfigTemplate
metadata:
  name: #@ "{}-md-2".format(data.values.CLUSTER_NAME)
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          name: '{{ ds.meta_data.local_hostname }}'
          kubeletExtraArgs:
            cloud-provider: aws
#@end
```

`provider/infrastructure-aws/v0.5.5/ytt/overlay.yaml`:

```yaml
#@overlay/match by=overlay.subset({"kind": "AWSMachineTemplate", "metadata":{"name": "${CLUSTER_NAME}-md-0"}})
---
apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
kind: AWSMachineTemplate
metadata:
  name: #@ "{}-md-0".format(data.values.CLUSTER_NAME)
spec:
  template:
    spec:
      instanceType: #@ data.values.NODE_MACHINE_TYPE
      iamInstanceProfile: "nodes.cluster-api-provider-aws.sigs.k8s.io"
      sshKeyName: #@ data.values.AWS_SSH_KEY_NAME
      ami:
        id: #@ getattr(bomData.ami, data.values.AWS_REGION).id
#@overlay/match by=overlay.subset({"kind":"MachineDeployment", "metadata":{"name": "${CLUSTER_NAME}-md-0"}})
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineDeployment
metadata:
  name: #@ "{}-md-0".format(data.values.CLUSTER_NAME)
spec:
  clusterName: #@ data.values.CLUSTER_NAME
  replicas: #@ data.values.WORKER_MACHINE_COUNT_0
  selector:
    matchLabels: null
  template:
    metadata:
      labels:
        node-pool: #@ "{}-worker-pool".format(data.values.CLUSTER_NAME)
    spec:
      clusterName: #@ data.values.CLUSTER_NAME
      version: #@ data.values.KUBERNETES_VERSION
      bootstrap:
        configRef:
          name: #@ "{}-md-0".format(data.values.CLUSTER_NAME)
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
          kind: KubeadmConfigTemplate
      infrastructureRef:
        name: #@ "{}-md-0".format(data.values.CLUSTER_NAME)
        apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
        kind: AWSMachineTemplate
      failureDomain: #@ data.values.AWS_NODE_AZ
#@overlay/match by=overlay.subset({"kind":"KubeadmConfigTemplate", "metadata":{"name": "${CLUSTER_NAME}-md-0"}})
---
apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
kind: KubeadmConfigTemplate
metadata:
  name: #@ "{}-md-0".format(data.values.CLUSTER_NAME)
spec:
  template:
    spec:
      joinConfiguration:
        nodeRegistration:
          name: '{{ ds.meta_data.local_hostname }}'
          kubeletExtraArgs:
            cloud-provider: aws
```

`provider/infrastructure-aws/v0.5.5/ytt/base-template.yaml`:

```yaml
---
apiVersion: cluster.x-k8s.io/v1alpha3
kind: MachineDeployment
metadata:
  name: "${CLUSTER_NAME}-md-0"
spec:
  clusterName: "${CLUSTER_NAME}"
  replicas: ${WORKER_MACHINE_COUNT}
  selector:
    matchLabels:
  template:
    metadata:
      labels:
        node-pool: "${CLUSTER_NAME}-worker-pool"
    spec:
      clusterName: "${CLUSTER_NAME}"
      version: "${KUBERNETES_VERSION}"
      bootstrap:
        configRef:
          name: "${CLUSTER_NAME}-md-0"
          apiVersion: bootstrap.cluster.x-k8s.io/v1alpha3
          kind: KubeadmConfigTemplate
      infrastructureRef:
        name: "${CLUSTER_NAME}-md-0"
        apiVersion: infrastructure.cluster.x-k8s.io/v1alpha3
        kind: AWSMachineTemplate
      failureDomain: "${AWS_NODE_AZ}"
```
