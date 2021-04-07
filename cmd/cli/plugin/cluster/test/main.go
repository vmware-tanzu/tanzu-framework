// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli"
	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/plugin"
	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/constants"
)

const (
	defaultClusterPrefix          = "tanzu-cli-cluster-"
	defaultInfrastructureProvider = "docker"
)

//
type testConfig struct {
	InfrastructureName    string `json:"infrastructure_name,omitempty"`
	UseExistingCluster    bool   `json:"use_existing_cluster,omitempty"`
	ManagementClusterName string `json:"management_cluster_name,omitempty"`
	ClusterPrefix         string `json:"cluster_prefix,omitempty"`
}

var tconf *testConfig
var descriptor = cli.NewTestFor("cluster")
var createManagementClusterTest *clitest.Test
var deleteManagementCusterTest *clitest.Test

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

	return nil
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
	mcConfigFile, err := os.CreateTemp("", tconf.ManagementClusterName)
	if err != nil {
		log.Fatal(err)
	}

	createMcCommand := fmt.Sprintf("management-cluster create -f %s", mcConfigFile.Name())
	createManagementClusterTest = clitest.NewTest("create management-cluster", createMcCommand, func(t *clitest.Test) error {
		defer os.Remove(mcConfigFile.Name())

		configVars := make(map[string]string)
		configVars[constants.ConfigVariableClusterName] = tconf.ManagementClusterName
		configVars[constants.ConfigVariableClusterPlan] = "dev"
		configVars[constants.ConfigVariableInfraProvider] = tconf.InfrastructureName
		out, err := yaml.Marshal(configVars)
		if err != nil {
			return err
		}

		if err = os.WriteFile(mcConfigFile.Name(), out, 0644); err != nil {
			return err
		}

		if err = t.ExecContainsErrorString("Management cluster created!"); err != nil {
			return err
		}

		return nil
	})

	deleteManagementCusterTest = clitest.NewTest("delete management-cluster", "management-cluster delete -y --force", func(t *clitest.Test) error {
		if err := t.Exec(); err != nil {
			return err
		}

		return nil
	})
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

// Cleanup the test.
func Cleanup() error {
	return nil
}
