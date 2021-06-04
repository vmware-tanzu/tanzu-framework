// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"strings"

	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
)

var availableUpgradesTest *clitest.Test

func initAvailableUpgrades() {
	clusterAvailableUpgradesCommand := fmt.Sprintf("cluster available-upgrades get %s", clusterName)

	availableUpgradesTest = clitest.NewTest("available-upgrades for cluster", clusterAvailableUpgradesCommand, func(t *clitest.Test) error {
		fmt.Printf("available upgrade test : clustername: %s\n", clusterName)
		fmt.Printf("available upgrade test : command : %s\n", t.Command)
		if err := t.Exec(); err != nil {
			return err
		}
		stdOut := t.StdOut()
		so := stdOut.String()
		// TODO: update the test to deterministically expects the available upgrades
		// If there are no available upgrades return success else check if the required columns are present
		if strings.Contains(so, "no available upgrades for cluster") {
			return nil
		}
		if !strings.Contains(so, "NAME") {
			return fmt.Errorf("available upgrade list doesn't contain  NAME column, got : %v", so)
		}
		if !strings.Contains(so, "VERSION") {
			return fmt.Errorf("available upgrade list doesn't contain  VERSION column, got: %v", so)
		}
		if !strings.Contains(so, "COMPATIBLE") {
			return fmt.Errorf("available upgrade list doesn't contain  COMPATIBLE column, got: %v", so)
		}
		return nil
	})
}
