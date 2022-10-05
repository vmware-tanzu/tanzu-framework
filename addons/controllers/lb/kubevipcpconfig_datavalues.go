// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"gopkg.in/yaml.v3"
)

type KubevipCloudProviderDataValues struct {
	LoadbalancerCIDRs    string `yaml:"loadbalancerCIDRs"`
	LoadbalancerIPRanges string `yaml:"loadbalancerIPRanges"`
}

func (v *KubevipCloudProviderDataValues) Serialize() ([]byte, error) {
	dataValues := struct {
		DataValues KubevipCloudProviderDataValues `yaml:"kubevipCloudProvider"`
	}{DataValues: *v}
	return yaml.Marshal(dataValues)
}
