// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

// Test_config_MalformedPathArg validates functionality when an invalid argument is provided.
func TestConfigMalformedPathArg(t *testing.T) {
	err := setConfiguration(nil, "invalid-arg", "")
	if err == nil {
		t.Error("Malformed path argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "unable to parse config path parameter") {
		t.Errorf("Unexpected error message returned for malformed path argument: %s", err.Error())
	}
}

// Test_config_InvalidPathArg validates functionality when an invalid argument is provided.
func TestConfigInvalidPathArg(t *testing.T) {
	err := setConfiguration(nil, "shouldbefeatures.plugin.feature", "")
	if err == nil {
		t.Error("Invalid path argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "unsupported config path parameter") {
		t.Errorf("Unexpected error message returned for invalid path argument: %s", err.Error())
	}
}

// TestConfigUnstableVersions validates functionality when path argument unstable-versions is provided.
func TestConfigUnstableVersions(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	const path = "unstable-versions"
	const value = "experimental"
	err := setConfiguration(cfg, path, value)
	if err != nil {
		t.Errorf("Unexpected error returned for %s path argument: %s", path, err.Error())
	}

	if cfg.ClientOptions.CLI.UnstableVersionSelector != value {
		t.Error("Unstable version was not set correctly for valid value")
	}
}

// TestConfigUnstableVersions validates functionality when path argument cli.unstable-versions is provided.
func TestConfigCliUnstableVersions(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	const path = "cli.unstable-versions"
	const value = "all"
	err := setConfiguration(cfg, path, value)
	if err != nil {
		t.Errorf("Unexpected error returned for %s path argument: %s", path, err.Error())
	}

	if cfg.ClientOptions.CLI.UnstableVersionSelector != value {
		t.Error("Unstable version was not set correctly for valid value")
	}
}

// TestConfigInvalidUnstableVersions validates functionality when invalid unstable-versions is provided.
func TestConfigInvalidUnstableVersions(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	err := setConfiguration(cfg, "unstable-versions", "invalid-unstable-versions-value")
	if err == nil {
		t.Error("Invalid unstable-versions should have resulted in error")
	}

	if !strings.Contains(err.Error(), "unknown unstable-versions setting") {
		t.Errorf("Unexpected error message returned for invalid unstable-versions value: %s", err.Error())
	}
}

// TestConfigGlobalFeature validates functionality when global feature path argument is provided.
func TestConfigGlobalFeature(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	value := "bar"
	err := setConfiguration(cfg, "features.global.foo", value)
	if err != nil {
		t.Errorf("Unexpected error returned for global features path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Features["global"]["foo"] != value {
		t.Error("cfg.ClientOptions.Features[\"global\"][\"foo\"] was not assigned the value \"" + value + "\"")
	}
}

// TestConfigFeature validates functionality when normal feature path argument is provided.
func TestConfigFeature(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	value := "barr"
	err := setConfiguration(cfg, "features.any-plugin.foo", value)
	if err != nil {
		t.Errorf("Unexpected error returned for any-plugin features path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Features["any-plugin"]["foo"] != value {
		t.Error("cfg.ClientOptions.Features[\"any-plugin\"][\"foo\"] was not assigned the value \"" + value + "\"")
	}
}

// TestConfigSetUnsetEnv validates set and unset functionality when env config path argument is provided.
func TestConfigSetUnsetEnv(t *testing.T) {
	assert := assert.New(t)

	cfg := &configapi.ClientConfig{}
	value := "baar"
	err := setConfiguration(cfg, "env.foo", value)
	assert.Nil(err)
	assert.Equal(value, cfg.ClientOptions.Env["foo"])

	err = unsetConfiguration(cfg, "env.foo")
	assert.Nil(err)
	assert.Equal(cfg.ClientOptions.Env["foo"], "")
}

// TestConfigIncorrectConfigLiteral validates incorrect config literal
func TestConfigIncorrectConfigLiteral(t *testing.T) {
	assert := assert.New(t)

	cfg := &configapi.ClientConfig{}
	value := "b"
	err := setConfiguration(cfg, "fake.any-plugin.foo", value)
	assert.NotNil(err)
	assert.Contains(err.Error(), "unsupported config path parameter [fake] (was expecting 'features.<plugin>.<feature>' or 'env.<env_variable>')")
}

// TestConfigEnv validates functionality when normal env path argument is provided.
func TestConfigEnv(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	value := "baarr"
	err := setConfiguration(cfg, "env.any-plugin", value)
	if err != nil {
		t.Errorf("Unexpected error returned for any-plugin env path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Env["any-plugin"] != value {
		t.Error("cfg.ClientOptions.Features[\"any-plugin\"][\"foo\"] was not assigned the value \"" + value + "\"")
	}
}

func TestConfigEditionCommunity(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	value := configapi.EditionCommunity
	err := setConfiguration(cfg, "cli.edition", value)
	if err != nil {
		t.Errorf("Unexpected error returned for cli.edition argument: %s", err.Error())
	}

	if cfg.ClientOptions.CLI.Edition != configapi.EditionCommunity { //nolint:staticcheck
		t.Error("cfg.ClientOptions.CLI.Edition was not assigned the value \"" + value + "\"")
	}
}

func TestConfigEditionStandard(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	value := configapi.EditionStandard
	err := setConfiguration(cfg, "cli.edition", value)
	if err != nil {
		t.Errorf("Unexpected error returned for cli.edition argument: %s", err.Error())
	}

	if cfg.ClientOptions.CLI.Edition != configapi.EditionStandard { //nolint:staticcheck
		t.Error("cfg.ClientOptions.CLI.Edition was not assigned the value \"" + value + "\"")
	}
}

func TestConfigEditionInvalid(t *testing.T) {
	cfg := &configapi.ClientConfig{}
	value := "invalidEdition"
	err := setConfiguration(cfg, "cli.edition", value)
	if err == nil {
		t.Errorf("Expected error returned for cli.edition argument: %s", value)
	}
}
