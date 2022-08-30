// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"sigs.k8s.io/yaml"

	clitest "github.com/vmware-tanzu/tanzu-framework/cli/runtime/test"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

var (
	createManagementClusterTest      *clitest.Test
	deleteManagementCusterTest       *clitest.Test
	mcUnAvailCmdSilentUsage          *clitest.Test
	mcCeipSetSilentUsage             *clitest.Test
	mcCreateNonExistsFlagSilentUsage *clitest.Test
	mcCreateSilentUage               *clitest.Test
)

const (
	usage = "Usage:"
)

func initManagementCluster() {
	mcConfigFile, err := os.CreateTemp("", tconf.ManagementClusterName)
	if err != nil {
		log.Fatal(err)
	}

	createMcCommand := fmt.Sprintf("management-cluster create -v3 -f %s", mcConfigFile.Name())
	createManagementClusterTest = clitest.NewTest("create management-cluster", createMcCommand, func(t *clitest.Test) error {
		defer os.Remove(mcConfigFile.Name())

		configVars := make(map[string]string)
		configVars[constants.ConfigVariableClusterName] = tconf.ManagementClusterName
		configVars[constants.ConfigVariableClusterPlan] = "dev"
		configVars[constants.ConfigVariableInfraProvider] = tconf.InfrastructureName
		configVars[constants.ConfigVariableCNI] = "calico"

		out, err := yaml.Marshal(configVars)
		if err != nil {
			return err
		}
		if err := os.WriteFile(mcConfigFile.Name(), out, 0644); err != nil {
			return err
		}
		if err := t.ExecContainsErrorString("Management cluster created!"); err != nil {
			return err
		}
		return nil
	})

	deleteManagementCusterTest = clitest.NewTest("delete management-cluster", "management-cluster delete -v3 -y --force", func(t *clitest.Test) error {
		if err := t.Exec(); err != nil {
			return err
		}
		return nil
	})
}

// initManagementClusterSilentUsage has test definitions to test silentUsage use cases for the command 'management-cluster' and its sub-commands.
// all test cases in this function, tests that the commands out put should not print "Usage:" info as "silentUSage" set true for command 'management-cluster' and its sub-commands.
func initManagementClusterSilentUsage() {
	mcUnAvailCmdSilentUsage = clitest.NewTest("create management-cluster", "management-cluster UnAvailableSubCommand", FuncToExecAndValidateStdError)
	mcCeipSetSilentUsage = clitest.NewTest("management-cluster ceip-participation set", "management-cluster ceip-participation set", FuncToExecAndValidateStdError)
	mcCreateNonExistsFlagSilentUsage = clitest.NewTest("management-cluster create -nonExistsFlag", "management-cluster create -nonExistsFlag", FuncToExecAndValidateStdError)
	mcCreateSilentUage = clitest.NewTest("management-cluster create", "management-cluster create", FuncToExecAndValidateStdError)
}
