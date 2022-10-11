// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

type VCenterClusterVar struct {
	DataCenter    string `json:"datacenter"`
	Server        string `json:"server"`
	Template      string `json:"template"`
	TLSThumbprint string `json:"tlsThumbprint"`
}
