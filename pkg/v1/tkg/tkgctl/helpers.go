// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"fmt"
	"net"
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

	clusterAttributePathToLegacyVarNameMap := constants.InfrastructureSpecificVariableMappingMap[providerName]

	// assign cluster class input values to legacy variables
	legacyVarMap := make(map[string]string)
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

	// Set TKG IP Family based on Cluster and Service CIDRs
	var isIPV6Primary bool
	if legacyVarMap[constants.TKGIPV6Primary] != "" {
		isIPV6Primary, _ = strconv.ParseBool(legacyVarMap[constants.TKGIPV6Primary])
	}
	IPFamily, err := GetIPFamilyForGiveClusterNetworkCIDRs(stringArrayToStringWithCommaSeparatedElements(legacyVarMap[constants.ConfigVariableClusterCIDR]), stringArrayToStringWithCommaSeparatedElements(legacyVarMap[constants.ConfigVariableServiceCIDR]), isIPV6Primary)
	if err != nil {
		return err
	}
	legacyVarMap[constants.TKGIPFamily] = IPFamily

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
			nextLevelFullPath := clusterAttributePath + "." + nextLevelName
			// There are attributes which has value of type array, no need to process those variables values further
			// convert array data into a string (with comma separated elements) and assign the value
			if _, ok := constants.ClusterAttributesWithArrayTypeValue[nextLevelFullPath]; ok && nextLevelVal != nil {
				inputVariablesMap[nextLevelFullPath] = stringArrayToStringWithCommaSeparatedElements(fmt.Sprintf("%v", nextLevelVal))
			} else {
				err = processYamlObjectAndAddToMap(nextLevelVal, nextLevelFullPath, inputVariablesMap)
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
			// if path spec.topology.variables.network.proxy has any child attributes then enable TKG_HTTP_PROXY_ENABLED, spec.topology.variables.network.proxy mapped to TKG_HTTP_PROXY_ENABLED
			if strings.HasPrefix(clusterAttributePath, "spec.topology.variables.network.proxy") {
				inputVariablesMap["spec.topology.variables.network.proxy"] = true
			}
		}
	default:
		if value != nil {
			errInfo := fmt.Errorf("unsupported input value type:%v in input cluster class file attribute at:%v", reflect.TypeOf(value), clusterAttributePath)
			return errInfo
		}
	}
	return err
}

