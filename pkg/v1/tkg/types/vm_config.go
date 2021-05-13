// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package types

// VSphereVirtualMachine struct to hold vSphere VM object
type VSphereVirtualMachine struct {
	Name          string
	Moid          string
	OVAVersion    string
	DistroName    string
	DistroVersion string
	DistroArch    string
	IsTemplate    bool
}
