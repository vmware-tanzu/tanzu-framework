// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndDeleteConfigMetadataSettings(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tests := []struct {
		name  string
		key   string
		value bool
	}{
		{
			name:  "success context-aware-cli-for-plugins",
			key:   SettingUseUnifiedConfig,
			value: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := SetConfigMetadataSetting(tc.key, strconv.FormatBool(tc.value))
			assert.NoError(t, err)

			useUnifiedConfig, err := UseUnifiedConfig()
			assert.NoError(t, err)
			assert.Equal(t, tc.value, useUnifiedConfig)

			err = DeleteConfigMetadataSetting(tc.key)
			assert.NoError(t, err)

			useUnifiedConfig, err = UseUnifiedConfig()
			assert.Equal(t, "not found", err.Error())
			assert.Equal(t, tc.value, useUnifiedConfig)

			err = SetConfigMetadataSetting(tc.key, strconv.FormatBool(!tc.value))
			assert.NoError(t, err)

			useUnifiedConfig, err = UseUnifiedConfig()
			assert.NoError(t, err)
			assert.Equal(t, !tc.value, useUnifiedConfig)
		})
	}
}

func TestSetConfigMetadataSetting(t *testing.T) {
	// setup
	func() {
		LocalDirName = TestLocalDirName
	}()
	defer func() {
		cleanupDir(LocalDirName)
	}()
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "success disable useUnifiedConfig",
			key:   SettingUseUnifiedConfig,
			value: "false",
		},
		{
			name:  "success enable useUnifiedConfig",
			key:   SettingUseUnifiedConfig,
			value: "true",
		},
		{
			name:  "success disable useUnifiedConfig",
			key:   SettingUseUnifiedConfig,
			value: "false",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := SetConfigMetadataSetting(tc.key, tc.value)
			assert.NoError(t, err)

			useUnifiedConfig, err := UseUnifiedConfig()
			assert.NoError(t, err)
			expected, err := strconv.ParseBool(tc.value)
			assert.NoError(t, err)
			assert.Equal(t, expected, useUnifiedConfig)
		})
	}
}
