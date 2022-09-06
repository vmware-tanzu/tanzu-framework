// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

// DefaultDistro is the core set of plugins that should be included with the CLI.
var DefaultDistro = []string{"login", "pinniped-auth", "cluster", "management-cluster", "kubernetes-release", "package", "secret"}
