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

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfighelper"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigpaths"
	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
)

var okResponsesMap = map[string]struct{}{
	"y": {},
	"Y": {},
}
var errMessageIPv6EnabledCIDRHasNoIPv6 = "the isIPV6Primary: true, but the first value in CIDRs:\"%v\" is not ipv6"
var errMessageCIDRsIPFamiliesNotSame = `the IP Families of input CIDRs: "%v":"%v" , "%v":"%v", both are not same IP Families`

var regExMachineDep = regexp.MustCompile(constants.RegexpMachineDeploymentsOverrides)
var regExpTopologyClassVal = regexp.MustCompile(constants.RegexpTopologyClassValue)

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

// CheckIfInputFileIsClusterClassBased checks user input file, if it has Cluster object then
// reads all non-empty variables in cluster.spec.topology.variables, and updates those variables in
// environment and also CreateClusterOptions.
// TODO (chandrareddyp) : need to make sure the legacy validation error and log messages should not have legacy variable names in case of cluster class use case, should refer Cluster Object attributes (https://github.com/vmware-tanzu/tanzu-framework/issues/2443)
func CheckIfInputFileIsClusterClassBased(clusterConfigFile string) (bool, unstructured.Unstructured, error) {
	var clusterObj unstructured.Unstructured

	isInputFileClusterClassBased := false
	if clusterConfigFile == "" {
		return isInputFileClusterClassBased, clusterObj, nil
	}
	content, err := os.ReadFile(clusterConfigFile)
	if err != nil {
		return isInputFileClusterClassBased, clusterObj, errors.Wrap(err, fmt.Sprintf("Unable to read input file: %v ", clusterConfigFile))
	}

	yamlObjects, err := utilyaml.ToUnstructured(content)
	if err != nil {
		return isInputFileClusterClassBased, clusterObj, errors.Wrap(err, fmt.Sprintf("Input file content is not yaml formatted, file path: %v", clusterConfigFile))
	}

	for i := range yamlObjects {
		obj := yamlObjects[i]
		if obj.GetKind() == constants.KindCluster {
			clusterObj = obj
			class, exists, _ := unstructured.NestedString((&clusterObj).UnstructuredContent(), "spec", "topology", "class")
			if exists && class != "" {
				isInputFileClusterClassBased = true
			} else {
				return isInputFileClusterClassBased, clusterObj, errors.New(constants.ClusterResourceWithoutTopologyNotSupportedErrMsg)
			}
			break
		}
	}
	return isInputFileClusterClassBased, clusterObj, nil
}

func getFieldfromUnstructuredObject(obj unstructured.Unstructured, fields ...string) (string, error) {
	path := strings.Join(fields, ".")
	value, exists, err := unstructured.NestedString(obj.UnstructuredContent(), fields...)
	if err != nil {
		return "", errors.Wrap(err, fmt.Sprintf("Failed to parse %s in %s %s/%s", path, obj.GetKind(), obj.GetNamespace(), obj.GetName()))
	}
	if !exists {
		return "", errors.New(fmt.Sprintf("%s not found in %s %s/%s", path, obj.GetKind(), obj.GetNamespace(), obj.GetName()))
	}
	return value, nil
}

func setVSphereCredentialFromInputfile(legacyVarMap *map[string]string, clusterConfigFile, clusterName, namespace string) error {
	content, err := os.ReadFile(clusterConfigFile)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Unable to read input file: %s", clusterConfigFile))
	}
	yamlObjects, err := utilyaml.ToUnstructured(content)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("Input file content is not yaml formatted, file path: %s", clusterConfigFile))
	}
	found := false
	for _, obj := range yamlObjects {
		if obj.GetKind() == "Secret" && obj.GetName() == clusterName && obj.GetNamespace() == namespace {
			if username, err := getFieldfromUnstructuredObject(obj, "stringData", "username"); err != nil {
				return err
			} else {
				(*legacyVarMap)[constants.ConfigVariableVsphereUsername] = username
			}
			if password, err := getFieldfromUnstructuredObject(obj, "stringData", "password"); err != nil {
				return err
			} else {
				(*legacyVarMap)[constants.ConfigVariableVspherePassword] = password
			}
			found = true
		}
	}
	if !found {
		return errors.New(fmt.Sprintf("Secret %s/%s not found in %s", namespace, clusterName, clusterConfigFile))
	}
	return nil
}

