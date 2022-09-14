// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package constants

const (
	// TanzuCLISystemNamespace  is the namespace for tanzu cli resources
	TanzuCLISystemNamespace = "tanzu-cli-system"

	// CLIPluginImageRepositoryOverrideLabel is the label on the configmap which specifies CLIPlugin image repository override
	CLIPluginImageRepositoryOverrideLabel = "cli.tanzu.vmware.com/cliplugin-image-repository-override"

	// DefaultQPS is the default maximum query per second for the rest config
	DefaultQPS = 200

	// DefaultBurst is the default maximum burst for throttle for the rest config
	DefaultBurst = 200
)
