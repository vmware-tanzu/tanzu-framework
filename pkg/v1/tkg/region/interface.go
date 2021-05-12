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

package region

// DeploymentStatus holds the region deployment status
type DeploymentStatus string

// Possible deployment statuses
const (
	Success DeploymentStatus = "Success"
	Failed  DeploymentStatus = "Failed"
)

// RegionContext holds the region context
type RegionContext struct { //nolint:golint
	ClusterName      string           `yaml:"name" json:"name"`
	ContextName      string           `yaml:"context" json:"context"`
	SourceFilePath   string           `yaml:"file" json:"file"`
	Status           DeploymentStatus `yaml:"status" json:"status"`
	IsCurrentContext bool             `yaml:"isCurrentContext" json:"isCurrentContext"`
}

// Manager manages tkg regions
type Manager interface {
	// ListRegionContexts lists all the regions in tkg config file
	ListRegionContexts() ([]RegionContext, error)
	// SaveRegionContext saves a new region object into the tkg config file,
	// Errors will be returned if a region with same name and context already exists
	SaveRegionContext(region RegionContext) error
	// UpsertRegionContext updates region context object  if already exists,
	// else saves the new region object into the tkg config file
	UpsertRegionContext(region RegionContext) error
	// DeleteRegionContext deletes all region info with the given cluster name, regardless of context
	DeleteRegionContext(clusterName string) error
	// SetCurrentContext sets current regional context into tkg config file
	SetCurrentContext(clusterName string, contextName string) error
	// GetCurrentContext gets current regional context from tkg config file
	GetCurrentContext() (RegionContext, error)
}
