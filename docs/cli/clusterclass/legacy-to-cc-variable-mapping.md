# Mapping legacy configuration values to cluster class fields/variables

## Introduction

The configuration variable space used for configuring Tanzu Clusters in the
Tanzu CLI has roots in CAPI/ClusterCtl, has persisted since the its
introduction in the TKG CLI (Tanzu CLI’s predecessor) and still to this day
employs a flat key-value style.

As part of the backwards compatibility handling to support the use of legacy
configuration files based on the above, these legacy variables are being mapped
to ClusterClass fields and variables. We will levarage the flexibility afforded by
OpenAPI v3 schema types to better organize and name the variables.

## Scope

General efforts are being made to consolidate the variables and provide some
consistency in the naming and casing style of the variables.

Below are sample representations of the variables section in the default cluster
classes provided for each the CAPA/CAPZ/CAPV infrastructure providers supported
in this repository and the corresponding legacy configuration values they correspond to.

<!-- markdownlint-capture -->
<!-- markdownlint-disable MD033 MD012 -->

### AWS Provider

<table>
<tr>
<td> ClusterClass Variables  </td>
<td> Legacy Cluster Config Values </td>
</tr>
<tr>
<td valign="top">

```yaml
variables:
- name: proxy
  value:
    httpProxy: http://10.0.0.1
    httpsProxy: https://10.0.0.1
    noProxy:
    - .svc.cluster.local
    - 192.168.1.1
    - …
- name: imageRepository
  value:
    host: stg-project.vmware.com
    tlsCertificateValidation:
      enabled: true
- name: trust
  value:
    - name: proxy
      data: LS0tLS1…
    - name: imageRepository
      data: LS0tLS1…
- name: plan
  value: prodcc
- name: region
  value: us-west-2
- name: vpc
  value:
    cidr: 10.0.0.0/16
    id: ""
- name: identityRef
  value:
    kind: AWSClusterRoleIdentity
    name: ""
- name: loadBalancerSchemeInternal
  value: false
- name: bastionHost
  value:
    enabled: true
- name: securityGroup
  value:
    controlPlane: sg-12345
    apiServerLB: sg-12345
    bastion: sg-12345
    lb: sg-12345
    node: sg-12345
- name: subnets
  value:
  - private:
      cidr: 10.0.0.0/24
      id: ""
    public:
      cidr: 10.0.1.0/24
      id: ""
    az: us-west-2a
  - private:
      cidr: 10.0.2.0/24
      id: ""
    public:
      cidr: 10.0.3.0/24
      id: ""
    az: us-west-2b
  - private:
      cidr: 10.0.4.0/24
      id: ""
    public:
      cidr: 10.0.5.0/24
      id: ""
    az: us-west-2c
- name: worker
  value:
    instanceType: m5.xlarge
    rootVolume:
      sizeGiB: 80
- name: controlPlane
  value:
    instanceType: t3.medium
    rootVolume:
      sizeGiB: 80

workers: (see +++)
  machineDeployments:
  - class: tkg-worker
    name: md-0
    replicas: 2
    failureDomain: us-west-2a
    variables:
      overrides:
      - name: worker
        value:
          instanceType: m5.small
          rootVolume:
            sizeGiB: 80
  - class: tkg-worker
    name: md-1
    replicas: 2
    failureDomain: us-west-2b
    variables:
      overrides:
      - name: worker
        value:
          instanceType: t3.small
          rootVolume:
            sizeGiB: 40
  - class: tkg-worker
    name: md-2
    replicas: 2
    failureDomain: us-west-2c
    variables:
      overrides:
      - name: worker
        value:
          instanceType: t3.xlarge
          rootVolume:
            sizeGiB: 120
```

</td>
<td valign="top">

