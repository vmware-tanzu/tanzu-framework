// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

const (
	ConfigFileName   = "config.yaml"
	ConfigNGFileName = "config-ng.yaml"
	ConfigFileDir    = ".config/tanzu/"
)

// ConfigOps performs "tanzu config" command operations
type ConfigOps interface {
	ConfigSetFeatureFlag(path, value string) error
	ConfigGetFeatureFlag(path string) (string, error)
	ConfigUnsetFeature(path string) error
	ConfigInit() error
	ConfigServerList() error
	ConfigServerDelete(serverName string) error
	DeleteCLIConfigurationFiles() error
	IsCLIConfigurationFilesExists() bool
}

// configOps is the implementation of ConfOps interface
type configOps struct {
	CmdOps
}

func NewConfOps() ConfigOps {
	return &configOps{
		CmdOps: NewCmdOps(),
	}
}

// ConfigSetFeature sets the tanzu config feature flag
func (co *configOps) ConfigSetFeatureFlag(path, value string) (err error) {
	confSetCmd := ConfigSet + path + " " + value
	_, _, err = co.Exec(confSetCmd)
	return err
}

// ConfigSetFeature sets the tanzu config feature flag
func (co *configOps) ConfigGetFeatureFlag(path string) (string, error) {
	out, _, err := co.Exec(ConfigGet)
	if err != nil {
		return "", err
	}

	var cnf *configapi.ClientConfig
	err = yaml.Unmarshal(out.Bytes(), &cnf)
	if err != nil {
		return "", errors.Wrap(err, "failed to construct yaml node from config get output")
	}
	featureName := strings.Split(path, ".")[len(strings.Split(path, "."))-1]
	pluginName := strings.Split(path, ".")[len(strings.Split(path, "."))-2]
	if cnf != nil && cnf.ClientOptions.Features[pluginName] != nil {
		return cnf.ClientOptions.Features[pluginName][featureName], nil
	}
	return "", err
}

// ConfigUnsetFeature un-sets the tanzu config feature flag
func (co *configOps) ConfigUnsetFeature(path string) (err error) {
	unsetFeatureCmd := ConfigUnset + path
	_, _, err = co.Exec(unsetFeatureCmd)
	return
}

// ConfigInit performs "tanzu config init"
func (co *configOps) ConfigInit() (err error) {
	_, _, err = co.Exec(ConfigInit)
	return
}

// ConfigServerList returns the server list
// TODO: should return the servers info in proper format
func (co *configOps) ConfigServerList() (err error) {
	_, _, err = co.Exec(ConfigServerList)
	return
}

// ConfigServerDelete deletes a server from tanzu config
func (co *configOps) ConfigServerDelete(serverName string) error {
	_, _, err := co.Exec(ConfigServerDelete + serverName)
	return err
}

// DeleteCLIConfigurationFiles deletes cli configuration files
func (co *configOps) DeleteCLIConfigurationFiles() error {
	homeDir, _ := os.UserHomeDir()
	configFile := filepath.Join(homeDir, ConfigFileDir, ConfigFileName)
	_, err := os.Stat(configFile)
	if err == nil {
		if ferr := os.Remove(configFile); ferr != nil {
			return ferr
		}
	}
	configNGFile := filepath.Join(homeDir, ConfigFileDir, ConfigNGFileName)
	if _, err := os.Stat(configNGFile); err == nil {
		if ferr := os.Remove(configNGFile); ferr != nil {
			return ferr
		}
	}
	return nil
}

// IsCLIConfigurationFilesExists checks the existence of cli configuration files
func (co *configOps) IsCLIConfigurationFilesExists() bool {
	homeDir, _ := os.UserHomeDir()
	configFilePath := filepath.Join(homeDir, ConfigFileDir, ConfigFileName)
	configNGFilePath := filepath.Join(homeDir, ConfigFileDir, ConfigNGFileName)
	_, err1 := os.Stat(configFilePath)
	_, err2 := os.Stat(configNGFilePath)
	if err1 == nil && err2 == nil {
		return true
	}
	return false
}