// processYamlObjectArrayInterfaceType process array interface type, does handles some special cases of cluster class attribute paths
func processYamlObjectArrayInterfaceType(value []interface{}, clusterAttributePath string, inputVariablesMap map[string]interface{}) error {
	var err error
	for index := range value {
		if clusterAttributePath == constants.TopologyVariablesNetworkSubnets || clusterAttributePath == constants.TopologyVariablesNodes || clusterAttributePath == constants.TopologyWorkersMachineDeployments {
			err = processYamlObjectAndAddToMap(value[index], clusterAttributePath+"."+strconv.Itoa(index), inputVariablesMap)
		} else if regExMachineDep.MatchString(clusterAttributePath) || clusterAttributePath == constants.TopologyVariables {
			// The attribute paths - spec.topology.workers.machineDeployments.[0-9].variables.overrides, spec.topology.variables
			// has array of data, each array element is of type map, with key(as "name")/value (as "value")
			// to organize better, get "name" and "value", append "name" value to attribute path and "value" value as value, process them as next level.
			variables := value
			for varIndex := range variables {
				nameValueMap := variables[varIndex].(map[string]interface{})
				varName := nameValueMap["name"]
				varValue := nameValueMap["value"]
				nextLevelName := clusterAttributePath + "." + varName.(string)
				// The attribute path "spec.topology.variables.trust.*" has value of type array, each array element is of type map, map has keys - "name" and "data",
				// need to process "spec.topology.variables.trust."+ (value of "name") as attribute path and value of "data" next level value, process it again.
				if varName == "trust" {
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
				} else if varName == "TKR_DATA" { // no need to process TKR_DATA, because there is no mapping as of now
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
// from the environment only if environment has value, the variable values in environment could be updated from input config file or input cluster class file.
func (t *tkgctl) overrideClusterOptionsWithLatestEnvironmentConfigurationValues(cc *CreateClusterOptions) {
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName); err == nil && val != "" {
		cc.ClusterName = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan); err == nil && val != "" {
		cc.Plan = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace); err == nil && val != "" {
		cc.Namespace = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider); err == nil && val != "" {
		cc.InfrastructureProvider = val
	}
	if val, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableControlPlaneMachineCount, t.TKGConfigReaderWriter()); err == nil && val != 0 {
		cc.ControlPlaneMachineCount = val
	}
	if val, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableWorkerMachineCount, t.TKGConfigReaderWriter()); err == nil && val != 0 {
		cc.WorkerMachineCount = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize); err == nil && val != "" {
		cc.Size = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize); err == nil && val != "" {
		cc.ControlPlaneSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize); err == nil && val != "" {
		cc.WorkerSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI); err == nil && val != "" {
		cc.CniType = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableEnableClusterOptions); err == nil && val != "" {
		cc.EnableClusterOptions = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint); err == nil && val != "" {
		cc.VsphereControlPlaneEndpoint = val
	}
}

// overrideManagementClusterOptionsWithLatestEnvironmentConfigurationValues overrides InitRegion attributes with latest values
// from the environment only if environment has value, the variable values in environment could be updated from input config file or input cluster class file.
func (t *tkgctl) overrideManagementClusterOptionsWithLatestEnvironmentConfigurationValues(ir *InitRegionOptions) {
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName); err == nil && val != "" {
		ir.ClusterName = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace); err == nil && val != "" {
		ir.Namespace = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan); err == nil && val != "" {
		ir.Plan = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider); err == nil && val != "" {
		ir.InfrastructureProvider = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize); err == nil && val != "" {
		ir.Size = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize); err == nil && val != "" {
		ir.ControlPlaneSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize); err == nil && val != "" {
		ir.WorkerSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI); err == nil && val != "" {
		ir.CniType = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint); err == nil && val != "" {
		ir.VsphereControlPlaneEndpoint = val
	}
}

// GetIPFamilyForGiveClusterNetworkCIDRs takes clusterNetwork - pods cirds and service cidrs, and returns IP family type
// The input cidrs string can have multiple cidr's separated with comma (,), but it can not have more than two on each input cidr string.
// If input cidr's has both ipv6 and ipv4, and if isIPV6Primary is true then returns "ipv6,ipv4" else if if isIPV6Primary is false then "ipv4,ipv6"
// If all input cidr's has only ipv6 then returns "ipv6"
// If all input cidr's has only ipv4 then returns "ipv4"
func GetIPFamilyForGiveClusterNetworkCIDRs(podsCIDRs string, serviceCIDRs string, isIPV6Primary bool) (string, error) {
	DefaultIPFamily := constants.IPv4Family
	var podsIPFamily, serviceIPFamily string
	var err error
	if podsCIDRs != "" {
		podsIPFamily, err = GetIPFamilyForCIDR(strings.Split(podsCIDRs, ","), isIPV6Primary)
		if err != nil {
			return DefaultIPFamily, err
		}
	} else {
		podsIPFamily = DefaultIPFamily
	}
	if serviceIPFamily != "" {
		serviceIPFamily, err = GetIPFamilyForCIDR(strings.Split(serviceCIDRs, ","), isIPV6Primary)
		if err != nil {
			return DefaultIPFamily, err
		}
	} else {
		serviceIPFamily = DefaultIPFamily
	}

	if podsIPFamily == constants.DualStackPrimaryIPv6Family || serviceIPFamily == constants.DualStackPrimaryIPv6Family {
		return constants.DualStackPrimaryIPv6Family, nil
	}
	if podsIPFamily == constants.DualStackPrimaryIPv4Family || serviceIPFamily == constants.DualStackPrimaryIPv4Family {
		return constants.DualStackPrimaryIPv4Family, nil
	}
	if podsIPFamily == constants.IPv6Family || serviceIPFamily == constants.IPv6Family {
		return constants.IPv6Family, nil
	}
	return constants.IPv4Family, nil
}

// stringArrayToStringWithCommaSeparatedElements converts given string (which has array data in fmt.Println output format) to string which has array elements separated with comma (,)
// eg: Input string is "[100.96.0.0/11 100.64.0.0/18]" converts to "100.64.0.0/18,100.64.0.0/18" of string type.
func stringArrayToStringWithCommaSeparatedElements(arrayDataInStringFormat string) string {
	if arrayDataInStringFormat == "" {
		return ""
	}
	return strings.Join(strings.Split(strings.Trim(arrayDataInStringFormat, "[]"), " "), ",")
}

// IPFamilyForCIDRStrings takes cidr array and returns ip family type (ipv4, ipv6 or dual)
// The input cidrs array lenth max 2 only
// If input cidr's has both ipv6 and ipv4, and if isIPV6Primary is true then returns "ipv6,ipv4" else if if isIPV6Primary is false then "ipv4,ipv6"
// If all input cidr's has only ipv6 then returns "ipv6"
// If all input cidr's has only ipv4 then returns "ipv4"
func GetIPFamilyForCIDR(cidrs []string, isIPV6Primary bool) (string, error) {
	DefaultIPFamily := constants.IPv4Family
	if len(cidrs) > 2 {
		return DefaultIPFamily, fmt.Errorf("too many CIDRs specified: %v", cidrs)
	}
	var foundIPv4 bool
	var foundIPv6 bool
	for _, cidr := range cidrs {
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return DefaultIPFamily, fmt.Errorf("could not parse CIDR %v, error: %s", cidr, err)
		}
		if ip.To4() != nil {
			foundIPv4 = true
		} else {
			foundIPv6 = true
		}
	}
	switch {
	case foundIPv4 && foundIPv6:
		if isIPV6Primary {
			return constants.DualStackPrimaryIPv6Family, nil
		}
		return constants.DualStackPrimaryIPv4Family, nil
	case foundIPv4:
		return constants.IPv4Family, nil
	case foundIPv6:
		return constants.IPv6Family, nil
	}
	return DefaultIPFamily, nil
}
