// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

// CarvelTools defines the carvel tools used for building packages
type CarvelTools struct {
	Tools []Tool `yaml:"tools"`
}

// Tool is the definition of the carvel tool
type Tool struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Url     string `yaml:"url"`
}
