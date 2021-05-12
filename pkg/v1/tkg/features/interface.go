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

// Package features provides TKG related feature enablement functionalities
package features

// Client defines methods to access feature flags
type Client interface {
	GetFeatureFlags() (map[string]string, error)
	IsFeatureFlagEnabled(string) (bool, error)
	WriteFeatureFlags(map[string]string) error
	GetFeatureFlag(string) (string, error)
}
