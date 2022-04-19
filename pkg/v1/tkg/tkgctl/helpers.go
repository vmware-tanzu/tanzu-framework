// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	utilyaml "sigs.k8s.io/cluster-api/util/yaml"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/vc"
)

var okResponsesMap = map[string]struct{}{
	"y": {},
	"Y": {},
}
var regExMachineDep = regexp.MustCompile(constants.RegexpMachineDeploymentsOverrides)

func askForConfirmation(message string) error {
	var response string
	msg := message + " [y/N]: "
	log.ForceWriteToStdErr([]byte(msg))
	_, err := fmt.Scanln(&response)
	if err != nil {
		return errors.New("aborted")
	}
	if _, exit := okResponsesMap[response]; !exit {
		return errors.New("aborted")
	}
	return nil
}

// verifyThumbprint verifies the vSphere thumbprint used for deploying the management cluster
func (t *tkgctl) verifyThumbprint(skipPrompt bool) error {
	if insecure, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereInsecure); err == nil && insecure == True {
		t.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereTLSThumbprint, "")
		return nil
	}

	vcHost, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereServer)
	if err != nil {
		return errors.Errorf("failed to get %s", constants.ConfigVariableVsphereServer)
	}

	thumbprint, err := vc.GetVCThumbprint(vcHost)
	if err != nil {
		return errors.Wrap(err, "cannot verify the thumbprint")
	}

	if configThumbprint, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereTLSThumbprint); err == nil {
		if configThumbprint == thumbprint {
			return nil
		}
		log.Warningf("The %s variable does not match the thumbprint of vSphere %s, the actual thumbprint is %s", constants.ConfigVariableVsphereTLSThumbprint, vcHost, thumbprint)
	}

	if !skipPrompt {
		err = askForConfirmation(fmt.Sprintf("Do you want to continue with the vSphere thumbprint %s", thumbprint))
		if err != nil {
			return err
		}
	}

	t.TKGConfigReaderWriter().Set(constants.ConfigVariableVsphereTLSThumbprint, thumbprint)
	return nil
}

func (t *tkgctl) ensureClusterConfigFile(clusterConfigFile string) (string, error) {
	var err error
	if clusterConfigFile == "" {
		// if clusterConfigFile is not provided use default cluster config file
		clusterConfigFile, err = tkgconfigpaths.New(t.configDir).GetDefaultClusterConfigPath()
		if err != nil {
			return "", errors.Wrap(err, "unable to get default cluster config file path")
		}
		log.V(3).Infof("cluster config file not provided using default config file at '%v'", clusterConfigFile)
	}

	// create empty clusterConfigFile if not present
	if _, err = os.Stat(clusterConfigFile); os.IsNotExist(err) {
		log.V(3).Infof("cluster config file does not exist. Creating new one at '%v'", clusterConfigFile)
		err = os.WriteFile(clusterConfigFile, []byte{}, constants.ConfigFilePermissions)
		if err != nil {
			return "", errors.Wrap(err, "cannot initialize cluster config file")
		}
	}

	// read cluster config file with tkgConfigReaderWriter and merge it into existing configuration
	if err := t.TKGConfigReaderWriter().MergeInConfig(clusterConfigFile); err != nil {
		return "", errors.Wrap(err, "error initializing cluster config")
	}

	// decode credentials in viper after reading config file
	err = t.tkgConfigUpdaterClient.DecodeCredentialsInViper()
	if err != nil {
		return "", errors.Wrap(err, "unable to update encoded credentials")
	}
	return clusterConfigFile, nil
}

// ensureLogDirectory returns the directory path where log files should be stored by default.
func (t *tkgctl) ensureLogDirectory() (string, error) {
	logDir, err := tkgconfigpaths.New(t.configDir).GetLogDirectory()
	if err != nil {
		return "", err
	}

	// We got the path it should be, make sure it exists
	if _, err = os.Stat(logDir); os.IsNotExist(err) {
		log.V(3).Infof("cluster log directory does not exist. Creating new one at %q", logDir)
		err := os.MkdirAll(logDir, os.ModePerm)
		if err != nil {
			return "", errors.Wrap(err, "cannot initialize cluster log directory")
		}
	}

	return logDir, nil
}

