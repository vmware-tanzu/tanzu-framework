// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"github.com/spf13/cobra"
)

// oddly this is just a simple var declaration,
// the addition of sub-commands does not happen at this level.
// though the parent main() func has a plugin.AddCommands().
// instead, in the init() block of hte kubeconfig_get file we
// see an invocation of the clusterKubeconfigCmd.AddCommand() call.
// this is probably fine, but not exactly how I would expect it to be ordered.
var clusterKubeconfigCmd = &cobra.Command{
	Use:          "kubeconfig",
	Short:        "Cluster kubeconfig operations",
	SilenceUsage: true,
}
