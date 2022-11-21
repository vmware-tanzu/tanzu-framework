// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/distribution"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/plugin"
)

// LocalDiscovery is an artifact discovery endpoint utilizing a local host os.
type LocalDiscovery struct {
	path string
	name string
}

// NewLocalDiscovery returns a new local repository.
// If provided localPath is not an absolute path
// search under `xdg.ConfigHome/tanzu-plugin/discovery` directory
func NewLocalDiscovery(name, localPath string) Discovery {
	if !filepath.IsAbs(localPath) {
		localPath = filepath.Join(common.DefaultLocalPluginDistroDir, "discovery", localPath)
	}
	return &LocalDiscovery{
		path: localPath,
		name: name,
	}
}

// List available plugins.
func (l *LocalDiscovery) List() ([]plugin.Discovered, error) {
	return l.Manifest()
}

// Describe a plugin.
func (l *LocalDiscovery) Describe(name string) (p plugin.Discovered, err error) {
	plugins, err := l.Manifest()
	if err != nil {
		return
	}

	for i := range plugins {
		if plugins[i].Name == name {
			p = plugins[i]
			return
		}
	}
	err = errors.Errorf("cannot find plugin with name '%v'", name)
	return
}

// Name of the repository.
func (l *LocalDiscovery) Name() string {
	return l.name
}

// Manifest returns the manifest for a local repository.
func (l *LocalDiscovery) Manifest() ([]plugin.Discovered, error) {
	plugins := make([]plugin.Discovered, 0)

	items, err := os.ReadDir(l.path)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading local plugin manifest directory")
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		path := filepath.Join(l.path, item.Name())

		// ignore non yaml files
		if filepath.Ext(path) != ".yaml" && filepath.Ext(path) != ".yml" {
			continue
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return nil, errors.Wrapf(err, "error while reading manifest file")
		}

		scheme, err := cliv1alpha1.SchemeBuilder.Build()
		if err != nil {
			return nil, errors.Wrap(err, "failed to create scheme")
		}
		s := apimachineryjson.NewSerializerWithOptions(apimachineryjson.DefaultMetaFactory, scheme, scheme,
			apimachineryjson.SerializerOptions{Yaml: true, Pretty: false, Strict: false})
		var p cliv1alpha1.CLIPlugin
		_, _, err = s.Decode(b, nil, &p)
		if err != nil {
			return nil, errors.Wrap(err, "could not decode catalog file")
		}

		dp, err := DiscoveredFromK8sV1alpha1(&p)
		if err != nil {
			return nil, err
		}
		if dp.Name == "" {
			continue
		}
		dp.Source = l.name
		dp.DiscoveryType = l.Type()
		plugins = append(plugins, dp)
	}
	return plugins, nil
}

// Type of the repository.
func (l *LocalDiscovery) Type() string {
	return common.DiscoveryTypeLocal
}

// DiscoveredFromK8sV1alpha1 returns discovered plugin object from k8sV1alpha1
func DiscoveredFromK8sV1alpha1(p *cliv1alpha1.CLIPlugin) (plugin.Discovered, error) {
	dp := plugin.Discovered{
		Name:               p.Name,
		Description:        p.Spec.Description,
		RecommendedVersion: p.Spec.RecommendedVersion,
		Optional:           p.Spec.Optional,
		Target:             p.Spec.Target,
	}
	dp.SupportedVersions = make([]string, 0)
	for v := range p.Spec.Artifacts {
		dp.SupportedVersions = append(dp.SupportedVersions, v)
	}
	if err := SortVersions(dp.SupportedVersions); err != nil {
		return dp, errors.Wrapf(err, "error parsing supported versions for plugin %s", p.Name)
	}
	dp.Distribution = distribution.ArtifactsFromK8sV1alpha1(p.Spec.Artifacts)
	return dp, nil
}
