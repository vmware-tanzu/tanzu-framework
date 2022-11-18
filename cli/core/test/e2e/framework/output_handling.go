// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package framework

type PluginListInfo struct {
	Description string `json:"description"`
	Discovery   string `json:"discovery"`
	Name        string `json:"name"`
	Scope       string `json:"scope"`
	Status      string `json:"status"`
	Version     string `json:"version"`
}
