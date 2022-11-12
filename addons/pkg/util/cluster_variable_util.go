// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

func ParseClusterVariableBool(cluster *clusterapiv1beta1.Cluster, variableName string) (bool, error) {
	var result interface{}
	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return false, err
	}
	return result.(bool), err
}

func ParseClusterVariableString(cluster *clusterapiv1beta1.Cluster, variableName string) (string, error) {
	var result interface{}
	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return "", err
	}
	return result.(string), err
}

func ParseClusterVariableInterface(cluster *clusterapiv1beta1.Cluster, variableName, keyName string) (string, error) {
	var result interface{}

	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return "", err
	}
	interfaceVars := result.(map[string]interface{})

	if valueName, ok := interfaceVars[keyName]; ok {
		if _, ok := valueName.(string); ok {
			return valueName.(string), err
		}
	}
	return "", err
}

func ParseClusterVariableInterfaceArray(cluster *clusterapiv1beta1.Cluster, variableName, keyName string) ([]string, error) {
	var result interface{}

	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return nil, err
	}
	interfaceVars := result.(map[string]interface{})

	if valueName, ok := interfaceVars[keyName]; ok {
		if _, ok := valueName.([]interface{}); ok {
			interfaceArr := valueName.([]interface{})

			aString := make([]string, len(interfaceArr))
			for i, v := range interfaceArr {
				aString[i] = v.(string)
			}
			return aString, err
		} else { //nolint: revive
			return nil, fmt.Errorf("failed to parse the value %v to target type []interface{}", valueName)
		}
	}
	return nil, err
}

func ParseClusterVariableList(cluster *clusterapiv1beta1.Cluster, variableName string) (string, error) {
	var result interface{}

	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return "", err
	}

	rec := result.([]interface{})
	tmpList := make([]string, len(rec))
	for i, v := range rec {
		tmpList[i] = fmt.Sprint(v)
	}
	varList := strings.Join(tmpList, ", ")
	return varList, err
}

func ParseClusterVariableCert(cluster *clusterapiv1beta1.Cluster, variableName, keyName, data string) (string, error) {
	var result interface{}
	var sb strings.Builder

	result, err := parseClusterVariable(cluster, variableName)
	if err != nil || result == nil {
		return "", err
	}
	interfaceVars := result.(map[string]interface{})
	if valueList, ok := interfaceVars[keyName]; ok {
		if rec, ok := valueList.([]interface{}); ok {
			for _, certRaw := range rec {
				if cert, ok := certRaw.(map[string]interface{}); ok {
					certData := cert[data]
					if _, ok := certData.(string); ok {
						sb.WriteString(certData.(string) + "\n")
					}
				}
			}
		}
	}
	return sb.String(), err
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
			case interface{}:
				result = t
			default:
				return nil, fmt.Errorf("invalid type for the cluster variable value for '%s'", variableName)
			}
			break
		}
	}
	return result, nil
}
