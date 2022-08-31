// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	clitest "github.com/vmware-tanzu/tanzu-framework/cli/runtime/test"
)

var (
	deleteClusterTest *clitest.Test
)

func initClusterDelete() {
	deleteClusterCommand := fmt.Sprintf("cluster delete %s -y", clusterName)
	deleteClusterTest = clitest.NewTest("delete cluster", deleteClusterCommand, func(t *clitest.Test) error {
		if err := t.ExecContainsErrorString(fmt.Sprintf("Workload cluster '%s' is being deleted ", clusterName)); err != nil {
			return err
		}

		return nil
	})
}
