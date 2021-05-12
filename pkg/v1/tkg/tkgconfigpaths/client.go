/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package tkgconfigpaths provides utilities to get info related to TKG configuration paths.
package tkgconfigpaths

type client struct {
	configDir string
}

// New creates new tkg configuration paths client
func New(configDir string) Client {
	tkgconfigpaths := &client{
		configDir: configDir,
	}
	return tkgconfigpaths
}

// Client implements TKG configuration paths functions
type Client interface {
	// GetTKGDirectory returns path to tkg config directory "$HOME/.tkg"
	GetTKGDirectory() (string, error)

	// GetTKGProvidersDirectory returns path to tkg config directory "$HOME/.tkg/providers"
	GetTKGProvidersDirectory() (string, error)

	// GetTKGBoMDirectory returns path to tkg config directory "$HOME/.tkg/bom"
	GetTKGBoMDirectory() (string, error)

	// GetTKGConfigDirectories returns tkg config directories in below order
	// (tkgDir, bomDir, providersDir, error)
	GetTKGConfigDirectories() (string, string, string, error)

	// GetProvidersConfigFilePath returns config file path from providers dir
	// "$HOME/.tkg/providers/config.yaml"
	GetProvidersConfigFilePath() (string, error)

	// GetTKGConfigPath returns tkg configfile path
	GetTKGConfigPath() (string, error)

	// GetDefaultClusterConfigPath returns default cluster config file path
	GetDefaultClusterConfigPath() (string, error)
}
