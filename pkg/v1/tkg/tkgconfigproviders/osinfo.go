// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfigproviders

// OsInfo struct defining os name, version and arch properties of VM image
type OsInfo struct {
	Name    string `yaml:"OS_NAME"`
	Version string `yaml:"OS_VERSION"`
	Arch    string `yaml:"OS_ARCH"`
}
