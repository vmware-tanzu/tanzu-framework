// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	// EnvPluginStateKey is the environment key that contains the path to the
	// plugin state file.
	EnvPluginStateKey = "TANZU_STATE"
)

// PluginState is state that will be passed to plugins.
type PluginState struct {
	Auth string `json:"auth" yaml:"auth"`
}

// ReadPluginStateFromPath read states from a path on disk.
func ReadPluginStateFromPath(p string) (*PluginState, error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, fmt.Errorf("read state: %w", err)
	}

	s := &PluginState{}
	if err := json.Unmarshal(b, s); err != nil {
		return nil, fmt.Errorf("decode state: %w", err)
	}

	return s, nil
}