```yaml

TKG_HTTP_PROXY_ENABLED: true

TKG_HTTP_PROXY: http://10.0.0.1
TKG_HTTPS_PROXY: https://10.0.0.1
TKG_NO_PROXY: 192.168.1.1





TKG_CUSTOM_IMAGE_REPOSITORY: stg-project.vmware.com
TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY: false


TKG_PROXY_CA_CERT: LS0tLS1…

TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE: LS0tLS1…

CLUSTER_PLAN: prod

AWS_REGION: us-west-2



AWS_VPC_CIDR: 10.0.0.0/16
AWS_VPC_ID: “”


AWS_IDENTITY_REF_KIND: AWSClusterRoleIdentity
AWS_IDENTITY_REF_NAME:
AWS_LOAD_BALANCER_SCHEME_INTERNAL: false

BASTION_HOST_ENABLED: true



AWS_SECURITY_GROUP_CONTROLPLANE: sg-12345
AWS_SECURITY_GROUP_APISERVER_LB: sg-12345
AWS_SECURITY_GROUP_BASTION: sg-12345
AWS_SECURITY_GROUP_LB: sg-12345
AWS_SECURITY_GROUP_NODE: sg-12345



AWS_PRIVATE_NODE_CIDR: 10.0.0.0/24
AWS_PRIVATE_NODE_ID:

AWS_PUBLIC_NODE_CIDR: 10.0.1.0/24
AWS_PUBLIC_NODE_ID:
AWS_NODE_AZ: us-west-2a

AWS_PRIVATE_NODE_CIDR_1: 10.0.2.0/24
AWS_PRIVATE_NODE_ID_1:

AWS_PUBLIC_NODE_CIDR_1: 10.0.3.0/24
AWS_PUBLIC_NODE_ID_1:
AWS_NODE_AZ_1: us-west-2b

AWS_PRIVATE_NODE_CIDR_2: 10.0.4.0/24
AWS_PRIVATE_NODE_ID_2:

AWS_PUBLIC_NODE_CIDR_2: 10.0.5.0/24
AWS_PUBLIC_NODE_ID_2:
AWS_NODE_AZ_2: us-west-2c


NODE_MACHINE_TYPE: m5.xlarge

AWS_NODE_OS_DISK_SIZE_GIB: 80


CONTROL_PLANE_MACHINE_TYPE: t3.medium

AWS_CONTROL_PLANE_OS_DISK_SIZE_GIB: 80





WORKER_MACHINE_COUNT: 2
AWS_NODE_AZ: us-west-2a




NODE_MACHINE_TYPE: m5.small

AWS_NODE_OS_DISK_SIZE_GIB: 80


WORKER_MACHINE_COUNT_1: 2
AWS_NODE_AZ_1: us-west-2b




NODE_MACHINE_TYPE_1: t3.small

Not Present This value is inherited from AWS_NODE_OS_DISK_SIZE_GIB in current clusters

WORKER_MACHINE_COUNT_2: 2
AWS_NODE_AZ_2: us-west-2c
```

</td>

</tr>
</table>


_+++_: Providing the machineDeployments section as an example of how multi
node-pool deployments can be customized. It is omitted in Azure and vSphere,
but they will both behave in the same way.


### Azure Provider

<table>
<tr>
<td> ClusterClass Variables  </td>
<td> Legacy Cluster Config Values </td>
</tr>
<tr>
<td valign="top">

```yaml
variables:
- name: proxy
  value:
    httpProxy: http://10.0.0.1
    httpsProxy: https://10.0.0.1
    noProxy:
    - .svc.cluster.local
    - 192.168.1.1
    - …
- name: imageRepository
  value:
    host: stg-project.vmware.com
    tlsCertifcateValidation:
      enabled: false
- name: trust
  value:
    - name: proxy
      data: LS0tLS1…
    - name: imageRepository
      data: LS0tLS1…
- name: location
  value: westus
- name: resourceGroup
  value: ""
- name: subscriptionID
  value: 6c2a2ce1-649f-9d9k-a19c-c729h3cf6126
- name: environment
  value: AzurePublicCloud
- name: sshPublicKey
  value: c3NoLXJzYSBB
- name: acceleratedNetworking
  value:
    enabled: true
- name: enablePrivateCluster
  value: false
- name: frontendPrivateIP
  value: 22.22.22.22
- name: vnet
  value:
    cidr: 10.0.0.0/8
    name: azure-vnet-name
    resourceGroup: azure-vnet-resource-group
- name: identity
  value:
    name: test-identity1
    namespace: test-ns1
- name: controlPlane
  value:
    dataDisks:
      - sizeGiB: 256
    machineType: i3.xlarge
    osDisk:
      sizeGiB: 128
      storageAccountType: Premi
    outboundLB:
      enabled: true
      frontendIPCount: 3
    subnet:
      cidr: 10.0.0.0/24
      securityGroup: SecurityGro
- name: worker
  value:
    machineType: Standard_Dv2s
    osDisk:
      sizeGiB: 128
      storageAccountType: Premium_LRS
    dataDisks:
      - sizeGiB: 256
    outboundLB:
      enabled: true
      frontendIPCount: 1
      idleTimeoutInMinutes: 8
    subnet:
      cidr: 10.1.0.0/16
      securityGroup: SecurityGroup
```

</td>
<td valign="top">

