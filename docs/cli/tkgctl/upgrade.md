# Upgrades using tkgctl library

Upgrade includes mainly 2 parts. Kubernetes version upgrade and provider upgrade. This documents discusses about both the type of upgrade.

Note: Provider upgrade only applies to management cluster.

# Upgrading kubernetes for TKG cluster

To upgrade kubernetes version user can use below command for workload cluster:
```
tanzu cluster upgrade <cluster-name> --tkr <tkr-name>
```

For management cluster:
```
tanzu management-cluster upgrade // upgrades management cluster and associated management component to latest versions
```

What validation does library performs?
- Makes sure this is not a downgrade and kubernetes minor version is not skipped during upgrade

Upgrade Steps:
- Read TKR BOM file to get the global `imageRepository`, `imageTag` and `imageRepository` information for `CoreDNS` and `Etcd`
- Create InfrastructureMachineTemplates(`VSphereMachineTemplate`, `AWSMachineTemplate`, `AzureMachineTemplate) required for upgrade (See below on how the templates are created)
- Patch KCP object to upgrade control-plane nodes with below configuration
```
 {
	"spec": {
	  "version": "%s",
	  "infrastructureTemplate": {
		"name": "%s",
		"namespace": "%s"
	  },
	  "kubeadmConfigSpec": {
		"clusterConfiguration": {
		  "imageRepository": "%s",
		  "dns": {
			"imageRepository": "%s",
			"imageTag": "%s"
		  },
		  "etcd": {
			"local": {
			  "imageRepository": "%s",
			  "imageTag": "%s"
			}
		  }
		}
	  }
	}
}
```
- Wait for kubernetes version to be updated for the cluster. For more details read [Detecting Upgrade Status](#detecting-upgrade-status)
- Patch MachineDeployment object to upgrade worker nodes with below configuration
```
{
	"spec": {
	  "template": {
		"spec": {
		  "version": "%s",
		  "infrastructureRef": {
			"name": "%s",
			"namespace": "%s"
		  }
		}
	  }
	}
}
```
- Wait for kubernetes version to be updated for all worker nodes. For more details read [Detecting Upgrade Status](#detecting-upgrade-status)


## How InfrastructureMachineTemplates(VSphereMachineTemplate, AWSMachineTemplate) are created for upgrade

To do upgrade one important part is updating the InfrastructureMachineTemplate reference in KubeadmControlPlane and MachineDeployment object.
Note: Updating `VSphereMachineTemplate.Spec.Template.Spec.Template` or `AWSMachineTemplate.Spec.Template.Spec.AMI.ID` or `AzureMachineTemplate.Spec.Template.Spec.image` information in existing template does not help as CAPI uses it as reference and controllers does not get reconciled unless the name of the InfrastructureMachineTemplate is changed under `KCP.Spec.InfrastructureTemplate` and `MD.Spec.Template.Spec.infrastructureRef`

### Steps to create new `InfrastructureMachineTemplate`
- copy the existing `InfrastructureMachineTemplate`
- update the template name (new template names: {CLUSTER_NAME}-{KUBERNETES_VERSION}, replace ./+ with – in kubernetes version)
- update the relavent image information in the new `InfrastructureMachineTemplate`.
- create new `InfrastructureMachineTemplate` by applying the template the the cluster

### For VSphereMachineTemplate
For vSphere only thing that needs to change in the existing VSphereMachineTemplate is to update the VM_TEMPLATE (`VSphereMachineTemplate.Spec.Template.Spec.Template`) to use for newer version of kubernetes.

Steps to get the VM_TEMPLATE:
- Get the `vSphereUserName` and `vSpherePassword` from management cluster's secret `capv-manager-bootstrap-credentials` and use `VSphereCluster` object to get the `VCUrl`. And create VC client.
- To find the VM_TEMPLATE associated with kubernetes version, vSphere inventory is scanned to find the matching OVA Template using comparing `vAppConfig` property of OVA Template.
- Confirm or filter the OVA templates by matching the ova version with `ova` section under TKR BOM file and select the matching OVA template

### For AWSMachineTemplate
For AWS only thing that needs to change in the existing AWSMachineTemplate is to update the AWS_AMI_ID (`AWSMachineTemplate.Spec.Template.Spec.AMI.ID`)

Steps to get the AWS_AMI_ID:
- Get the `AWS_REGION` information from AWSCluster object(`AWSCluster.Spec.Region`)
- Use this region information and map from TKR BOM file to get the correct AWS_AMI_ID


## Detecting Upgrade Status
As right now there is no easy way to tell the upgrade is successful or not from cluster-api, tkg-cli uses following ways to detect the upgrade status and failure.

### Control-plane nodes upgrade status detection
    - Success:
		* Verify KCP with Spec.Replicas == Status.Replicas == Status.UpdatedReplicas == Status.ReadyReplicas
		* Verify kubectl version (server version with discovery client returns updated kubernetes version)
		* If both above conditions are true, it’s a success.

    - Failure:
		* Actual wait-time: 15 minutes
		* Monitor KCP status Status.Replicas, Status.UpdatedReplicas, Status.ReadyReplicas.
		  Monitor All control-plane Machine object’s Phase (running, provisioning)
		  If any update happens in any of the above state, update the wait-time again to 15 minutes.
		* If none of the above information is updated for next 15 minutes. Mark upgrade as failure.

### Worker nodes upgrade status detection
	- Success:
		* Verify MD with Spec.Replicas == Status.Replicas == Status.UpdatedReplicas == Status.ReadyReplicas
		* Verify all worker node Machine objects are upgraded to newer kubernetes version
		  Machine.Spec.Version and in Phase==Running
		* If both above conditions are true, it’s a success.

      Failure:
		* Actual wait-time: 15 minutes
		* Monitor MD status Status.Replicas, Status.UpdatedReplicas, Status.ReadyReplicas.
		  Monitor All worker Machine object’s Phase (running, provisioning)
		* If none of the above information is updated for next 15 minutes. Mark upgrade as failure.


# Upgrading providers for management cluster
Providers would be upgraded as part of management cluster upgrade. During management cluster upgrade process, providers would be upgraded first followed by the kubernetes version upgrade.

Providers Upgrade Steps:
- Read the providers upgrade versions from the TKG BOM file
- List the current providers in Management cluster and for each provider installed in the management cluster, get the latest version from the TKG BoM file
- Generate the ManagementGroup name by parsing the CoreProvider’s InstanceName(eg: capi-system/cluster-api).
- Call the upstream clusterctl  ApplyUpgrade  API to upgrade the providers to the latest version
- Wait for the providers to be up and running

# Upgrading addons for TKG cluster

As TKG v1.3 introduces new components and controllers running on the cluster itself which were not present in TKG v1.2, these controllers get deployed to the clusters during cluster upgrade from v1.2 to v1.3.

Below are the things deployed additionally during cluster upgrade:
- "addons-management/kapp-controller" (management-cluster, workload-cluster)
- "addons-management/tanzu-addons-manager" (management-cluster only)
- "tkr/tkr-controller" (management-cluster only)

Note: Below are the things that DOES NOT get upgraded during TKG cluster upgrade command.
- CNI (antrea, calico)
- vSphere CPI
- vSphere CSI

Follow the instructions in the below section to manually upgrade the addons

## Manually upgrading addons for TKG cluster

After TKG cluster is upgraded from v1.2 to v1.3 using TKG cluster upgrade command the following
steps can be followed to manually upgrade the addons

1. Set the following variables

    ```shell
    # CLUSTER_NAME, CLUSTER_NAMESPACE need to be given as inputs
    CLUSTER_NAME="<CLUSTER_NAME>"
    CLUSTER_NAMESPACE="<CLUSTER_NAMESPACE>"
    CLUSTER_CIDR=$(kubectl get cluster "${CLUSTER_NAME}" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.clusterNetwork.pods.cidrBlocks[0]}')
    SERVICE_CIDR=$(kubectl get cluster "${CLUSTER_NAME}" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.clusterNetwork.services.cidrBlocks[0]}')


    # Applicable only for vsphere infra provider
    # VSPHERE_INSECURE, VSPHERE_USERNAME, VSPHERE_PASSWORD need to be given as inputs
    VSPHERE_SERVER=$(kubectl get VsphereCluster "${CLUSTER_NAME}" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.server}')
    VSPHERE_DATACENTER=$(kubectl get VsphereMachineTemplate  "${CLUSTER_NAME}-control-plane" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.template.spec.datacenter}')
    VSPHERE_RESOURCE_POOL=$(kubectl get VsphereMachineTemplate  "${CLUSTER_NAME}-control-plane" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.template.spec.resourcePool}')
    VSPHERE_DATASTORE=$(kubectl get VsphereMachineTemplate  "${CLUSTER_NAME}-control-plane" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.template.spec.datastore}')
    VSPHERE_FOLDER=$(kubectl get VsphereMachineTemplate  "${CLUSTER_NAME}-control-plane" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.template.spec.folder}')
    VSPHERE_NETWORK=$(kubectl get VsphereMachineTemplate  "${CLUSTER_NAME}-control-plane" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.template.spec.network.devices[0].networkName}')
    VSPHERE_SSH_AUTHORIZED_KEY=$(kubectl get KubeadmControlPlane "${CLUSTER_NAME}-control-plane" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.kubeadmConfigSpec.users[0].sshAuthorizedKeys[0]}')
    VSPHERE_TLS_THUMBPRINT=$(kubectl get VsphereCluster "${CLUSTER_NAME}" -n "${CLUSTER_NAMESPACE}" -o jsonpath='{.spec.thumbprint}')
    VSPHERE_INSECURE=<true/false>
    VSPHERE_USERNAME='<VSPHERE_USERNAME>'
    VSPHERE_PASSWORD='<VSPHERE_PASSWORD>'
    ```
2. Create a config.yaml file with the above configurations

    ```shell
    rm -rf config.yaml
    echo "CLUSTER_CIDR: ${CLUSTER_CIDR}" >> config.yaml
    echo "SERVICE_CIDR: ${SERVICE_CIDR}" >> config.yaml
    echo "NAMESPACE: ${CLUSTER_NAMESPACE}" >> config.yaml

    # Applicable only for vsphere
    echo "VSPHERE_SERVER: ${VSPHERE_SERVER}" >> config.yaml
    echo "VSPHERE_DATACENTER: ${VSPHERE_DATACENTER}" >> config.yaml
    echo "VSPHERE_RESOURCE_POOL: ${VSPHERE_RESOURCE_POOL}" >> config.yaml
    echo "VSPHERE_DATASTORE: ${VSPHERE_DATASTORE}" >> config.yaml
    echo "VSPHERE_FOLDER: ${VSPHERE_FOLDER}" >> config.yaml
    echo "VSPHERE_NETWORK: ${VSPHERE_NETWORK}" >> config.yaml
    echo "VSPHERE_SSH_AUTHORIZED_KEY: ${VSPHERE_SSH_AUTHORIZED_KEY}" >> config.yaml
    echo "VSPHERE_TLS_THUMBPRINT: ${VSPHERE_TLS_THUMBPRINT}" >> config.yaml
    echo "VSPHERE_INSECURE: ${VSPHERE_INSECURE}" >> config.yaml
    echo "VSPHERE_USERNAME: '${VSPHERE_USERNAME}'" >> config.yaml
    echo "VSPHERE_PASSWORD: '${VSPHERE_PASSWORD}'" >> config.yaml
    ```
3. Create a cluster with dry-run option to only get the addons manifest

    ```shell
    export _TKG_CLUSTER_FORCE_ROLE=<management/worklod>
    # Set the addons for which addons manifest needs to be generated. Remove other addons from the the below list.
    export FILTER_BY_ADDON_TYPE="cni/antrea,cloud-provider/vsphere-cpi,csi/vsphere-csi"
    export REMOVE_CRS_FOR_ADDON_TYPE="cni/antrea,cloud-provider/vsphere-cpi,csi/vsphere-csi"

    # Note: vsphere-control-plane-endpoint 10.10.10.10 in the below command is dummy and any IP can be used
    tanzu cluster create ${CLUSTER_NAME} --dry-run --plan dev --vsphere-controlplane-endpoint 10.10.10.10 -f config.yaml > ${CLUSTER_NAME}-addons-manifest.yaml
    ```
    **Addon types**
   ```yaml
    calico: cni/calico
    antrea: cni/antrea
    vsphere CPI: cloud-provider/vsphere-cpi
    vsphere CSI: cloud-provider/vsphere-csi
    metrics server: metrics/metrics-server
    pinniped: authentication/pinniped
    ```

4. Check the generate addons manifest and apply it on cluster

    ```shell
    kubectl apply -f ${CLUSTER_NAME}-addons-manifest.yaml
    ```
