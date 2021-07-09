// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

// CreateAWSCloudFormationStack create aws cloud formation stack
func (t *tkgctl) CreateAWSCloudFormationStack(clusterConfigFile string) error {
	var err error
	_, err = t.ensureClusterConfigFile(clusterConfigFile)
	if err != nil {
		return err
	}

	return t.tkgClient.CreateAWSCloudFormationStack()
}
