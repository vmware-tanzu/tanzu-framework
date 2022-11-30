// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package publish implements plugin and plugin api publishing related function
package publish

import (
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
)

// Publisher is an interface to publish plugin and CLIPlugin resource files to discovery
type Publisher interface {
	// PublishPlugin publishes plugin binaries to distribution
	PublishPlugin(version, os, arch, plugin, sourcePath string) (string, error)
	// PublishDiscovery publishes the CLIPlugin resources YAML to a discovery
	PublishDiscovery() error
	// Type returns type of publisher (local, oci)
	Type() string
}

// Metadata defines metadata required for plugins publishing
type Metadata struct {
	Plugins            []string
	InputArtifactDir   string
	OSArch             []string
	RecommendedVersion string
	LocalDiscoveryPath string
	PublisherInterface Publisher
}

// PublishPlugins publishes the plugin based on provided metadata
// This function is responsible for auto-detecting the available plugin versions
// as well as os-arch and publishing artifacts to correct discovery and distribution
// based on the publisher type
func PublishPlugins(pm *Metadata) error {
	_ = ensureResourceDir(pm.LocalDiscoveryPath, true)

	availablePluginInfo, err := detectAvailablePluginInfo(pm.InputArtifactDir, pm.Plugins, pm.OSArch, pm.RecommendedVersion)
	if err != nil {
		return err
	}

	for plugin, pluginInfo := range availablePluginInfo {

		mapVersionArtifactList := make(map[string]v1alpha1.ArtifactList)

		// Create version based artifact list
		for version, arrOSArch := range pluginInfo.versions {
			artifacts := make([]v1alpha1.Artifact, 0)
			for _, oa := range arrOSArch {
				sourcePath, digest, err := getPluginPathAndDigestFromMetadata(pm.InputArtifactDir, plugin, version, oa.os, oa.arch)
				if err != nil {
					return err
				}

				destPath, err := pm.PublisherInterface.PublishPlugin(sourcePath, version, oa.os, oa.arch, plugin)
				if err != nil {
					return err
				}

				artifacts = append(artifacts, newArtifactObject(oa.os, oa.arch, pm.PublisherInterface.Type(), digest, destPath))
			}
			mapVersionArtifactList[version] = artifacts
		}

		// Create new CLIPlugin resource based on plugin and artifact info
		cliPlugin := newCLIPluginResource(plugin, pluginInfo.target, pluginInfo.description, pluginInfo.recommendedVersion, mapVersionArtifactList)

		err := saveCLIPluginResource(&cliPlugin, filepath.Join(pm.LocalDiscoveryPath, plugin+".yaml"))
		if err != nil {
			return errors.Wrap(err, "could not write CLIPlugin to file")
		}
	}

	return pm.PublisherInterface.PublishDiscovery()
}

func getPluginNameAndTarget(pluginTarget string) (string, string) {
	arr := strings.Split(pluginTarget, ":")
	plugin := arr[0]
	target := ""
	if len(arr) > 1 {
		target = arr[1]
	}
	return plugin, target
}
