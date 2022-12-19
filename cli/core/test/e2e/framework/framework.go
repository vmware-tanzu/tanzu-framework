// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package test defines the integration and end-to-end test case for cli core
package framework

import (
	"os"
	"path/filepath"

	"github.com/onsi/ginkgo"
)

const (
	CliCore = "[CLI-Core]"

	TanzuInit    = "tanzu init"
	TanzuVersion = "tanzu version"

	ConfigSet          = "tanzu config set "
	ConfigGet          = "tanzu config get "
	ConfigUnset        = "tanzu config unset "
	ConfigInit         = "tanzu config init"
	ConfigServerList   = "tanzu config server list"
	ConfigServerDelete = "tanzu config server delete "

	KindCreateCluster = "kind create cluster --name "
	DockerInfo        = "docker info"
	TestDir           = ".tanzu-cli-e2e"
)

// CLICoreDescribe annotates the test with the CLICore label.
func CLICoreDescribe(text string, body func()) bool {
	return ginkgo.Describe(CliCore+text, body)
}

// Framework has all helper functions to write CLI e2e test cases
type Framework struct {
	CliOps
	ConfigOps
	ClusterOps
}

func NewFramework() *Framework {
	return &Framework{
		CliOps:     NewCliOps(),
		ConfigOps:  NewConfOps(),
		ClusterOps: NewKindCluster(NewDocker()),
	}
}

func init() {
	homeDir, _ := os.UserHomeDir()
	testDirPath := filepath.Join(homeDir, TestDir)
	os.Setenv("HOME", testDirPath)
}