// processClusterObjectForConfigurationVariables takes cluster object, process it to capture all configuration variables and add them in environment.
// TODO (chandrareddyp): validate the cluster class inputs without mapping to legacy variables and deprecate legacy validation (https://github.com/vmware-tanzu/tanzu-framework/issues/2432)
func (t *tkgctl) processClusterObjectForConfigurationVariables(clusterObj unstructured.Unstructured, clusterConfigFile string) error {
	inputVariablesMap := make(map[string]interface{})
	inputVariablesMap["metadata.name"] = clusterObj.GetName()
	inputVariablesMap["metadata.namespace"] = clusterObj.GetNamespace()
	spec := clusterObj.Object[constants.SPEC].(map[string]interface{})
	err := processYamlObjectAndAddToMap(spec, constants.SPEC, inputVariablesMap)
	if err != nil {
		return err
	}

	// TODO (chandrareddyp): Validate the cluster class and identify the provider type  by querying the cluster class object instead of depending on the cluster class name (https://github.com/vmware-tanzu/tanzu-framework/issues/2424)
	providerName, err := getProviderNameFromTopologyClassName(inputVariablesMap[constants.TopologyClass])
	// TODO (chandrareddyp) : Allow user to create workload cluster by disabling validation when provider type is unknown (https://github.com/vmware-tanzu/tanzu-framework/issues/2433)
	if err != nil {
		return err
	}

	// get infra specific variable map
	clusterAttributePathToLegacyVarNameMap := constants.InfrastructureSpecificVariableMappingMap[providerName]

	// assign cluster class input values to legacy variables
	legacyVarMap := make(map[string]string)
	for inputVariable := range inputVariablesMap {
		if legacyNameForClusterObjectInputVariable, ok := clusterAttributePathToLegacyVarNameMap[inputVariable]; ok {
			legacyVarMap[legacyNameForClusterObjectInputVariable] = fmt.Sprintf("%v", inputVariablesMap[inputVariable])
		}
	}
	legacyVarMap[constants.ConfigVariableInfraProvider] = providerName

	// VSPHERE_USERNAME and VSPHERE_PASSWORD do not have mapping to any Cluster variables and should be retrieved from VSphereCluster's identityRef
	if providerName == constants.InfrastructureProviderVSphere {
		if err := setVSphereCredentialFromInputfile(&legacyVarMap, clusterConfigFile, clusterObj.GetName(), clusterObj.GetNamespace()); err != nil {
			return err
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

	// Identify TKG_IP_FAMILY based on Cluster and Service CIDRs, we need to set TKG_IP_FAMILY value, because we need this value for legacy validation in cluster creation flow. Cluster CIDR and Service CIDR IP Family's should be same, either "ipv4", "ip6", or dual stack ("ipv4,ipv6" or "ipv6,ipv4")
	// If Cluster Attribute isIPV6Primary is true, then the first value of cluster/service CIDR should be ipv6 otherwise throws error.
	// If isIPV6Primary is true, and cluster/service CIDR is dual stack, then the IP Family value should be "ipv6,ipv4" (ipv6 should come first), if isIPV6Primary is not true and  cluster/service CIDR is dual stack then IP Family value should be "ipv4,ipv6"
	var isIPV6Primary bool
	if legacyVarMap[constants.TKGIPV6Primary] != "" {
		isIPV6Primary, _ = strconv.ParseBool(legacyVarMap[constants.TKGIPV6Primary])
	}
	clusterCIDRs := stringArrayToStringWithCommaSeparatedElements(legacyVarMap[constants.ConfigVariableClusterCIDR])
	serviceCIDRs := stringArrayToStringWithCommaSeparatedElements(legacyVarMap[constants.ConfigVariableServiceCIDR])
	IPFamily, err := GetIPFamilyForGivenClusterNetworkCIDRs(clusterCIDRs, serviceCIDRs, isIPV6Primary)
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
			// if path spec.topology.variables.proxy has any child attributes then enable TKG_HTTP_PROXY_ENABLED, spec.topology.variables.proxy mapped to TKG_HTTP_PROXY_ENABLED
			if strings.HasPrefix(clusterAttributePath, "spec.topology.variables.proxy") {
				inputVariablesMap["spec.topology.variables.proxy"] = true
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
// tkg-unknown-cluster1 : this is not valid name,  "unknown" is not valid infra provider name.
func getProviderNameFromTopologyClassName(topologyClassValue interface{}) (string, error) {
	var provider string
	if topologyClassValue == nil || !regExpTopologyClassVal.MatchString(topologyClassValue.(string)) {
		return provider, errors.New(constants.TopologyClassIncorrectValueErrMsg)
	}
	return strings.Split(topologyClassValue.(string), "-")[1], nil
}

// overrideClusterOptionsWithLatestEnvironmentConfigurationValues overrides CreateClusterOptions attributes with latest values
// from the environment only if environment has value, the variable values in environment could be updated from input config file or input cluster class file.
func (t *tkgctl) overrideClusterOptionsWithLatestEnvironmentConfigurationValues(cc *CreateClusterOptions) {
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName); err == nil {
		cc.ClusterName = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan); err == nil {
		cc.Plan = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace); err == nil {
		cc.Namespace = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider); err == nil {
		cc.InfrastructureProvider = val
	}
	if val, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableControlPlaneMachineCount, t.TKGConfigReaderWriter()); err == nil {
		cc.ControlPlaneMachineCount = val
	}
	if val, err := tkgconfighelper.GetIntegerVariableFromConfig(constants.ConfigVariableWorkerMachineCount, t.TKGConfigReaderWriter()); err == nil {
		cc.WorkerMachineCount = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize); err == nil {
		cc.Size = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize); err == nil {
		cc.ControlPlaneSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize); err == nil {
		cc.WorkerSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI); err == nil {
		cc.CniType = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableEnableClusterOptions); err == nil {
		cc.EnableClusterOptions = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint); err == nil {
		cc.VsphereControlPlaneEndpoint = val
	}
}

// overrideManagementClusterOptionsWithLatestEnvironmentConfigurationValues overrides InitRegion attributes with latest values
// from the environment only if environment has value, the variable values in environment could be updated from input config file or input cluster class file.
func (t *tkgctl) overrideManagementClusterOptionsWithLatestEnvironmentConfigurationValues(ir *InitRegionOptions) {
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterName); err == nil {
		ir.ClusterName = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableNamespace); err == nil {
		ir.Namespace = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableClusterPlan); err == nil {
		ir.Plan = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableInfraProvider); err == nil {
		ir.InfrastructureProvider = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableSize); err == nil {
		ir.Size = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableControlPlaneSize); err == nil {
		ir.ControlPlaneSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableWorkerSize); err == nil {
		ir.WorkerSize = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableCNI); err == nil {
		ir.CniType = val
	}
	if val, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableVsphereControlPlaneEndpoint); err == nil {
		ir.VsphereControlPlaneEndpoint = val
	}
}

