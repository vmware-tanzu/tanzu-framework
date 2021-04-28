// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"fmt"

	clitest "github.com/vmware-tanzu-private/core/pkg/v1/test/cli"
)

var listClusterTest = clitest.NewTest("list clusters", "cluster list -o json", func(t *clitest.Test) error {
	if err := t.Exec(); err != nil {
		return err
	}

	var clusters []map[string]interface{}
	stdOut := t.StdOut()
	stdOutPtr := &stdOut
	if err := json.Unmarshal(stdOutPtr.Bytes(), &clusters); err != nil {
		return err
	}

	for _, cluster := range clusters {
		name := cluster["name"].(string)
		if name == clusterName {
			return nil
		}
	}

	return fmt.Errorf("unable to get cluster %s", clusterName)
})
