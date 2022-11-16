// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"gopkg.in/yaml.v3"
)

// OracleCPIDataValues serializes the CPIConfig CR
type OracleCPIDataValues struct {
	Auth OracleCPIDataValuesAuth `yaml:"auth"`

	Compartment string `yaml:"compartment"`

	VCN string `yaml:"vcn"`

	LoadBalancer struct {
		Subnet1 string `yaml:"subnet1"`
		Subnet2 string `yaml:"subnet2"`
	} `yaml:"loadBalancer"`
}

type OracleCPIDataValuesAuth struct {
	Region      string `yaml:"region"`
	Tenancy     string `yaml:"tenancy"`
	User        string `yaml:"user"`
	Key         string `yaml:"key"`
	Fingerprint string `yaml:"fingerprint"`
	Passphrase  string `yaml:"passphrase"`
}

func (v *OracleCPIDataValues) Serialize() ([]byte, error) {
	return yaml.Marshal(v)
}