// GetIPFamilyForGivenClusterNetworkCIDRs takes clusterNetwork's pods CIDRs and service CIDRs, and returns IP family type
// The each input cidrs string can have multiple CIDR's separated with comma (,), but it can not have more than two on each input CIDR string.
// If input CIDR has both ipv6 and ipv4, and if isIPV6Primary is true then the first CIDR should be ipv6, and returns "ipv6,ipv4",  else if isIPV6Primary is false then returns "ipv4,ipv6"
// both input CIDRs should be same family
func GetIPFamilyForGivenClusterNetworkCIDRs(podsCIDRs, serviceCIDRs string, isIPV6Primary bool) (string, error) {
	DefaultIPFamily := constants.IPv4Family
	var podsIPFamily, serviceIPFamily string
	var err error
	if podsCIDRs != "" {
		podsIPFamily, err = GetIPFamilyForGivenCIDRs(strings.Split(podsCIDRs, ","), isIPV6Primary)
		if err != nil {
			return DefaultIPFamily, err
		}
	} else {
		podsIPFamily = DefaultIPFamily
	}
	if serviceCIDRs != "" {
		serviceIPFamily, err = GetIPFamilyForGivenCIDRs(strings.Split(serviceCIDRs, ","), isIPV6Primary)
		if err != nil {
			return DefaultIPFamily, err
		}
	} else {
		serviceIPFamily = DefaultIPFamily
	}

	// If the IP Family of both CIDRs not same then return error
	if serviceIPFamily != podsIPFamily {
		return DefaultIPFamily, fmt.Errorf(errMessageCIDRsIPFamiliesNotSame, podsCIDRs, podsIPFamily, serviceCIDRs, serviceIPFamily)
	}
	return serviceIPFamily, nil
}

