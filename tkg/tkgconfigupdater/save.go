// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

func encodeMap(data map[string]string) {
	for key, value := range data {
		if shouldEncode(key) {
			encodedValue, err := encodeValueIfRequired(value)
			if err == nil {
				data[key] = encodedValue
			}
		}
	}
}

func shouldEncode(key string) bool {
	return utils.ContainsString(KeysToEncode, key)
}

// SaveConfig saves configuration to ReaderWriter and/or to File
// allows clusterConfigPath to be empty, in which case no file is written
// allows TKGConfigReaderWriter to be empty, in which case no update is made to that writer
// (allows both to be empty, in which case the method does nothing)
func SaveConfig(clusterConfigPath string, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter, configObj interface{}) error {
	configMap, err := CreateConfigMap(configObj)
	if err != nil {
		return err
	}

	if tkgConfigReaderWriter != nil {
		tkgConfigReaderWriter.SetMap(configMap)
	}
	if len(clusterConfigPath) > 0 {
		err = SetConfig(configMap, clusterConfigPath)
	}

	return err
}

// CreateConfigMap transforms a configuration object into a map
func CreateConfigMap(configObj interface{}) (map[string]string, error) {
	// turn the object into a YAML byte array
	configByte, err := yaml.Marshal(configObj)
	if err != nil {
		return nil, errors.Wrap(err, "unable to marshal config object")
	}

	// turn the byte array into a map
	configMap := make(map[string]string)

	err = yaml.Unmarshal(configByte, &configMap)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling file")
	}

	// certain values (like passwords) need encoding
	encodeMap(configMap)

	return configMap, nil
}

// SetConfig saves map of key-value pairs to config file (creating if nec)
func SetConfig(data map[string]string, clusterConfigPath string) error {
	// Ensure the directory exists
	clusterConfigDir := filepath.Dir(clusterConfigPath)
	if _, err := os.Stat(clusterConfigDir); os.IsNotExist(err) {
		if err = os.MkdirAll(clusterConfigDir, constants.DefaultDirectoryPermissions); err != nil {
			return errors.Wrapf(err, "cannot create cluster config directory '%s'", clusterConfigDir)
		}
	}

	// unmarshal the config into map
	tkgConfigMap := make(map[string]interface{})
	// read tkg config file (if it exists)
	fileData, err := os.ReadFile(clusterConfigPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.Errorf("unable to read tkg configuration from: %s", clusterConfigPath)
		}
		// if the file doesn't exist, just continue
	} else {
		// unmarshal the existing config into map
		err = yaml.Unmarshal(fileData, &tkgConfigMap)
		if err != nil {
			return errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", clusterConfigPath)
		}
	}

	mergeParamDataWithConfigFile(data, tkgConfigMap)

	outBytes, err := yaml.Marshal(&tkgConfigMap)
	if err != nil {
		return errors.Wrapf(err, "error marshaling configuration file")
	}
	err = os.WriteFile(clusterConfigPath, outBytes, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "error writing configuration file")
	}

	return nil
}

func mergeParamDataWithConfigFile(data map[string]string, tkgConfigMap map[string]interface{}) {
	// copy the parameter data over top any existing configuration data
	for key, value := range data {
		if !shouldExcludeFromFile(key) {
			tkgConfigMap[key] = value
		}
	}
}

func shouldExcludeFromFile(key string) bool {
	return utils.ContainsString(KeysToNeverPersist, key)
}

// GetNodeIndex returns index of the node
func GetNodeIndex(node []*yaml.Node, key string) int {
	appIdx := -1
	for i, k := range node {
		if i%2 == 0 && k.Value == key {
			appIdx = i + 1
			break
		}
	}
	return appIdx
}
