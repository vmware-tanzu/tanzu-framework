// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package core

import (
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"strings"
	"testing"
)

// Test_config_MalformedPathArg validates functionality when an invalid argument is provided.
func Test_config_MalformedPathArg(t *testing.T) {
	err := setFeature(nil, "invalid-arg", "")
	if err == nil {
		t.Error("Malformed path argument should have resulted in an error")
	}

	if !strings.Contains(err.Error(), "unable to parse config path parameter") {
		t.Errorf("Unexpected error message returned for malformed path argument: %s", err.Error())
	}
}

// Test_config_InvalidPathArg validates functionality when an invalid argument is provided.
func Test_config_InvalidPathArg(t *testing.T) {
	err := setFeature(nil, "shouldbefeatures.plugin.feature", "")
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
	err := setFeature(cfg, "unstable-versions", "experimental")
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
	err := setFeature(cfg, "unstable-versions", "invalid-unstable-versions-value")
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
	err := setFeature(cfg, "features.global.foo", "bar")
	if err != nil {
		t.Errorf("Unexpected error returned for global features path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Features["global"]["foo"] != "bar" {
		t.Error("cfg.ClientOptions.Features[\"global\"][\"foo\"] was not assigned the value \"bar\"")
	}
}

// Test_config_Feature validates functionality when normal feature path argument is provided.
func Test_config_Feature(t *testing.T) {
	cfg := &configv1alpha1.ClientConfig{}
	err := setFeature(cfg, "features.any-plugin.foo", "bar")
	if err != nil {
		t.Errorf("Unexpected error returned for any-plugin features path argument: %s", err.Error())
	}

	if cfg.ClientOptions.Features["any-plugin"]["foo"] != "bar" {
		t.Error("cfg.ClientOptions.Features[\"any-plugin\"][\"foo\"] was not assigned the value \"bar\"")
	}
}
