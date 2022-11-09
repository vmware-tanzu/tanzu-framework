// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"encoding/json"
	"reflect"

	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// SetVariable sets the cluster variable, to the given value.
// Pre-reqs: cluster.Spec.Topology != nil
func SetVariable(cluster *clusterv1.Cluster, name string, value interface{}) error {
	jsonValue, err := jsonValue(value)
	if err != nil {
		return err
	}

	cVar := ensureClusterVariableForName(cluster, name)
	cVar.Value = *jsonValue
	return nil
}

// GetVariable gets the value of the cluster variable.
// Pre-reqs: cluster.Spec.Topology != nil
func GetVariable(cluster *clusterv1.Cluster, name string, value interface{}) error {
	var jsonValue apiextensionsv1.JSON
	if cVar := clusterVariableForName(cluster, name); cVar != nil {
		jsonValue = cVar.Value
	}
	data, _ := json.Marshal(jsonValue) // apiextensionsv1.JSON never returns errors when (un)marshaling JSON
	err := json.Unmarshal(data, value)
	return errors.Wrap(err, "unmarshalling from JSON into value")
}

// SetMDVariable sets the variable for the given machineDeployment, overriding it if necessary, to the given value.
// Pre-reqs: cluster.Spec.Topology != nil && cluster.Spec.Topology.Workers != nil
func SetMDVariable(cluster *clusterv1.Cluster, mdIndex int, name string, value interface{}) error {
	jsonValue, err := jsonValue(value)
	if err != nil {
		return err
	}

	md := &cluster.Spec.Topology.Workers.MachineDeployments[mdIndex]

	cVar := ensureClusterVariableForName(cluster, name)
	if reflect.DeepEqual(*jsonValue, cVar.Value) { // if cluster variable with the same name already has the value being set
		removeMDVariableForName(md, name) // make sure the machineDeployment override is not set
		return nil
	}

	mdVar := ensureMDVariableForName(md, name)
	mdVar.Value = *jsonValue
	return nil
}

// GetMDVariable gets the value of the variable, possibly overridden for the given machineDeployment.
// Pre-reqs: cluster.Spec.Topology != nil && cluster.Spec.Topology.Workers != nil
func GetMDVariable(cluster *clusterv1.Cluster, mdIndex int, name string, value interface{}) error {
	var jsonValue apiextensionsv1.JSON

	md := &cluster.Spec.Topology.Workers.MachineDeployments[mdIndex]

	if mdVar := mdVariableForName(md, name); mdVar != nil {
		jsonValue = mdVar.Value
	} else if cVar := clusterVariableForName(cluster, name); cVar != nil {
		jsonValue = cVar.Value
	}

	data, _ := json.Marshal(jsonValue) // apiextensionsv1.JSON never returns errors when (un)marshaling JSON
	err := json.Unmarshal(data, value)
	return errors.Wrap(err, "unmarshalling from JSON into value")
}

// IsSingleNodeCluster checks if the cluster topology is single node cluster(with CP count as 1 and worker count as 0).
// Pre-reqs: cluster.Spec.Topology != nil
func IsSingleNodeCluster(cluster *clusterv1.Cluster) bool {
	return *cluster.Spec.Topology.ControlPlane.Replicas == *pointer.Int32(1) && HasWorkerNodes(cluster)
}

// HasWorkerNodes checks if the cluster topology has worker nodes.
// Pre-reqs: cluster.Spec.Topology != nil
func HasWorkerNodes(cluster *clusterv1.Cluster) bool {
	return cluster.Spec.Topology.Workers == nil || len(cluster.Spec.Topology.Workers.MachineDeployments) == 0
}

func jsonValue(value interface{}) (*apiextensionsv1.JSON, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling value to JSON")
	}
	result := &apiextensionsv1.JSON{}
	_ = json.Unmarshal(data, result) // apiextensionsv1.JSON never returns errors when (un)marshaling JSON
	return result, nil
}

// clusterVariableForName finds the *ClusterVariable for the given name.
// Pre-reqs: cluster.Spec.Topology != nil
func clusterVariableForName(cluster *clusterv1.Cluster, name string) *clusterv1.ClusterVariable {
	for i := range cluster.Spec.Topology.Variables {
		v := &cluster.Spec.Topology.Variables[i]
		if v.Name == name {
			return v
		}
	}
	return nil
}

// ensureClusterVariableForName finds or creates in the cluster the *ClusterVariable for the given name.
// Pre-reqs: cluster.Spec.Topology != nil
func ensureClusterVariableForName(cluster *clusterv1.Cluster, name string) *clusterv1.ClusterVariable {
	for i := range cluster.Spec.Topology.Variables {
		v := &cluster.Spec.Topology.Variables[i]
		if v.Name == name {
			return v
		}
	}
	cluster.Spec.Topology.Variables = append(cluster.Spec.Topology.Variables, clusterv1.ClusterVariable{Name: name})
	return &cluster.Spec.Topology.Variables[len(cluster.Spec.Topology.Variables)-1]
}

// mdVariableForName finds in the machineDeployment the *ClusterVariable for the given name.
// Pre-reqs: cluster.Spec.Topology != nil && cluster.Spec.Topology.Workers != nil
func mdVariableForName(md *clusterv1.MachineDeploymentTopology, name string) *clusterv1.ClusterVariable {
	if md.Variables == nil {
		return nil
	}
	for i := range md.Variables.Overrides {
		v := &md.Variables.Overrides[i]
		if v.Name == name {
			return v
		}
	}
	return nil
}

// ensureMDVariableForName finds or creates in the machineDeployment the *ClusterVariable for the given name.
func ensureMDVariableForName(md *clusterv1.MachineDeploymentTopology, name string) *clusterv1.ClusterVariable {
	if md.Variables == nil {
		md.Variables = &clusterv1.MachineDeploymentVariables{}
	}
	for i := range md.Variables.Overrides {
		v := &md.Variables.Overrides[i]
		if v.Name == name {
			return v
		}
	}
	md.Variables.Overrides = append(md.Variables.Overrides, clusterv1.ClusterVariable{Name: name})
	return &md.Variables.Overrides[len(md.Variables.Overrides)-1]
}

// removeMDVariableForName finds or creates in the cluster the *ClusterVariable for the given name.
func removeMDVariableForName(md *clusterv1.MachineDeploymentTopology, name string) {
	if md.Variables == nil {
		return
	}
	for i := range md.Variables.Overrides {
		v := &md.Variables.Overrides[i]
		if v.Name == name {
			md.Variables.Overrides = append(md.Variables.Overrides[:i], md.Variables.Overrides[i+1:]...)
			return
		}
	}
}
