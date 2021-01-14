// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/aunum/log"

	"github.com/vmware-tanzu-private/core/pkg/v1/cli/command/core"
)

func main() {
	if err := core.Execute(); err != nil {
		log.Fatal(err)
	}
}
