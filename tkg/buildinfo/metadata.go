// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package buildinfo

import "github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo"

// IsOfficialBuild is the flag that gets set to True if it is an official build being released, it is set with the go linker's -X flag.
// Deprecated: use github.com/vmware-tanzu/tanzu-framework/pkg/v1/buildinfo.IsOfficialBuild
var IsOfficialBuild = buildinfo.IsOfficialBuild
