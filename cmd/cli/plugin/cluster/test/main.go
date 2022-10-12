// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/plugin"
	clitest "github.com/vmware-tanzu/tanzu-framework/cli/runtime/test"
)

const (
	defaultClusterPrefix          = "tanzu-cli-cluster-"
	defaultInfrastructureProvider = "docker"
)

type testConfig struct {
	InfrastructureName    string `json:"infrastructure_name,omitempty"`
	UseExistingCluster    bool   `json:"use_existing_cluster,omitempty"`
	ManagementClusterName string `json:"management_cluster_name,omitempty"`
	ClusterPrefix         string `json:"cluster_prefix,omitempty"`
}

var tconf *testConfig
var descriptor = clitest.NewTestFor("cluster")

var _ = func() error {
	tconf = &testConfig{}
	var testConfigData []byte
	var err error

	if testConfigPath, ok := os.LookupEnv("CLUSTER_TEST_CONFIG"); ok {
		log.Printf("Reading test config from file %s\n", testConfigPath)

		if testConfigData, err = os.ReadFile(testConfigPath); err != nil {
			log.Fatal(err)
		}

		if err = yaml.Unmarshal(testConfigData, tconf); err != nil {
			log.Fatal(err)
		}
	}

	tconf.defaults()

	return err
}()

func (c *testConfig) defaults() {
	if c.ClusterPrefix == "" {
		c.ClusterPrefix = defaultClusterPrefix
	}

	if c.InfrastructureName == "" {
		c.InfrastructureName = defaultInfrastructureProvider
	}

	if c.ManagementClusterName == "" {
		c.ManagementClusterName = c.ClusterPrefix + "mc"
	}
}

func init() {
	// Init the tests as per the dependency
	initManagementClusterSilentUsage()
	initManagementCluster()
	initClusterSilentModeUsecases()
	initClusterCreate()
	initAvailableUpgrades()
	initClusterDelete()
}

func main() {
	p, err := plugin.NewPlugin(descriptor)
	if err != nil {
		log.Fatal(err)
	}

	p.Cmd.RunE = test
	if err := p.Execute(); err != nil {
		os.Exit(1)
	}
}

func test(c *cobra.Command, _ []string) error {
	m := clitest.NewMain("cluster", c, Cleanup)
	defer m.Finish()

	err := silenceUsageMCTestCases(m)
	if err != nil {
		return err
	}
	err = silenceUsageClusterTestCases(m)
	if err != nil {
		return err
	}
	// create management-cluster
	if !tconf.UseExistingCluster {
		m.AddTest(createManagementClusterTest)
		if err := createManagementClusterTest.Run(); err != nil {
			return err
		}
	}

	// create workload cluster
	m.AddTest(createClusterTest)
	if err := createClusterTest.Run(); err != nil {
		return err
	}

	// list workload clusters
	m.AddTest(listClusterTest)
	if err := listClusterTest.Run(); err != nil {
		return err
	}

	// get the available upgrades for cluster
	m.AddTest(availableUpgradesTest)
	if err := availableUpgradesTest.Run(); err != nil {
		return err
	}

	// delete workload cluster
	if err := deleteClusterTest.Run(); err != nil {
		return err
	}

	// delete management cluster
	if !tconf.UseExistingCluster {
		if err := deleteManagementCusterTest.Run(); err != nil {
			return err
		}
	}

	return nil
}

func silenceUsageMCTestCases(m *clitest.Main) error {
	// Test use case# management-cluster : SilenceUsage: true : should not print Usage info
	m.AddTest(mcUnAvailCmdSilentUsage)
	if err := mcUnAvailCmdSilentUsage.Run(); err != nil {
		return err
	}

	// Test use case# management-cluster ceip-participation set : SilenceUsage: true : should not print Usage: information
	m.AddTest(mcCeipSetSilentUsage)
	if err := mcCeipSetSilentUsage.Run(); err != nil {
		return err
	}

	// Test use case# management-cluster create -nonExistsFlag : SilenceUsage: true : should not print Usage: information
	m.AddTest(mcCreateNonExistsFlagSilentUsage)
	if err := mcCreateNonExistsFlagSilentUsage.Run(); err != nil {
		return err
	}

	// Test use case# management-cluster create :  SilenceUsage: true : should not print Usage: information
	m.AddTest(mcCreateSilentUage)
	if err := mcCreateSilentUage.Run(); err != nil {
		return err
	}
	return nil
}

func silenceUsageClusterTestCases(m *clitest.Main) error {
	// Test use case# cluster create : SilenceUsage: true : should not print Usage info
	m.AddTest(createClusterNegTestSilentMode)
	if err := createClusterNegTestSilentMode.Run(); err != nil {
		return err
	}

	// Test use case# cluster delete : SilenceUsage: true : should not print Usage info
	m.AddTest(deleteClusterNegTestSilentMode)
	if err := deleteClusterNegTestSilentMode.Run(); err != nil {
		return err
	}

	// Test use case# cluster available-upgrades get  : SilenceUsage: true : should not print Usage info
	m.AddTest(availableUpgradesGetNegTestSilentMode)
	if err := availableUpgradesGetNegTestSilentMode.Run(); err != nil {
		return err
	}

	// Test use case# cluster available-upgrades get  : SilenceUsage: true : should not print Usage info
	m.AddTest(credUpdateNegTestSilentMode)
	if err := credUpdateNegTestSilentMode.Run(); err != nil {
		return err
	}
	// Test use case: cluster get : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterGetTestSilentMode)
	if err := clusterGetTestSilentMode.Run(); err != nil {
		return err
	}
	// Test use case: cluster kubeconfig get : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterKubeconfGetSilentMode)
	if err := clusterKubeconfGetSilentMode.Run(); err != nil {
		return err
	}

	// Test use case: cluster machinehealthcheck node get : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterMHGet)
	if err := clusterMHGet.Run(); err != nil {
		return err
	}

	// Test use case: cluster node-pool set : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterNodePoolSetSilentMode)
	if err := clusterNodePoolSetSilentMode.Run(); err != nil {
		return err
	}

	// Test use case: cluster node-pool list : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterNodePoolListSilentMode)
	if err := clusterNodePoolListSilentMode.Run(); err != nil {
		return err
	}

	// Test use case: cluster node-pool delete : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterNodePoolDeleteSilentMode)
	if err := clusterNodePoolDeleteSilentMode.Run(); err != nil {
		return err
	}

	// Test use case: cluster scale : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterScaleSilentMode)
	if err := clusterScaleSilentMode.Run(); err != nil {
		return err
	}

	// Test use case: cluster upgrade : SilenceUsage : true : should not print Usage info
	m.AddTest(clusterUpgradeSilentMode)
	if err := clusterUpgradeSilentMode.Run(); err != nil {
		return err
	}
	return nil
}

// Cleanup the test.
func Cleanup() error {
	return nil
}
