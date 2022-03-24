// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	clitest "github.com/vmware-tanzu/tanzu-framework/pkg/v1/test/cli"
)

var (
	createClusterNegTestSilentMode        *clitest.Test
	deleteClusterNegTestSilentMode        *clitest.Test
	availableUpgradesGetNegTestSilentMode *clitest.Test
	credUpdateNegTestSilentMode           *clitest.Test
	clusterGetTestSilentMode              *clitest.Test
	clusterKubeconfGetSilentMode          *clitest.Test
	clusterMHGet                          *clitest.Test
	clusterNodePoolSetSilentMode          *clitest.Test
	clusterNodePoolListSilentMode         *clitest.Test
	clusterNodePoolDeleteSilentMode       *clitest.Test
	clusterScaleSilentMode                *clitest.Test
	clusterUpgradeSilentMode              *clitest.Test
)

var FuncToExecAndValidateStdError = func(t *clitest.Test) error {
	if err := t.ExecNotContainsStdErrorString(usage); err != nil {
		return err
	}
	return nil
}

func initClusterSilentModeUsecases() {
	createClusterNegTestSilentMode = clitest.NewTest("create cluster", "cluster create", FuncToExecAndValidateStdError)
	deleteClusterNegTestSilentMode = clitest.NewTest("delete cluster", "cluster delete", FuncToExecAndValidateStdError)
	availableUpgradesGetNegTestSilentMode = clitest.NewTest("cluster available-upgrades get", "cluster available-upgrades get", FuncToExecAndValidateStdError)
	clusterGetTestSilentMode = clitest.NewTest("cluster get", "cluster get", FuncToExecAndValidateStdError)
	credUpdateNegTestSilentMode = clitest.NewTest("cluster credentials update", "cluster credentials update", FuncToExecAndValidateStdError)
	clusterKubeconfGetSilentMode = clitest.NewTest("cluster kubeconfig get", "cluster kubeconfig get", FuncToExecAndValidateStdError)
	clusterMHGet = clitest.NewTest("cluster machinehealthcheck node get", "cluster machinehealthcheck node get", FuncToExecAndValidateStdError)
	clusterNodePoolSetSilentMode = clitest.NewTest("cluster node-pool set", "cluster node-pool set", FuncToExecAndValidateStdError)
	clusterNodePoolListSilentMode = clitest.NewTest("cluster node-pool list", "cluster node-pool list", FuncToExecAndValidateStdError)
	clusterNodePoolDeleteSilentMode = clitest.NewTest("cluster node-pool delete", "cluster node-pool delete", FuncToExecAndValidateStdError)
	clusterScaleSilentMode = clitest.NewTest("cluster scale", "cluster scale", FuncToExecAndValidateStdError)
	clusterUpgradeSilentMode = clitest.NewTest("cluster upgrade", "cluster upgrade", FuncToExecAndValidateStdError)
}
