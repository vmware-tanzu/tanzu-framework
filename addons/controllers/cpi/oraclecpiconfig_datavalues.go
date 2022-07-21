// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"gopkg.in/yaml.v3"
)

// OracleCPIDataValues serializes the CPIConfig CR
type OracleCPIDataValues struct {
	Compartment string `yaml:"compartment"`

	VCN string `yaml:"vcn"`

	LoadBalancer struct {
		Subnet1 string `yaml:"subnet1"`
		Subnet2 string `yaml:"subnet2"`
	} `yaml:"loadBalancer"`
}

func (v *OracleCPIDataValues) Serialize() ([]byte, error) {
	return yaml.Marshal(v)
}
