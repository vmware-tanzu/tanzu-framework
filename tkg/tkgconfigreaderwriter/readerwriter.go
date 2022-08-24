// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigreaderwriter

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/client/config"
)

// tkgConfigReaderWriter is a customized implementation of viperReader in clusterctl repo to make it compatible with
// tkg configuration file with some additional functionality to update configuration file
type tkgConfigReaderWriter struct {
	viperStore *viper.Viper
}

//go:generate counterfeiter -o ../fakes/readerwriter.go --fake-name TKGConfigReaderWriter . TKGConfigReaderWriter

// TKGConfigReaderWriter defines methods of reader which is implemented using viper for reading from environment variables
// and from a tkg config file. Also defines methods of writer to set/update the variables and config file
type TKGConfigReaderWriter interface {
	config.Reader

	MergeInConfig(configFilePath string) error
	SetMap(data map[string]string)
}

// newTKGConfigReaderWriter returns a viper and config reader writer
func newTKGConfigReaderWriter() TKGConfigReaderWriter {
	return &tkgConfigReaderWriter{}
}

// Init initialize the readerWriter
func (v *tkgConfigReaderWriter) Init(tkgConfigFile string) error {
	v.viperStore = viper.New()

	// Configure for reading environment variables as well, and more specifically:
	// AutomaticEnv force viper to check for an environment variable any time a v.viperStore.Get request is made.
	// It will check for a environment variable with a name matching the key uppercased; in case name use the - delimiter,
	// the SetEnvKeyReplacer forces matching to name use the _ delimiter instead (- is not allowed in linux env variable names).
	replacer := strings.NewReplacer("-", "_")
	v.viperStore.SetEnvKeyReplacer(replacer)
	// Allow user to use empty env variable value, it will allow empty variable value,
	// which would be necessary for Optional template variable
	v.viperStore.AllowEmptyEnv(true)
	v.viperStore.AutomaticEnv()

	// Use path file from the flag.
	v.viperStore.SetConfigFile(tkgConfigFile)
	v.viperStore.SetConfigType("yaml")
	// If a path file is found, read it in.
	if err := v.viperStore.ReadInConfig(); err != nil {
		return errors.Wrapf(err, "Error reading configuration file %q", v.viperStore.ConfigFileUsed())
	}

	return nil
}

func (v *tkgConfigReaderWriter) MergeInConfig(configFilePath string) error {
	// Use path file from the flag.
	v.viperStore.SetConfigFile(configFilePath)
	v.viperStore.SetConfigType("yaml")
	// If a path file is found, read it in.
	if err := v.viperStore.MergeInConfig(); err != nil {
		return errors.Wrapf(err, "Error reading configuration file %q", v.viperStore.ConfigFileUsed())
	}
	return nil
}

func (v *tkgConfigReaderWriter) Get(key string) (string, error) {
	if v.viperStore.Get(key) == nil {
		return "", errors.Errorf("Failed to get value for variable %q. Please set the variable value using os env variables or using the config file", key)
	}
	return v.viperStore.GetString(key), nil
}

func (v *tkgConfigReaderWriter) Set(key, value string) {
	v.viperStore.Set(key, value)
}

func (v *tkgConfigReaderWriter) SetMap(data map[string]string) {
	for key, val := range data {
		v.Set(key, val)
	}
}

func (v *tkgConfigReaderWriter) UnmarshalKey(key string, rawval interface{}) error {
	return v.viperStore.UnmarshalKey(key, rawval)
}
