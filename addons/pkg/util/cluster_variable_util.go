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
	_, result, err := parseClusterVariable(cluster, variableName)
	return result, err
}

func ParseClusterVariableString(cluster *clusterapiv1beta1.Cluster, variableName string) (string, error) {
	result, _, err := parseClusterVariable(cluster, variableName)
	return result, err
}

func parseClusterVariable(cluster *clusterapiv1beta1.Cluster, variableName string) (string, bool, error) {
	var (
		clusterVariableValue interface{}
		result               string
	)

	if cluster == nil {
		return "", false, errors.New("cluster resource is nil")
	}
	if cluster.Spec.Topology == nil || variableName == "" {
		return "", false, nil
	}
	clusterVariables := cluster.Spec.Topology.Variables
	for _, clusterVariable := range clusterVariables {
		if clusterVariable.Name == variableName {
			if err := json.Unmarshal(clusterVariable.Value.Raw, &clusterVariableValue); err != nil {
				return "", false, fmt.Errorf("failed in json unmarshal of cluster variable value for '%s'", variableName)
			}
			switch t := clusterVariableValue.(type) {
			case string:
				result = t
			case bool:
				return "", t, nil
			default:
				return "", false, fmt.Errorf("invalid type for the cluster variable value for '%s'", variableName)
			}
			break
		}
	}
	return result, false, nil
}

/*
func ParseClusterVariableString(cluster *clusterapiv1beta1.Cluster, variableName string) (string, error) {
	var (
		clusterVariableValue interface{}
		result               string
	)

	if cluster == nil {
		return "", errors.New("cluster resource is nil")
	}
	if cluster.Spec.Topology == nil || variableName == "" {
		return "", nil
	}
	clusterVariables := cluster.Spec.Topology.Variables
	for _, clusterVariable := range clusterVariables {
		if clusterVariable.Name == variableName {
			if err := json.Unmarshal(clusterVariable.Value.Raw, &clusterVariableValue); err != nil {
				return "", fmt.Errorf("failed in json unmarshal of cluster variable value for '%s'", variableName)
			}
			switch t := clusterVariableValue.(type) {
			case string:
				result = t
			default:
				return "", fmt.Errorf("invalid type for the cluster variable value for '%s'", variableName)
			}

			break
		}
	}
	return result, nil
}

*/
