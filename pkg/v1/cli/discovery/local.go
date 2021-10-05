// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/distribution"
	apimachineryjson "k8s.io/apimachinery/pkg/runtime/serializer/json"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/plugin"
)

// LocalDiscovery is a artifact discovery endpoint utilizing a local host os.
type LocalDiscovery struct {
	path string
	name string
}

// NewLocalDiscovery returns a new local repository.
func NewLocalDiscovery(name, localPath string) Discovery {
	return &LocalDiscovery{
		path: localPath,
		name: name,
	}
}

// List available plugins.
func (l *LocalDiscovery) List() ([]plugin.Discovered, error) {
	plugins, err := l.Manifest()
	if err != nil {
		return nil, err
	}
	return plugins, nil
}

// Describe a plugin.
func (l *LocalDiscovery) Describe(name string) (plugin plugin.Discovered, err error) {
	plugins, err := l.Manifest()
	if err != nil {
		return
	}

	for _, p := range plugins {
		if p.Name == name {
			plugin = p
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

	items, err := ioutil.ReadDir(l.path)
	if err != nil {
		return nil, errors.Wrapf(err, "error while reading local plugin manifest directory")
	}
	for _, item := range items {
		if item.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(l.path, item.Name()))
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

		dp := DiscoveredFromK8sV1alpha1(p)
		dp.Source = l.name
		plugins = append(plugins, dp)
	}
	return plugins, nil
}

// Type of the repository.
func (l *LocalDiscovery) Type() string {
	return "local"
}

func DiscoveredFromK8sV1alpha1(p cliv1alpha1.CLIPlugin) plugin.Discovered {
	dp := plugin.Discovered{
		Name:               p.Name,
		Description:        p.Spec.Description,
		RecommendedVersion: p.Spec.RecommendedVersion,
		Optional:           p.Spec.Optional,
	}
	dp.SupportedVersions = make([]string, len(p.Spec.Artifacts))
	for v := range p.Spec.Artifacts {
		dp.SupportedVersions = append(dp.SupportedVersions, v)
	}
	dp.Distribution = distribution.ArtifactsFromK8sV1alpha1(p.Spec.Artifacts)
	return dp
}
