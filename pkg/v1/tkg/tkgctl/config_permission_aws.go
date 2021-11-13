// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

// CreateAWSCloudFormationStack create aws cloud formation stack
func (t *tkgctl) CreateAWSCloudFormationStack(clusterConfigFile string) error {
	if _, err := t.ensureClusterConfigFile(clusterConfigFile); err != nil {
		return err
	}
	return t.tkgClient.CreateAWSCloudFormationStack()
}

// CreateAWSCloudFormationStack create aws cloud formation stack
func (t *tkgctl) GenerateAWSCloudFormationTemplate(clusterConfigFile string) (string, error) {
	if _, err := t.ensureClusterConfigFile(clusterConfigFile); err != nil {
		return "", err
	}
	return t.tkgClient.GenerateAWSCloudFormationTemplate()
}
