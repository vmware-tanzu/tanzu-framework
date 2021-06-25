// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackagedatamodel

import "time"

// PackageInstalledOptions includes fields for package list
type PackageInstalledOptions struct {
	Namespace            string
	PackageName          string
	KubeConfig           string
	ClusterRoleName      string
	PkgInstallName       string
	ServiceAccountName   string
	ValuesFile           string
	Version              string
	PollInterval         time.Duration
	PollTimeout          time.Duration
	AllNamespaces        bool
	ListInstalled        bool
	CreateNamespace      bool
	CreateSecret         bool
	CreateServiceAccount bool
	Install              bool
	Wait                 bool
}

// NewPackageInstalledOptions instantiates PackageListOptions
func NewPackageInstalledOptions() *PackageInstalledOptions {
	return &PackageInstalledOptions{}
}

// PackageProgress channels for sending messages
type PackageProgress struct {
	// Use buffered chan so that sending goroutine doesn't block
	ProgressMsg chan string
	// Empty struct for indicating that goroutine is finished
	Done chan struct{}
	// Err chan for reporting errors
	Err chan error
}
