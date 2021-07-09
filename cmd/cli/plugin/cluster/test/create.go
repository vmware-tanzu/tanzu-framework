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
	createClusterCommand string
	clusterName          string
	plan                 string
	createClusterTest    *clitest.Test
)

const (
	devPlanName = "dev"
)

func initCreate() {
	clusterName = tconf.ClusterPrefix + clitest.GenerateName()
	plan = devPlanName

	configFile, err := os.CreateTemp("", clusterName)
	if err != nil {
		log.Fatal(err)
	}

	createClusterCommand = fmt.Sprintf("cluster create -v3 -f %s", configFile.Name())
	createClusterTest = clitest.NewTest("create cluster", createClusterCommand, func(t *clitest.Test) error {
		defer os.Remove(configFile.Name())

		configVars := make(map[string]string)
		configVars[constants.ConfigVariableClusterName] = clusterName
		configVars[constants.ConfigVariableClusterPlan] = plan
		out, err := yaml.Marshal(configVars)
		if err != nil {
			return err
		}

		err = os.WriteFile(configFile.Name(), out, 0644)
		if err != nil {
			return err
		}

		err = t.ExecContainsErrorString(fmt.Sprintf("Workload cluster '%s' created", clusterName))
		if err != nil {
			return err
		}

		return nil
	})
}
