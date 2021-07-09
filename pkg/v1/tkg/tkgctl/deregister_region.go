// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

// DeregisterFromTMCOptions Deregister from TMC options
type DeregisterFromTMCOptions struct {
	ClusterName string
	SkipPrompt  bool
}

// DeregisterFromTmc deregister management cluster from TMC
func (t *tkgctl) DeregisterFromTmc(options DeregisterFromTMCOptions) error {
	return t.tkgClient.DeRegisterManagementClusterFromTmc(options.ClusterName)
}
