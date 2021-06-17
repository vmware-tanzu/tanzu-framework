// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

var repositoryCmd = &cobra.Command{
	Use:   "repository",
	Short: "Repository operations",
	Long:  `Add, list, get or delete a package repository for Tanzu packages. A package repository is a collection of packages that are grouped together into an imgpkg bundle.`,
}