// getAuditLogPath gets the full path to where an audit log should be located
// for a give clusterName.
func (t *tkgctl) getAuditLogPath(clusterName string) (string, error) {
	if clusterName == "" {
		return "", errors.New("cluster name is required to determine audit log path")
	}

	path, err := t.ensureLogDirectory()
	if err != nil {
		return "", fmt.Errorf("unable to determine audit log path: %s", err.Error())
	}

	return filepath.Join(path, fmt.Sprintf("%s.log", clusterName)), nil
}

// removeAuditLog will remove a cluster's audit log from the local filesystem.
// This is done on a best effort basis, so if the file does not exist, or if
// there is an issue deleting the file there will be no indication.
func (t *tkgctl) removeAuditLog(clusterName string) {
	path, err := t.getAuditLogPath(clusterName)
	if err != nil {
		// We delete on a best-effort basis, so just return
		return
	}

	_ = os.Remove(path)
}

// checkIfInputFileIsClusterClassBased checks user input file, if it has Cluster object then
// reads all non-empty variables in cluster.spec.topology.variables, and updates those variables in
// environment and also CreateClusterOptions.
func (t *tkgctl) checkIfInputFileIsClusterClassBased(clusterConfigFile string) (bool, unstructured.Unstructured, error) {
	var clusterobj unstructured.Unstructured

	isInputFileClusterClassBased := false
	if clusterConfigFile == "" {
		return isInputFileClusterClassBased, clusterobj, nil
	}
	content, err := os.ReadFile(clusterConfigFile)
	if err != nil {
		return isInputFileClusterClassBased, clusterobj, errors.Wrap(err, fmt.Sprintf("Unable to read input file: %v ", clusterConfigFile))
	}
	yamlObjects, err := utilyaml.ToUnstructured(content)
	if err != nil {
		return isInputFileClusterClassBased, clusterobj, errors.Wrap(err, fmt.Sprintf("Input file content is not yaml formatted, file path: %v", clusterConfigFile))
	}

	for i := range yamlObjects {
		obj := yamlObjects[i]
		if obj.GetKind() == constants.KindCluster {
			isInputFileClusterClassBased = true
			clusterobj = obj
			break
		}
	}
	return isInputFileClusterClassBased, clusterobj, nil
}

// processClusterObjectForConfigurationVariables takes cluster object, process it to capture all configuration variables and add them in environment.
func (t *tkgctl) processClusterObjectForConfigurationVariables(clusterObj unstructured.Unstructured) error {
	inputVariablesMap := make(map[string]interface{})
	inputVariablesMap["metadata.name"] = clusterObj.GetName()
	inputVariablesMap["metadata.namespace"] = clusterObj.GetNamespace()
	spec := clusterObj.Object[constants.SPEC].(map[string]interface{})
	err := processYamlObjectAndAddToMap(spec, constants.SPEC, inputVariablesMap)
	if err != nil {
		return err
	}
	providerName, err := getProviderNameFromTopologyClassName(inputVariablesMap[constants.TopologyClass])
	if err != nil {
		return err
	}
	legacyVarMap := make(map[string]string)
	clusterAttributePathToLegacyVarNameMap := constants.InfrastructureSpecificVariableMappingMap[providerName]

	// assign cluster class input values to legacy variables
	for inputVariable := range inputVariablesMap {
		if legacyNameForClusterObjectInputVariable, ok := clusterAttributePathToLegacyVarNameMap[inputVariable]; ok {
			legacyVarMap[legacyNameForClusterObjectInputVariable] = fmt.Sprintf("%v", inputVariablesMap[inputVariable])
		}
	}

	// Some properties (NODE_MACHINE_TYPE_1, NODE_MACHINE_TYPE_2, etc)
	// can have two values from Cluster object attributes, key and value of constants.ClusterAttributesHigherPrecedenceToLowerMap are two attribute paths points to same legacy variable.
	// eg: key - "spec.topology.workers.machineDeployments.1.variables.overrides.NODE_MACHINE_TYPE", value - "spec.topology.variables.nodes.1.machineType"
	// these two key/value attribute paths mapped to NODE_MACHINE_TYPE_1 legacy variable, if these two attribute paths has values in Cluster Object
	// then need to consider higher precedence attribute path value which is key of constants.ClusterAttributesHigherPrecedenceToLowerMap.
	for higherPrecedenceKey := range constants.ClusterAttributesHigherPrecedenceToLowerMap {
		legacyName, ok1 := clusterAttributePathToLegacyVarNameMap[higherPrecedenceKey]
		value, ok2 := inputVariablesMap[higherPrecedenceKey]
		if ok1 && ok2 {
			legacyVarMap[legacyName] = fmt.Sprintf("%v", value)
		}
	}

	// update legacyVarMap in environment
	t.TKGConfigReaderWriter().SetMap(legacyVarMap)
	return nil
}

