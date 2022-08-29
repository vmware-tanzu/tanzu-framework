// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package helper implements helper functions used for unit tests
package helper

import (
	"github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ###################### Fake CAPI objects creation helper ######################

// NewCLIPlugin returns new NewCLIPlugin object
func NewCLIPlugin(options TestCLIPluginOption) v1alpha1.CLIPlugin {
	artifacts := []v1alpha1.Artifact{
		{
			Image: "fake.image.repo.com/tkg/plugin/test-darwin-plugin:v1.4.0",
			OS:    "darwin",
			Arch:  "amd64",
		},
		{
			Image: "fake.image.repo.com/tkg/plugin/test-linux-plugin:v1.4.0",
			OS:    "linux",
			Arch:  "amd64",
		},
		{
			Image: "fake.image.repo.com/tkg/plugin/test-windows-plugin:v1.4.0",
			OS:    "windows",
			Arch:  "amd64",
		},
	}
	cliplugin := v1alpha1.CLIPlugin{
		ObjectMeta: metav1.ObjectMeta{
			Name: options.Name,
		},
		Spec: v1alpha1.CLIPluginSpec{
			Description:        options.Description,
			RecommendedVersion: options.RecommendedVersion,
			Artifacts: map[string]v1alpha1.ArtifactList{
				"v1.0.0": artifacts,
			},
		},
	}
	return cliplugin
}
