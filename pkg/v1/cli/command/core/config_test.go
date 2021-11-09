// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
)

// Test_config_MalformedPathArg validates functionality when an invalid argument is provided.
func Test_config_MalformedPathArg(t *testing.T) {
	err := setConfiguration(nil, "invalid-arg", "")
	if err == nil {
		t.Error("Malformed path argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "unable to parse config path parameter") {
		t.Errorf("Unexpected error message returned for malformed path argument: %s", err.Error())
	}
}

// Test_config_InvalidPathArg validates functionality when an invalid argument is provided.
func Test_config_InvalidPathArg(t *testing.T) {
	err := setConfiguration(nil, "shouldbefeatures.plugin.feature", "")
	if err == nil {
		t.Error("Invalid path argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "unsupported config path parameter") {
		t.Errorf("Unexpected error message returned for invalid path argument: %s", err.Error())
	}
}

// Test_config_UnstableVersions validates functionality when path argument unstable-versions is provided.
func Test_config_UnstableVersions(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{}
	err := setConfiguration(cfg, "unstable-versions", "experimental")
	if err != nil {
		t.Errorf("Unexpected error returned for unstable-versions path argument: %s", err.Error())
	}

	if cfg.ClientOptions.CLI.UnstableVersionSelector != "experimental" {
		t.Error("Unstable version was not set correctly for valid value")
	}
}

// Test_config_InvalidUnstableVersions validates functionality when invalid unstable-versions is provided.
func Test_config_InvalidUnstableVersions(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{}
	err := setConfiguration(cfg, "unstable-versions", "invalid-unstable-versions-value")
	if err == nil {
		t.Error("Invalid unstable-versions should have resulted in error")
	}

	if !strings.Contains(err.Error(), "unknown unstable-versions setting") {
		t.Errorf("Unexpected error message returned for invalid unstable-versions value: %s", err.Error())
	}
}

// Test_config_GlobalFeature validates functionality when global feature path argument is provided.
func Test_config_GlobalFeature(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{}
	value := "bar"
	err := setConfiguration(cfg, "features.global.foo", value)
	if err != nil {
		t.Errorf("Unexpected error returned for global features path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Features["global"]["foo"] != value {
		t.Error("cfg.ClientOptions.Features[\"global\"][\"foo\"] was not assigned the value \"" + value + "\"")
	}
}

// Test_config_Feature validates functionality when normal feature path argument is provided.
func Test_config_Feature(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{}
	value := "barr"
	err := setConfiguration(cfg, "features.any-plugin.foo", value)
	if err != nil {
		t.Errorf("Unexpected error returned for any-plugin features path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Features["any-plugin"]["foo"] != value {
		t.Error("cfg.ClientOptions.Features[\"any-plugin\"][\"foo\"] was not assigned the value \"" + value + "\"")
	}
}

// Test_config_GlobalEnv validates functionality when env feature path argument is provided.
func Test_config_GlobalEnv(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{}
	value := "baar"
	err := setConfiguration(cfg, "env.global.foo", value)
	if err != nil {
		t.Errorf("Unexpected error returned for global env path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Env["global"]["foo"] != value {
		t.Error("cfg.ClientOptions.Env[\"global\"][\"foo\"] was not assigned the value \"" + value + "\"")
	}
}

// Test_config_Env validates functionality when normal env path argument is provided.
func Test_config_Env(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{}
	value := "baarr"
	err := setConfiguration(cfg, "env.any-plugin.foo", value)
	if err != nil {
		t.Errorf("Unexpected error returned for any-plugin env path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Env["any-plugin"]["foo"] != value {
		t.Error("cfg.ClientOptions.Features[\"any-plugin\"][\"foo\"] was not assigned the value \"" + value + "\"")
	}
}

// Test_config_Env validates functionality when normal env path argument is provided.
func Test_config_IncorrectConfigLiteral(t *testing.T) {
	assert := assert.New(t)

	cfg := &configv1alpha1.ClientConfig{}
	value := "b"
	err := setConfiguration(cfg, "fake.any-plugin.foo", value)
	assert.NotNil(err)
	assert.Contains(err.Error(), "unsupported config path parameter [fake] (was expecting 'features.<plugin>.<feature>' or 'env.<plugin>.<env_variable>')")
}
