// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	capi "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
)

func DoSetMachineDeploymentCC(clusterClient clusterclient.Client, cluster *capi.Cluster, options *SetMachineDeploymentOptions) error {
	var update *capi.MachineDeploymentTopology
	var base *capi.MachineDeploymentTopology

	if cluster.Spec.Topology.Workers == nil || len(cluster.Spec.Topology.Workers.MachineDeployments) < 1 {
		return errors.New("cluster topology workers are not set. please repair your cluster before trying again")
	}

	for i := range cluster.Spec.Topology.Workers.MachineDeployments {
		if cluster.Spec.Topology.Workers.MachineDeployments[i].Name == options.Name {
			update = &cluster.Spec.Topology.Workers.MachineDeployments[i]
		}
		if cluster.Spec.Topology.Workers.MachineDeployments[i].Name == options.BaseMachineDeployment {
			base = cluster.Spec.Topology.Workers.MachineDeployments[i].DeepCopy()
		}
	}

	if update != nil && base != nil {
		return errors.New("can not specify a base machine deployment when updating a node pool")
	}

	if update != nil {
		base = update
	} else {
		if base == nil {
			if options.NodePool.WorkerClass == "" || options.TKRResolver == "" {
				return errors.New("workerClass and tkrResolver are required for new node pools without a baseMachineDeployment")
			}
			base = &capi.MachineDeploymentTopology{
				Metadata: capi.ObjectMeta{
					Annotations: map[string]string{},
				},
			}
			if options.Replicas == nil {
				base.Replicas = func(i int32) *int32 { return &i }(1) // default new node pool to 1 replica if not specified
			}
		}

		cluster.Spec.Topology.Workers.MachineDeployments = append(cluster.Spec.Topology.Workers.MachineDeployments, *base)
		base = &cluster.Spec.Topology.Workers.MachineDeployments[len(cluster.Spec.Topology.Workers.MachineDeployments)-1]
	}

	base.Name = options.Name

	if options.Replicas != nil {
		base.Replicas = options.Replicas
	}

	if options.WorkerClass != "" {
		base.Class = options.WorkerClass
	}

	if options.TKRResolver != "" {
		base.Metadata.Annotations["run.tanzu.vmware.com/resolve-os-image"] = options.TKRResolver
	}

	if base.Variables == nil {
		base.Variables = &capi.MachineDeploymentVariables{}
	}

	if base.Variables.Overrides == nil {
		base.Variables.Overrides = make([]capi.ClusterVariable, 0, 5)
	}

	if options.Labels != nil {
		nodeLabelsVar := getClusterVariableByName("nodePoolLabels", base.Variables.Overrides)
		if nodeLabelsVar == nil {
			nodeLabelsVar = &capi.ClusterVariable{
				Name:  "nodePoolLabels",
				Value: v1.JSON{},
			}
			base.Variables.Overrides = append(base.Variables.Overrides, *nodeLabelsVar)
			nodeLabelsVar = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		var labels []map[string]string
		// ignore nodePoolLabels not existing
		_ = json.NewDecoder(bytes.NewBuffer(nodeLabelsVar.Value.Raw)).Decode(&labels)

		for k, v := range *options.Labels {
			labels = append(labels, map[string]string{
				"key":    k,
				"values": v,
			})
		}

		output, _ := json.Marshal(labels)
		nodeLabelsVar.Value.Raw = output
	}

	if options.Taints != nil {
		nodeTaintsVar := getClusterVariableByName("nodePoolTaints", base.Variables.Overrides)
		if nodeTaintsVar == nil {
			nodeTaintsVar = &capi.ClusterVariable{
				Name:  "nodePoolTaints",
				Value: v1.JSON{},
			}
			base.Variables.Overrides = append(base.Variables.Overrides, *nodeTaintsVar)
			nodeTaintsVar = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		output, _ := json.Marshal(options.Taints)
		nodeTaintsVar.Value.Raw = output
	}

	if update != nil {
		return clusterClient.UpdateResource(cluster, options.ClusterName, options.Namespace)
	}

	return createNewMachineDeployment(clusterClient, cluster, options, base)
}

func createNewMachineDeployment(clusterClient clusterclient.Client, cluster *capi.Cluster, options *SetMachineDeploymentOptions, base *capi.MachineDeploymentTopology) error {
	if options.NodeMachineType != "" {
		var workerVariable = getClusterVariableByName("worker", base.Variables.Overrides)
		if workerVariable == nil {
			workerVariable = getClusterVariableByName("worker", cluster.Spec.Topology.Variables).DeepCopy()
			base.Variables.Overrides = append(base.Variables.Overrides, *workerVariable)
			workerVariable = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		var worker map[string]interface{}
		if err := json.NewDecoder(bytes.NewBuffer(workerVariable.Value.Raw)).Decode(&worker); err != nil {
			return errors.New("unable to unmarshal worker")
		}

		if _, ok := worker["instanceType"]; ok {
			worker["instanceType"] = options.NodeMachineType
		} else if _, ok := worker["vmSize"]; ok {
			worker["vmSize"] = options.NodeMachineType
		}

		output, _ := json.Marshal(worker)
		workerVariable.Value.Raw = output
	}

	if options.AZ != "" {
		base.FailureDomain = &options.AZ
	}

	if options.VMClass != "" {
		var vmClassVariable = getClusterVariableByName("vmClass", base.Variables.Overrides)
		if vmClassVariable == nil {
			vmClassVariable = getClusterVariableByName("vmClass", cluster.Spec.Topology.Variables).DeepCopy()
			base.Variables.Overrides = append(base.Variables.Overrides, *vmClassVariable)
			vmClassVariable = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		output, _ := json.Marshal(options.VMClass)
		vmClassVariable.Value.Raw = output
	}

	if options.StorageClass != "" {
		var storageClassVariable = getClusterVariableByName("storageClass", base.Variables.Overrides)
		if storageClassVariable == nil {
			storageClassVariable = getClusterVariableByName("storageClass", cluster.Spec.Topology.Variables).DeepCopy()
			base.Variables.Overrides = append(base.Variables.Overrides, *storageClassVariable)
			storageClassVariable = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		output, _ := json.Marshal(options.StorageClass)
		storageClassVariable.Value.Raw = output
	}

	if options.Volumes != nil {
		var volumesVariable = getClusterVariableByName("nodePoolVolumes", base.Variables.Overrides)
		if volumesVariable == nil {
			volumesVariable = getClusterVariableByName("nodePoolVolumes", cluster.Spec.Topology.Variables).DeepCopy()
			base.Variables.Overrides = append(base.Variables.Overrides, *volumesVariable)
			volumesVariable = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		var volumes []map[string]interface{}

		for _, vol := range *options.Volumes {
			volumes = append(volumes, map[string]interface{}{
				"mountPath": vol.MountPath,
				"name":      vol.Name,
				"capacity": map[string]interface{}{
					"storage": vol.Capacity.Storage(),
				},
			})
		}

		output, _ := json.Marshal(volumes)
		volumesVariable.Value.Raw = output
	}

	if err := setVSphereWorkerOptions(options, base, cluster); err != nil {
		return err
	}

	if err := setVSphereVCenterOptions(options, base, cluster); err != nil {
		return err
	}

	return clusterClient.UpdateResource(cluster, options.ClusterName, options.Namespace)
}

func setVSphereVCenterOptions(options *SetMachineDeploymentOptions, base *capi.MachineDeploymentTopology, cluster *capi.Cluster) error {
	if options.VSphere.CloneMode != "" || options.VSphere.Datacenter != "" || options.VSphere.Datastore != "" ||
		options.VSphere.Folder != "" || options.VSphere.Network != "" || options.VSphere.ResourcePool != "" ||
		options.VSphere.StoragePolicyName != "" || options.VSphere.VCIP != "" {
		var vcenterVariable = getClusterVariableByName("vcenter", base.Variables.Overrides)
		if vcenterVariable == nil {
			vcenterVariable = getClusterVariableByName("vcenter", cluster.Spec.Topology.Variables).DeepCopy()
			base.Variables.Overrides = append(base.Variables.Overrides, *vcenterVariable)
			vcenterVariable = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		var vcenter map[string]interface{}
		if err := json.NewDecoder(bytes.NewBuffer(vcenterVariable.Value.Raw)).Decode(&vcenter); err != nil {
			return errors.New("unable to unmarshal vcenter")
		}

		if options.VSphere.CloneMode != "" {
			vcenter["cloneMode"] = options.VSphere.CloneMode
		}

		if options.VSphere.Datacenter != "" {
			vcenter["datacenter"] = options.VSphere.Datacenter
		}

		if options.VSphere.Datastore != "" {
			vcenter["datastore"] = options.VSphere.Datastore
		}

		if options.VSphere.Folder != "" {
			vcenter["folder"] = options.VSphere.Folder
		}

		if options.VSphere.Network != "" {
			vcenter["network"] = options.VSphere.Network
		}

		if options.VSphere.ResourcePool != "" {
			vcenter["resourcePool"] = options.VSphere.ResourcePool
		}

		if options.VSphere.StoragePolicyName != "" {
			vcenter["storagePolicyName"] = options.VSphere.StoragePolicyName
		}

		if options.VSphere.VCIP != "" {
			vcenter["server"] = options.VSphere.VCIP
		}

		output, _ := json.Marshal(vcenter)
		vcenterVariable.Value.Raw = output
	}

	return nil
}

func setVSphereWorkerOptions(options *SetMachineDeploymentOptions, base *capi.MachineDeploymentTopology, cluster *capi.Cluster) error {
	if options.VSphere.NumCPUs > 0 || options.VSphere.DiskGiB > 0 || options.VSphere.MemoryMiB > 0 || len(options.VSphere.Nameservers) > 0 {
		var workerVariable = getClusterVariableByName("worker", base.Variables.Overrides)
		if workerVariable == nil {
			workerVariable = getClusterVariableByName("worker", cluster.Spec.Topology.Variables).DeepCopy()
			base.Variables.Overrides = append(base.Variables.Overrides, *workerVariable)
			workerVariable = &base.Variables.Overrides[len(base.Variables.Overrides)-1]
		}

		var worker map[string]interface{}
		if err := json.NewDecoder(bytes.NewBuffer(workerVariable.Value.Raw)).Decode(&worker); err != nil {
			return errors.New("unable to unmarshal worker")
		}

		if machineInterface, ok := worker["machine"]; ok {
			if machine, ok := machineInterface.(map[string]interface{}); ok {
				if options.VSphere.NumCPUs > 0 {
					machine["numCPUs"] = options.VSphere.NumCPUs
				}
				if options.VSphere.DiskGiB > 0 {
					machine["diskGiB"] = options.VSphere.DiskGiB
				}
				if options.VSphere.MemoryMiB > 0 {
					machine["memoryMiB"] = options.VSphere.MemoryMiB
				}
				worker["machine"] = machine
			}
		}

		if networkInterface, ok := worker["network"]; ok {
			if network, ok := networkInterface.(map[string]interface{}); ok {
				if len(options.VSphere.Nameservers) > 0 {
					network["nameservers"] = options.VSphere.Nameservers
				}
				worker["network"] = network
			}
		}

		output, _ := json.Marshal(worker)
		workerVariable.Value.Raw = output
	}

	return nil
}

func DoDeleteMachineDeploymentCC(clusterClient clusterclient.Client, cluster *capi.Cluster, options *DeleteMachineDeploymentOptions) error {
	var mds []capi.MachineDeploymentTopology
	for _, machineDeployment := range cluster.Spec.Topology.Workers.MachineDeployments {
		if machineDeployment.Name != options.Name {
			mds = append(mds, machineDeployment)
		}
	}

	if len(mds) == 0 {
		return errors.New("unable to delete last node pool")
	}

	if len(mds) == len(cluster.Spec.Topology.Workers.MachineDeployments) {
		return fmt.Errorf("could not find node pool %s to delete", options.Name)
	}

	cluster.Spec.Topology.Workers.MachineDeployments = mds

	if err := clusterClient.UpdateResource(cluster, cluster.ClusterName, cluster.Namespace); err != nil {
		return fmt.Errorf("unable to delete node pools on cluster %s", options.ClusterName)
	}

	return nil
}

func DoGetMachineDeploymentsCC(clusterClient clusterclient.Client, cluster *capi.Cluster, options *GetMachineDeploymentOptions) ([]capi.MachineDeployment, error) {
	workers, err := clusterClient.GetMDObjectForCluster(options.ClusterName, options.Namespace)
	if err != nil {
		return nil, errors.New("error retrieving node pools")
	}

	mds := make([]capi.MachineDeployment, 0, len(workers))
	for i := range workers {
		for _, md := range cluster.Spec.Topology.Workers.MachineDeployments {
			if workers[i].Spec.Template.Labels["topology.cluster.x-k8s.io/deployment-name"] == md.Name {
				workers[i].Name = md.Name
				mds = append(mds, workers[i])
			}
		}
	}

	if options.Name != "" {
		for i := range mds {
			if mds[i].Name == options.Name {
				return []capi.MachineDeployment{
					mds[i],
				}, nil
			}
		}

		return nil, fmt.Errorf("node pool named %s not found", options.Name)
	}

	return mds, nil
}

func getClusterVariableByName(name string, variables []capi.ClusterVariable) *capi.ClusterVariable {
	var variable *capi.ClusterVariable
	for i := range variables {
		if variables[i].Name == name {
			variable = &variables[i]
			break
		}
	}

	return variable
}
