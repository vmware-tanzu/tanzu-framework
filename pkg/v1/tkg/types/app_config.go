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

// Package types package to store configs
package types

import (
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/providerinterface"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/region"
)

// AppConfig stores configuration related to running app
type AppConfig struct {
	TKGConfigDir      string
	ProviderGetter    providerinterface.ProviderInterface
	CustomizerOptions CustomizerOptions
	TKGSettingsFile   string
}

// CustomizerOptions provides overrides for CreateAllClients that allows a
// user to customize the underying clients.
type CustomizerOptions struct {
	RegionManagerFactory region.ManagerFactory
}
