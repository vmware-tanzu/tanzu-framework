// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"errors"
	"fmt"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func ParseClusterVariableBool(cluster *clusterapiv1beta1.Cluster, variableName string) (bool, error) {
	var result interface{}
	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return false, err
	}
	return result.(bool), nil
}

func ParseClusterVariableString(cluster *clusterapiv1beta1.Cluster, variableName string) (string, error) {
	var result interface{}
	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return "", err
	}
	return result.(string), nil
}

func parseClusterVariable(cluster *clusterapiv1beta1.Cluster, variableName string) (interface{}, error) {
	var (
		clusterVariableValue interface{}
		result               interface{}
	)

	if cluster == nil {
		return nil, errors.New("cluster resource is nil")
	}
	if cluster.Spec.Topology == nil || variableName == "" {
		return nil, nil
	}
	clusterVariables := cluster.Spec.Topology.Variables
	for _, clusterVariable := range clusterVariables {
		if clusterVariable.Name == variableName {
			if err := json.Unmarshal(clusterVariable.Value.Raw, &clusterVariableValue); err != nil {
				return nil, fmt.Errorf("failed in json unmarshal of cluster variable value for '%s'", variableName)
			}
			switch t := clusterVariableValue.(type) {
			case string:
				result = t
			case bool:
				result = t
			default:
				return nil, fmt.Errorf("invalid type for the cluster variable value for '%s'", variableName)
			}
			break
		}
	}
	return result, nil
}
