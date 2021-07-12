// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/command/core"
)

func main() {
	if err := core.Execute(); err != nil {
		fmt.Println("")
		log.Fatal(err)
	}
}
