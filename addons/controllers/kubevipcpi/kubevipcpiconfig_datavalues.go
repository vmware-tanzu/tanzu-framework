// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"gopkg.in/yaml.v3"
)

type KubevipCPIDataValues struct {
	LoadbalancerCIDRs    string `yaml:"loadbalancerCIDRs"`
	LoadbalancerIPRanges string `yaml:"loadbalancerIPRanges"`
}

func (v *KubevipCPIDataValues) Serialize() ([]byte, error) {
	dataValues := struct {
		DataValues KubevipCPIDataValues `yaml:"kubevipCloudProvider"`
	}{DataValues: *v}
	return yaml.Marshal(dataValues)
}