// processYamlObjectAndAddToMap process specific value of the Cluster yaml object
// value is the value for the given attribute path clusterAttributePath in Cluster object
// if input value is final child value then value is added to variablesMap with clusterAttributePath as key,
// if the input value is not final child value, then its process again for next level child.
func processYamlObjectAndAddToMap(value interface{}, clusterAttributePath string, inputVariablesMap map[string]interface{}) error {
	var err error
	switch value := value.(type) {
	case []interface{}:
		err = processYamlObjectArrayInterfaceType(value, clusterAttributePath, inputVariablesMap)
	case []map[string]interface{}:
		for index := range value {
			err = processYamlObjectAndAddToMap(value[index], clusterAttributePath, inputVariablesMap)
			if err != nil {
				return err
			}
		}
	case map[string]interface{}:
		for key := range value {
			nextLevelName := key
			nextLevelVal := value[nextLevelName]
			// noProxy has value of type array, no need process array values, just assign array value.
			if clusterAttributePath+"."+nextLevelName == "spec.topology.variables.network.proxy.noProxy" {
				inputVariablesMap[clusterAttributePath+"."+nextLevelName] = nextLevelVal
			} else {
				err = processYamlObjectAndAddToMap(nextLevelVal, clusterAttributePath+"."+nextLevelName, inputVariablesMap)
				if err != nil {
					return err
				}
			}
		}
	case interface{}:
		if _, ok := inputVariablesMap[clusterAttributePath]; ok {
			log.Warningf("duplicate variable in input cluster class config file, variable path: %v", clusterAttributePath)
		} else if fmt.Sprintf("%v", value) != "" {
			inputVariablesMap[clusterAttributePath] = value
		}
	default:
		if value != nil {
			errInfo := fmt.Errorf("unsupported input value type:%v in input cluster class file attribute at:%v", reflect.TypeOf(value), clusterAttributePath)
			return errInfo
		}
	}
	return err
}

