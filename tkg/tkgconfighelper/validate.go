// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgconfighelper provides various helpers and utilities
package tkgconfighelper

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/util/version"
)

// ManagementClusterVersionToK8sVersionSupportMatrix defines the support matrix of which k8s version
// are supported based on management cluster version
var ManagementClusterVersionToK8sVersionSupportMatrix = map[string][]string{
	"v1.0": {"v1.17"},
	"v1.1": {"v1.17", "v1.18"},
	"v1.2": {"v1.17", "v1.18", "v1.19"},
	"v1.3": {"v1.17", "v1.18", "v1.19", "v1.20"},
	"v1.4": {"v1.17", "v1.18", "v1.19", "v1.20", "v1.21"},
	"v1.5": {"v1.19", "v1.20", "v1.21", "v1.22"},
	"v1.6": {"v1.20", "v1.21", "v1.22", "v1.23"},
	"v1.7": {"v1.21", "v1.22", "v1.23", "v1.24"},
	"v2.1": {"v1.21", "v1.22", "v1.23", "v1.24"},
}

// ValidateK8sVersionSupport validates the k8s version is supported on management cluster or not
func ValidateK8sVersionSupport(mgmtClusterTkgVersion, kubernetesVersion string) error {
	mgmtClusterSemVersion, err := version.ParseSemantic(mgmtClusterTkgVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to parse management cluster version %s", mgmtClusterTkgVersion)
	}

	k8sSemVersion, err := version.ParseSemantic(kubernetesVersion)
	if err != nil {
		return errors.Wrapf(err, "unable to parse kubernetes version %s", kubernetesVersion)
	}

	mgmtClusterMajorMinorVersion := fmt.Sprintf("v%v.%v", mgmtClusterSemVersion.Major(), mgmtClusterSemVersion.Minor())
	k8sMajorMinorVersion := fmt.Sprintf("v%v.%v", k8sSemVersion.Major(), k8sSemVersion.Minor())

	supportedK8sVersions, exists := ManagementClusterVersionToK8sVersionSupportMatrix[mgmtClusterMajorMinorVersion]
	if !exists {
		supportedManagementClusterVersions := []string{}
		for mcVersion := range ManagementClusterVersionToK8sVersionSupportMatrix {
			supportedManagementClusterVersions = append(supportedManagementClusterVersions, mcVersion)
		}
		sort.Strings(supportedManagementClusterVersions)
		return errors.Errorf("only %v management cluster versions are supported with current version of TKG CLI. Please upgrade TKG CLI to latest version if you are using it on latest version of management cluster.", supportedManagementClusterVersions)
	}

	for _, supportedK8sVersion := range supportedK8sVersions {
		if supportedK8sVersion == k8sMajorMinorVersion {
			// if k8sMajorMinorVersion matches with the supported supportedK8sVersion then return nil and
			// specified k8s version is compatible with the current management cluster
			return nil
		}
	}

	errMsg := fmt.Sprintf("kubernetes version %s is not supported on current %v management cluster. ", kubernetesVersion, mgmtClusterTkgVersion)

	// if current CLI version supports specified k8s version after upgrading management cluster show the below error message as well.
	if mgmtClusterMajorMinorVersion != "v1.1" && ValidateK8sVersionSupport("v1.1.0", kubernetesVersion) == nil {
		errMsg += "Please upgrade management cluster if you are trying to deploy latest version of kubernetes."
	}

	return errors.New(errMsg)
}
