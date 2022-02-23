// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"errors"
	"fmt"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

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