func processYamlObjectArrayInterfaceType(value []interface{}, clusterAttributePath string, inputVariablesMap map[string]interface{}) error {
	var err error
	for index := range value {
		if clusterAttributePath == constants.TopologyVariablesNetworkSubnets || clusterAttributePath == constants.TopologyVariablesNodes || clusterAttributePath == constants.TopologyWorkersMachineDeployments {
			err = processYamlObjectAndAddToMap(value[index], clusterAttributePath+"."+strconv.Itoa(index), inputVariablesMap)
		} else if regExMachineDep.MatchString(clusterAttributePath) || clusterAttributePath == constants.TopologyVariables {
			// attribute path is spec.topology.workers.machineDeployments.[0-9].variables.overrides OR spec.topology.variables
			// which has key(as name)/value (as value), process them as next level
			variables := value
			for varIndex := range variables {
				nameValueMap := variables[varIndex].(map[string]interface{})
				varName := nameValueMap["name"]
				varValue := nameValueMap["value"]
				nextLevelName := clusterAttributePath + "." + varName.(string)
				// spec.topology.variables.proxy has any value then enable TKG_HTTP_PROXY_ENABLED, spec.topology.variables.proxy mapped to TKG_HTTP_PROXY_ENABLED
				if varName == "proxy" {
					inputVariablesMap[nextLevelName] = true
				} else if varName == "trust" {
					trustValArr := varValue.([]interface{})
					for trustIndex := range trustValArr {
						trustName := trustValArr[trustIndex].(map[string]interface{})["name"].(string)
						trustData := trustValArr[trustIndex].(map[string]interface{})["data"]
						err = processYamlObjectAndAddToMap(trustData, nextLevelName+"."+trustName, inputVariablesMap)
						if err != nil {
							return err
						}
					}
					continue // we are done processing "trust" variable, so process next variable
				} else if varName == "TKR_DATA" { // no need to process TKR_DATA, because there is no mapping
					continue
				}
				err = processYamlObjectAndAddToMap(varValue, nextLevelName, inputVariablesMap)
				if err != nil {
					return err
				}
			}
			break // all variables are processed in above loop so break it.
		} else {
			// process specific index value, this index value could be of any type
			err = processYamlObjectAndAddToMap(value[index], clusterAttributePath, inputVariablesMap)
		}
		if err != nil {
			return err
		}
	}
	return err
}

// getProviderNameFromTopologyClassName takes input cluster class spec.topology.class value and validates, returns provides name.
// The Cluster spec.topology.class attribute value should be valid kubernetes object name : it should be string,
// should be splittable with "-", after split, it should have at least 3 parts, second part should be infra provider name - [aws, azure, vsphere, docker]
// eg: tkg-aws-default : infra provider is aws.
// tkg-unknow-cluster1 : this is not valid name,  "unknow" is not valid infra provider name.

func getProviderNameFromTopologyClassName(topologyClassValue interface{}) (string, error) {
	var provider string

	if topologyClassValue == nil || reflect.ValueOf(topologyClassValue).Kind().String() != reflect.String.String() ||
		topologyClassValue.(string) == "" || len(strings.Split(topologyClassValue.(string), "-")) < 3 {
		return provider, errors.New(constants.TopologyClassIncorrectValueErrMsg)
	}
	topologyClassSplit := strings.Split(topologyClassValue.(string), "-")
	if _, ok := constants.InfrastructureProviders[topologyClassSplit[1]]; !ok {
		return provider, errors.New(constants.TopologyClassIncorrectValueErrMsg)
	}
	return topologyClassSplit[1], nil
}

// overrideClusterOptionsWithLatestEnvironmentConfigurationValues overrides CreateClusterOptions attributes with latest values
// from the environment, which could be updated from input config or input cluster class file.
func (t *tkgctl) overrideClusterOptionsWithLatestEnvironmentConfigurationValues(cc *CreateClusterOptions) {
	cc.ClusterName, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
	cc.Plan, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan)
	cc.Namespace, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
	cc.InfrastructureProvider, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider)
	cc.ControlPlaneMachineCount, _ = tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableControlPlaneMachineCount, t.TKGConfigReaderWriter())
	cc.WorkerMachineCount, _ = tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableWorkerMachineCount, t.TKGConfigReaderWriter())
	cc.Size, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize)
	cc.ControlPlaneSize, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize)
	cc.WorkerSize, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize)
	cc.CniType, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)
	cc.EnableClusterOptions, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableEnableClusterOptions)
	cc.VsphereControlPlaneEndpoint, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint)
}

// overrideManagementClusterOptionsWithLatestEnvironmentConfigurationValues overrides InitRegion attributes with latest values
// from the environment, which could be updated from input config or cluster class file.
func (t *tkgctl) overrideManagementClusterOptionsWithLatestEnvironmentConfigurationValues(ir *InitRegionOptions) {
	ir.ClusterName, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
	ir.Namespace, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
	ir.Plan, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan)
	ir.InfrastructureProvider, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider)
	ir.Size, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize)
	ir.ControlPlaneSize, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize)
	ir.WorkerSize, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize)
	ir.CniType, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)
	ir.VsphereControlPlaneEndpoint, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint)
}
