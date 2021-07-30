// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigupdater

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

// SaveConfig saves and updates key-value pairs in existing config file
func SaveConfig(clusterConfigPath string, rw tkgconfigreaderwriter.TKGConfigReaderWriter, configObj interface{}) error {
	configByte, err := yaml.Marshal(configObj)
	if err != nil {
		return errors.Wrap(err, "unable to marshal config object")
	}

	configMap := make(map[string]string)

	err = yaml.Unmarshal(configByte, &configMap)
	if err != nil {
		return errors.Wrap(err, "error unmarshalling file")
	}

	for k, v := range configMap {
		if utils.ContainsString(KeysToEncode, k) {
			encodedValue, err := encodeValueIfRequired(v)
			if err == nil {
				v = encodedValue
			}
		}
		err = os.Setenv(k, v)
		if err != nil {
			log.Warningf("%s is not set to environment variable properly", k)
		}
		// sets variable in viper store
		rw.Set(k, v)

		if utils.ContainsString(KeysToNeverPersist, k) {
			continue
		}
		// sets variable in config file
		err = SetVariableInConfig(k, v, "", clusterConfigPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// SetVariableInConfig allows to explicit override of config value
// with updating the value in provided tkg-config file
// i.e. it will add the "key: value" in the tkg-config file if key is not present
// if key is present it will update the value of the key in the file
func SetVariableInConfig(key, value, comment, clusterConfigPath string) error {
	clusterConfigDir := filepath.Dir(clusterConfigPath)

	// read tkg config file
	fileData, err := os.ReadFile(clusterConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return errors.Errorf("unable to read tkg configuration from: %s", clusterConfigPath)
	} else if _, err := os.Stat(clusterConfigDir); os.IsNotExist(err) {
		if err = os.MkdirAll(clusterConfigDir, constants.DefaultDirectoryPermissions); err != nil {
			return errors.Wrapf(err, "cannot create cluster config directory '%s'", clusterConfigDir)
		}
	}

	// unmarshal the config into map
	tkgConfigMap := make(map[string]interface{})
	err = yaml.Unmarshal(fileData, &tkgConfigMap)
	if err != nil {
		return errors.Wrapf(err, "unable to unmarshal tkg configuration file %s", clusterConfigPath)
	}

	var out []byte
	tkgConfigMap[key] = value
	out, err = yaml.Marshal(&tkgConfigMap)

	if err != nil {
		return errors.Wrapf(err, "error marshaling while adding key: %v, value: %v", key, value)
	}
	err = os.WriteFile(clusterConfigPath, out, constants.ConfigFilePermissions)
	if err != nil {
		return errors.Wrapf(err, "error writing file while adding key: %v, value: %v", key, value)
	}

	return nil
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
