// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

import "github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"

// Variables defined in this file are deprecated in favor of their equivalents defined in a dedicated package.

var (
	// BuildDate is the date the CLI was built.
	// Deprecated: use github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Date
	BuildDate = buildinfo.Date

	// BuildSHA is the git sha the CLI was built with.
	// Deprecated: use github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.SHA
	BuildSHA = buildinfo.SHA

	// BuildVersion is the version the CLI was built with.
	// Deprecated: use github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.Version
	BuildVersion = buildinfo.Version
)