```yaml

TKG_HTTP_PROXY_ENABLED: true

TKG_HTTP_PROXY: http://10.0.0.1
TKG_HTTPS_PROXY: https://10.0.0.1
TKG_NO_PROXY: 192.168.1.1





TKG_CUSTOM_IMAGE_REPOSITORY: stg-project.vmware.com
TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY: true



TKG_PROXY_CA_CERT: LS0tLS1…
TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE: LS0tLS1

AZURE_LOCATION: westus

AZURE_RESOURCE_GROUP:

AZURE_SUBSCRIPTION_ID: 6c2a2ce1-649f-9d9k-a19c

AZURE_ENVIRONMENT: AzurePublicCloud

AZURE_SSH_PUBLIC_KEY_B64: c3NoLXJzYSBB

AZURE_ENABLE_ACCELERATED_NETWORKING: true

AZURE_ENABLE_PRIVATE_CLUSTER: false

AZURE_FRONTEND_PRIVATE_IP: 22.22.22.22



AZURE_VNET_CIDR: 10.0.0.0/8
AZURE_VNET_NAME: azure-vnet-name
AZURE_VNET_RESOURCE_GROUP: azure-vnet-resource-group


AZURE_IDENTITY_NAME: test-identity1
AZURE_IDENTITY_NAMESPACE: test-ns1



AZURE_CONTROL_PLANE_DATA_DISK_SIZE: 512
AZURE_CONTROL_PLANE_MACHINE_TYPE: Standard_B2s

AZURE_CONTROL_PLANE_OS_DISK_SIZE: 256
AZURE_CONTROL_PLANE_OS_DISK_STORAGE_ACCOUNT_TYPE: Premi

AZURE_ENABLE_CONTROL_PLANE_OUTBOUND_LB: true
AZURE_CONTROL_PLANE_OUTBOUND_LB_FRONTEND_IP_COUNT: 3

AZURE_CONTROL_PLANE_SUBNET_CIDR: 10.1.0.0/16
AZURE_CONTROL_PLANE_SUBNET_SECURITY_GROUP: SecurityGro


AZURE_NODE_MACHINE_TYPE: Standard_Dv2s

AZURE_NODE_OS_DISK_SIZE: 256
AZURE_NODE_OS_DISK_STORAGE_ACCOUNT_TYPE: Premium_LRS
AZURE_ENABLE_NODE_DATA_DISK: true
AZURE_NODE_DATA_DISK_SIZE_GIB:

AZURE_ENABLE_NODE_OUTBOUND_LB: true
AZURE_NODE_OUTBOUND_LB_FRONTEND_IP_COUNT: 1
AZURE_NODE_OUTBOUND_LB_IDLE_TIMEOUT_IN_MINUTES: 8

AZURE_NODE_SUBNET_CIDR: 10.1.0.0/16
AZURE_NODE_SUBNET_SECURITY_GROUP: SecurityGroup
```

</td>

</tr>
</table>

### Management Cluster Based vSphere Provider

<table>
<tr>
<td> ClusterClass Variables  </td>
<td> Legacy Cluster Config Values </td>
</tr>
<tr>
<td valign="top">

```yaml
variables:
- name: proxy
  value:
    httpProxy: http://10.0.0.1
    httpsProxy: https://10.0.0.1
    noProxy:
    - .svc.cluster.local
    - 192.168.1.1
    - …
- name: imageRepository
  value:
    host: stg-project.vmware.com
    tlsCertificateValidation:
      enabled: false
- name: trust
  value:
    - name: proxy
      data: LS0tLS1…
    - name: imageRepository
      data: LS0tLS1…
- name: apiServerEndpoint
  value: 10.10.10.10
- name: aipServerPort
  value: 443
- name: vipNetworkInterface
  value: eth0
- name: aviControlPlaneHAProvider
  value: false
- name: auditLogging
  value:
    enabled: true
- name: vcenter
  value:
    cloneMode: fullClone
    datacenter: /dc0
    datastore: ds1
    folder: vm0
    network: TESTNETWORK
    resourcePool: rp0
    server: somehostname
    storagePolicyID: ""
    template: photon-3-v1.19.3+vmware.1
    tlsThumbprint: dummythumbprint
- name: controlPlane
  value:
    count: 5
    machine:
      diskGiB: 40
      memoryMiB: 8192
      numCPUs: 2
- name: worker
  value:
    count: 3
    machine:
      diskGiB: 20
      memoryMiB: 4096
      numCPUs: 2
- name: user
  value:
    sshAuthorizedKey: ssh-rsa A
```

</td>
<td valign="top">

```yaml

TKG_HTTP_PROXY_ENABLED: true

TKG_HTTP_PROXY: http://10.0.0.1
TKG_HTTPS_PROXY: https://10.0.0.1
TKG_NO_PROXY: 192.168.1.1





TKG_CUSTOM_IMAGE_REPOSITORY: stg-project.vmware.com
TKG_CUSTOM_IMAGE_REPOSITORY_SKIP_TLS_VERIFY: true



TKG_PROXY_CA_CERT: LS0tLS1…

TKG_CUSTOM_IMAGE_REPOSITORY_CA_CERTIFICATE: LS0tLS1…

VSPHERE_CONTROL_PLANE_ENDPOINT: 10.10.10.10

CLUSTER_API_SERVER_PORT: 443

VIP_NETWORK_INTERFACE: eth0

AVI_CONTROL_PLANE_HA_PROVIDER: false

ENABLE_AUDIT_LOGGING:true




VSPHERE_CLONE_MODE: fullClone
VSPHERE_DATACENTER: /dc0
VSPHERE_DATASTORE: ds1
VSPHERE_FOLDER: vm0
VSPHERE_NETWORK: TESTNETWORK
VSPHERE_RESOURCE_POOL: rp0
VSPHERE_SERVER: somehostname
VSPHERE_STORAGE_POLICY_ID:
VSPHERE_TEMPLATE: photon-3-v1.19.3+vmware.1
VSPHERE_TLS_THUMBPRINT: dummythumbprint


CONTROL_PLANE_MACHINE_COUNT: 5






WORKER_MACHINE_COUNT: 3






VSPHERE_SSH_AUTHORIZED_KEY: ssh-rsa A
```

</td>
</tr>
</table>
<!-- markdownlint-restore -->
