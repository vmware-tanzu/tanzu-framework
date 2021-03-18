// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cli

// cmdGroup is a group of CLI commands.
type cmdGroup string

const (
	// RunCmdGroup are commands associated with Tanzu Run.
	RunCmdGroup cmdGroup = "Run"

	// ManageCmdGroup are commands associated with Tanzu Manage.
	ManageCmdGroup cmdGroup = "Manage"

	// BuildCmdGroup are commands associated with Tanzu Build.
	BuildCmdGroup cmdGroup = "Build"

	// ObserveCmdGroup are commands associated with Tanzu Observe.
	ObserveCmdGroup cmdGroup = "Observe"

	// SystemCmdGroup are system commands.
	SystemCmdGroup cmdGroup = "System"

	// VersionCmdGroup are version commands.
	VersionCmdGroup cmdGroup = "Version"

	// AdminCmdGroup are admin commands.
	AdminCmdGroup cmdGroup = "Admin"

	// TestCmdGroup is the test command group.
	TestCmdGroup cmdGroup = "Test"

	// ExtraCmdGroup is the extra command group.
	ExtraCmdGroup cmdGroup = "Extra"
)
