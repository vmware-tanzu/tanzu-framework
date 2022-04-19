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

// processCClusterObjectForConfigurationVariables takes ccluster object, process it to capture all configuration variables and add them in environment.
func (t *tkgctl) processCClusterObjectForConfigurationVariables(cclusterObj unstructured.Unstructured) error {
	inputVariablesMap := make(map[string]interface{})
	inputVariablesMap["metadata.name"] = cclusterObj.GetName()
	inputVariablesMap["metadata.namespace"] = cclusterObj.GetNamespace()
	spec := cclusterObj.Object[constants.SPEC].(map[string]interface{})
	err := processYamlObjectAndAddToMap(spec, constants.SPEC, inputVariablesMap)
	if err != nil {
		return err
	}
	providerName, err := validateTopologyClassName(inputVariablesMap[constants.TopologyClass].(string))
	if err != nil {
		return err
	}
	legacyVarMap := make(map[string]string)
	ccToLegacyNamesMap := constants.InfrastructureSpecificVariableMappingMap[providerName]
	// assign cluster class input values to legacy variables
	for inputVariable := range inputVariablesMap {
		if legacyNameForClusterObjectInputVariable, ok := ccToLegacyNamesMap[inputVariable]; ok {
			legacyVarMap[legacyNameForClusterObjectInputVariable] = fmt.Sprintf("%v", inputVariablesMap[inputVariable])
		}
	}
	// override legacy variables with higher precedence values if exists
	if providerName == constants.InfrastructureProviderAWS {
		for higherPrecedenceKey := range constants.CclusterVariablesWithHigherPrecedence {
			legacyName, ok1 := ccToLegacyNamesMap[higherPrecedenceKey]
			value, ok2 := inputVariablesMap[higherPrecedenceKey]
			if ok1 && ok2 {
				legacyVarMap[legacyName] = fmt.Sprintf("%v", value)
			}
		}
	}

	t.TKGConfigReaderWriter().SetMap(legacyVarMap)
	return nil
}

// processYamlObjectAndAddToMap process specific value of the Cluster yaml object, ccMappingName is the path of the value, if input value is simple value then its updated in the variablesMap with ccMappingName as key.
func processYamlObjectAndAddToMap(value interface{}, ccMappingName string, variablesMap map[string]interface{}) error {
	var err error
	switch value := value.(type) {
	case []interface{}:
		err = processYamlObjectArrayInterfaceType(value, ccMappingName, variablesMap)
	case []map[string]interface{}:
		for index := range value {
			err = processYamlObjectAndAddToMap(value[index], ccMappingName, variablesMap)
			if err != nil {
				return err
			}
		}
	case map[string]interface{}:
		for key := range value {
			nextLevelName := key
			nextLevelVal := value[nextLevelName]
			// noProxy has value of type array, so assign it here
			if ccMappingName+"."+nextLevelName == "spec.topology.variables.proxy.noProxy" {
				variablesMap[ccMappingName+"."+nextLevelName] = nextLevelVal
			} else {
				err = processYamlObjectAndAddToMap(nextLevelVal, ccMappingName+"."+nextLevelName, variablesMap)
				if err != nil {
					return err
				}
			}
		}
	case interface{}:
		if _, ok := variablesMap[ccMappingName]; ok {
			log.Warningf("duplicate variable in input cluster class config file, variable path: %v", ccMappingName)
		} else if fmt.Sprintf("%v", value) != "" {
			variablesMap[ccMappingName] = value
		}
	default:
		if value != nil {
			errInfo := fmt.Errorf("un-supported input value type:%v in input cluster class file at:%v", reflect.TypeOf(value), ccMappingName)
			return errInfo
		}
	}
	return err
}

func processYamlObjectArrayInterfaceType(value interface{}, ccMappingName string, variablesMap map[string]interface{}) error {
	var err error
	for index := range value.([]interface{}) {
		if ccMappingName == constants.TopologyVariablesSubnets || ccMappingName == constants.TopologyVariablesNodes || ccMappingName == constants.TopologyWorkersMachineDeployments {
			err = processYamlObjectAndAddToMap(value.([]interface{})[index], ccMappingName+"."+strconv.Itoa(index), variablesMap)
		} else if regExMachineDep.MatchString(ccMappingName) || ccMappingName == constants.TopologyVariables {
			variables := value.([]interface{})
			for varIndex := range variables {
				nameValueMap := variables[varIndex].(map[string]interface{})
				varName := nameValueMap["name"]
				varValue := nameValueMap["value"]
				nextLevelName := ccMappingName + "." + varName.(string)
				// spec.topology.variables.proxy has any value then enable TKG_HTTP_PROXY_ENABLED, spec.topology.variables.proxy mapped to TKG_HTTP_PROXY_ENABLED
				if varName == "proxy" {
					variablesMap[nextLevelName] = true
				}
				err = processYamlObjectAndAddToMap(varValue, nextLevelName, variablesMap)
				if err != nil {
					return err
				}
			}
			break // all variables are processed in above loop so break it.
		} else {
			err = processYamlObjectAndAddToMap(value.([]interface{})[index], ccMappingName, variablesMap)
		}
		if err != nil {
			return err
		}
	}
	return err
}

// validateTopologyClassName takes input cluster class spec.topology.class string and validates, returns provides name
func validateTopologyClassName(topologyClass string) (string, error) {
	var provider string
	if topologyClass == "" {
		return provider, errors.New(constants.TopologyClassIncorrectValueErrMsg)
	}
	topologyClassSplit := strings.Split(topologyClass, "-")
	if len(topologyClassSplit) < 3 {
		return provider, errors.New(constants.TopologyClassIncorrectValueErrMsg)
	} else if _, ok := constants.InfrastructureProviders[topologyClassSplit[1]]; !ok {
		return provider, errors.New(constants.TopologyClassIncorrectValueErrMsg)
	}
	return topologyClassSplit[1], nil
}

// overrideClusterOptionsWithCClusterConfigurationValues overrides CreateClusterOptions attributes with latest values from the environment.
func (t *tkgctl) overrideClusterOptionsWithCClusterConfigurationValues(cc *CreateClusterOptions) {
	cc.Namespace, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
	cc.ClusterName, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
	cc.Plan, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan)
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

// overrideManagementClusterOptionsWithCClusterConfigurationValues overrides InitRegion attributes with latest values from the environment.
func (t *tkgctl) overrideManagementClusterOptionsWithCClusterConfigurationValues(ir *InitRegionOptions) {
	ir.Namespace, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace)
	ir.ClusterName, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName)
	ir.Plan, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan)
	ir.InfrastructureProvider, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider)
	ir.Size, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize)
	ir.ControlPlaneSize, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize)
	ir.WorkerSize, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize)
	ir.CniType, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI)
	ir.VsphereControlPlaneEndpoint, _ = t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint)
}
