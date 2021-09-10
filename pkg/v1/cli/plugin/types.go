// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package plugin

import (
	"github.com/pkg/errors"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/artifacts"
)

// plugin is an installable CLI plugin.
type plugin struct {
	Name string

	Description string

	Optional bool

	RecommendedVersion string

	Artifacts map[string]cliv1alpha1.ArtifactList

	// Discovery specificies the name of the discovery from where
	// this plugin is discovered.
	Discovery string
	// Scope specificies the scope of the plugin. Stand-Alone or Context
	Scope string
	// Status specificies the current plugin installation status
	Status string
}

func NewPlugin(p cliv1alpha1.CLIPlugin) Plugin {
	return &plugin{
		Name:               p.GetObjectMeta().GetName(),
		Description:        p.Spec.Description,
		Optional:           p.Spec.Optional,
		RecommendedVersion: p.Spec.RecommendedVersion,
		Artifacts:          p.Spec.Artifacts,
	}
}

// Name is the name of the plugin.
func (po *plugin) GetName() string {
	return po.Name
}

// Description is the plugin's description.
func (po *plugin) GetDescription() string {
	return po.Description
}

// Required denotes if this plugin is needed for all or at least most use cases.
func (po *plugin) IsRequired() bool {
	return !po.Optional
}

// GetDiscovery specificies the name of the discovery from where
// this plugin is discovered.
func (po *plugin) GetDiscovery() string {
	return po.Discovery
}

// Scope specificies the scope of the plugin. Stand-Alone or Context
func (po *plugin) GetScope() string {
	return po.Scope
}

// Status specificies the current plugin installation status
func (po *plugin) GetStatus() string {
	return po.Status
}

// SupportedVersions determines the list of supported CLI plugin versions.
// The values are sorted in the semver prescribed order as defined in
// https://github.com/Masterminds/semver#sorting-semantic-versions.
func (po *plugin) GetSupportedVersions() []string {
	supportedVersions := []string{}
	for v := range po.Artifacts {
		supportedVersions = append(supportedVersions, v)
	}
	return supportedVersions
}

// RecommendedVersion version that Tanzu CLI should use if available.
// The value should be a valid semantic version as defined in
// https://semver.org/.
func (po *plugin) GetRecommendedVersion() string {
	return po.RecommendedVersion
}

// Fetch the binary for a plugin version.
func (po *plugin) Fetch(version, os, arch string) ([]byte, error) {
	artifactList, exists := po.Artifacts[version]
	if !exists {
		return nil, errors.Errorf("unable to find requested version '%v' of plugin", version)
	}

	for i := range artifactList {
		if artifactList[i].OS == os && artifactList[i].Arch == arch {
			switch artifactList[i].Type {
			case "OCIImage":
				return artifacts.NewOCIArtifact(string(artifactList[i].Image)).Fetch()
			case "GCP":
				return artifacts.NewGCPArtifact(string(artifactList[i].GCP), string(artifactList[i].GCP)).Fetch()
			case "local":
				return artifacts.NewLocalArtifact(string(artifactList[i].Local)).Fetch()
			}
		}
	}
	return nil, errors.Errorf("unable to find requested artifact for '%v-%v' of plugin", os, arch)
}

// SetDiscovery specificies the name of the discovery from where
// this plugin is discovered.
func (po *plugin) SetDiscovery(discovery string) {
	po.Discovery = discovery
}

// SetScope specificies the scope of the plugin. Stand-Alone or Context
func (po *plugin) SetScope(scope string) {
	po.Scope = scope
}

// SetStatus specificies the current plugin installation status
func (po *plugin) SetStatus(status string) {
	po.Status = status
}
