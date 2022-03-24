// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"

	"sigs.k8s.io/yaml"

	clitest "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cli"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

var (
	createManagementClusterTest                *clitest.Test
	deleteManagementCusterTest                 *clitest.Test
	managementClusterUnAvailableSubCommandTest *clitest.Test
	mcCeipSetSilent                            *clitest.Test
	mcCreateNonExistsFlag                      *clitest.Test
)

const (
	usage = "Usage:"
)

func initManagementCluster() {
	mcConfigFile, err := os.CreateTemp("", tconf.ManagementClusterName)
	if err != nil {
		log.Fatal(err)
	}
	managementClusterUnAvailableSubCommandTest = clitest.NewTest("create management-cluster", "management-cluster UnAvailableSubCommand", FuncToExecAndValidateStdError)
	mcCeipSetSilent = clitest.NewTest("management-cluster ceip-participation set", "management-cluster ceip-participation set", FuncToExecAndValidateStdError)
	mcCreateNonExistsFlag = clitest.NewTest("management-cluster create -nonExistsFlag", "management-cluster create -nonExistsFlag", FuncToExecAndValidateStdError)

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
