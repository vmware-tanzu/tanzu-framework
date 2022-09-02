// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package region

// DeploymentStatus holds the region deployment status
type DeploymentStatus string

// Possible deployment statuses
const (
	Success DeploymentStatus = "Success"
	Failed  DeploymentStatus = "Failed"
)

// RegionContext holds the region context
type RegionContext struct {
	ClusterName      string           `yaml:"name" json:"name"`
	ContextName      string           `yaml:"context" json:"context"`
	SourceFilePath   string           `yaml:"file" json:"file"`
	Status           DeploymentStatus `yaml:"status" json:"status"`
	IsCurrentContext bool             `yaml:"isCurrentContext" json:"isCurrentContext"`
}

//go:generate counterfeiter -o ../fakes/regionmanager.go --fake-name RegionManager . Manager

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
