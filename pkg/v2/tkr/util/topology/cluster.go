// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package topology

import (
	"encoding/json"

	"github.com/pkg/errors"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// SetVariable sets the cluster variable, to the given value.
// Pre-reqs: cluster.Spec.Topology != nil
func SetVariable(cluster *clusterv1.Cluster, name string, value interface{}) error {
	jsonValue, err := jsonValue(value)
	if err != nil {
		return err
	}

	cVar := clusterVariableForName(cluster, name)
	cVar.Value = *jsonValue
	return nil
}

// GetVariable gets the value of the cluster variable.
func GetVariable(cluster *clusterv1.Cluster, name string, value interface{}) error {
	cVar := clusterVariableForName(cluster, name)
	data, _ := json.Marshal(cVar.Value) // apiextensionsv1.JSON never returns errors when (un)marshaling JSON
	err := json.Unmarshal(data, value)
	return errors.Wrap(err, "unmarshalling from JSON into value")
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

// clusterVariableForName finds or creates in the cluster the *ClusterVariable for the given name.
// Pre-reqs: cluster.Spec.Topology != nil
func clusterVariableForName(cluster *clusterv1.Cluster, name string) *clusterv1.ClusterVariable {
	for i := range cluster.Spec.Topology.Variables {
		v := &cluster.Spec.Topology.Variables[i]
		if v.Name == name {
			return v
		}
	}
	cluster.Spec.Topology.Variables = append(cluster.Spec.Topology.Variables, clusterv1.ClusterVariable{Name: name})
	return &cluster.Spec.Topology.Variables[len(cluster.Spec.Topology.Variables)-1]
}