// stringArrayToStringWithCommaSeparatedElements converts given string (which has array data in fmt.Println output format) to string which has array elements separated with comma (,)
// eg: Input string is "[100.96.0.0/11 100.64.0.0/18]" converts to "100.64.0.0/18,100.64.0.0/18" of string type.
func stringArrayToStringWithCommaSeparatedElements(arrayDataInStringFormat string) string {
	if arrayDataInStringFormat == "" {
		return ""
	}
	return strings.Join(strings.Split(strings.Trim(arrayDataInStringFormat, "[]"), " "), ",")
}

// GetIPFamilyForGivenCIDRs takes cidrs array and returns ip family type (ipv4, ipv6 or dual)
// Maximum input cidrs array length is 2 only - "[100.64.0.0/18,100.64.0.0/18]"
// If input cidr's can be  ipv4, ipv6 or both.
// if isIPV6Primary is true then the first cidrs value, which is cidrs[0] must be ipv6 otherwise throws error
// if input cidrs has both ipv6 and ipv4 then returns "ipv6,ipv4" only if isIPV6Primary true, otherwise returns "ipv4,ipv6"
// If all input cidr's has only ipv6 then returns "ipv6"
// If all input cidr's has only ipv4 then returns "ipv4"
func GetIPFamilyForGivenCIDRs(cidrs []string, isIPV6Primary bool) (string, error) {
	DefaultIPFamily := constants.IPv4Family
	if len(cidrs) > 2 {
		return DefaultIPFamily, fmt.Errorf("too many CIDRs specified: %v", cidrs)
	}
	var foundIPv4 bool
	var foundIPv6 bool
	for i, cidr := range cidrs {
		ip, _, err := net.ParseCIDR(cidr)
		if err != nil {
			return DefaultIPFamily, fmt.Errorf("could not parse CIDR %v, error: %s", cidr, err)
		}
		if ip.To4() != nil {
			foundIPv4 = true
			// if isIPV6Primary true, then the first value should be ipv6 family otherwise throw error
			if isIPV6Primary && i == 0 {
				return DefaultIPFamily, fmt.Errorf(errMessageIPv6EnabledCIDRHasNoIPv6, cidrs)
			}
		} else {
			foundIPv6 = true
		}
	}
	switch {
	case foundIPv4 && foundIPv6:
		if isIPV6Primary {
			return constants.DualStackPrimaryIPv6Family, nil
		}
		// return "ipv4,ipv6" even first value is ipv6, if isIPV6Primary is false
		return constants.DualStackPrimaryIPv4Family, nil
	case foundIPv4:
		return constants.IPv4Family, nil
	case foundIPv6:
		return constants.IPv6Family, nil
	default:
		return DefaultIPFamily, nil
	}
}
