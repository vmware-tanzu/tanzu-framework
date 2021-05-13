// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package features

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const (
	defaultConfigFile = "features.json"
)

type client struct {
	featureFlagConfigPath string
}

// New creates new features client, defaults config file path to ~/.tkg/features.yaml
func New(featureFlagsConfigFolder, featureFlagsConfigFile string) (Client, error) {
	if featureFlagsConfigFile == "" {
		featureFlagsConfigFile = defaultConfigFile
	}
	if featureFlagsConfigFolder == "" {
		return nil, errors.New("featureFlagsConfigFolder connot be empty")
	}
	return &client{
		featureFlagConfigPath: filepath.Join(featureFlagsConfigFolder, featureFlagsConfigFile),
	}, nil
}

// Write feature flags, overwrites config file if it already exists
func (c *client) WriteFeatureFlags(featureFlags map[string]string) error {
	featureFlagsList := make(map[string]string)
	for feature, value := range featureFlags {
		featureFlagsList[strings.TrimSpace(feature)] = value
	}
	jsonString, err := json.Marshal(featureFlagsList)
	if err != nil {
		return errors.Wrap(err, "failed to parse the feature flags as valid json")
	}
	file, err := os.Create(c.featureFlagConfigPath)
	if err != nil {
		return errors.Wrap(err, "failed to create feature flag file")
	}
	defer file.Close()
	_, err = file.WriteString(string(jsonString))
	if err != nil {
		return errors.Wrap(err, "failed to write to feature flag file")
	}
	return nil
}

// List all feature flags
func (c *client) GetFeatureFlags() (map[string]string, error) {
	jsonFile, err := os.Open(c.featureFlagConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open feature flag file")
	}
	defer jsonFile.Close()
	byteValue, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read feature flag file")
	}
	result := make(map[string]string)
	err = json.Unmarshal(byteValue, &result)
	if err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal feature flag json data")
	}
	return result, nil
}

func (c *client) IsFeatureFlagEnabled(featureName string) (bool, error) {
	featureFlags, err := c.GetFeatureFlags()
	if err != nil {
		return false, err
	}
	for feature := range featureFlags {
		if feature == featureName {
			return true, nil
		}
	}
	return false, nil
}

func (c *client) GetFeatureFlag(featureName string) (string, error) {
	featureFlags, err := c.GetFeatureFlags()
	if err != nil {
		return "", err
	}
	for feature, v := range featureFlags {
		if feature == featureName {
			return v, nil
		}
	}
	return "", errors.Errorf("feature flag %s not set", featureName)
}
