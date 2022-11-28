// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package v1

import (
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Compatibility struct {
	ManagementClusterVersions []ManagementClusterVersions `yaml:"managementClusterVersions"`
}

type ManagementClusterVersions struct {
	SupportedKubernetesVersions []string `yaml:"supportedKubernetesVersions"`
	Version                     string   `yaml:"version"`
}

func NewCompatibility(data []byte) (Compatibility, error) {
	var compatibility Compatibility
	if err := yaml.Unmarshal(data, &compatibility); err != nil {
		return Compatibility{}, errors.Wrap(err, "unable to unmarshal compatibility file data to Compatibility struct")
	}
	return compatibility, nil
}

func FilterTkgVersionByTkr(compatibility Compatibility, tkrName string) (version string, err error) {
	tkrVersion := strings.Replace(tkrName, "---", "+", 1)
	for _, managementVersions := range compatibility.ManagementClusterVersions {
		for _, supportVersion := range managementVersions.SupportedKubernetesVersions {
			if supportVersion == "" {
				return "", errors.New("invalid supportVersion")
			}
			if tkrVersion == supportVersion {
				return managementVersions.Version, nil
			}
		}
	}
	return "", errors.New("there are no compatible tkg versions")
}
