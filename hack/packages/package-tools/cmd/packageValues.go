// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cmd

// PackageValues defines the packages configuration
type PackageValues struct {
	Repositories map[string]Repository `yaml:"repositories"`
}

// Repository defines package repository configuration
type Repository struct {
	Name        string      `yaml:"name"`
	Domain      string      `yaml:"domain"`
	PackageSpec PackageSpec `yaml:"packageSpec"`
	Packages    []Package   `yaml:"packages"`
	Sha256      string      `yaml:"sha256"`
	Registry    string      `yaml:"registry"`
}

// PackageSpec defines a particular package configuration
type PackageSpec struct {
	SyncPeriod string `yaml:"syncPeriod"`
	Deploy     Deploy `yaml:"deploy"`
}

// Deploy defines package deployment configuration
type Deploy struct {
	KappWaitTimeout string `yaml:"kappWaitTimeout"`
	KubeAPIQPS      string `yaml:"kubeAPIQPS"`
	KubeAPIBurst    string `yaml:"kubeAPIBurst"`
}

// Package holds the information about a package
type Package struct {
	Name                string            `yaml:"name"`
	DisplayName         string            `yaml:"displayName"`
	Path                string            `yaml:"path"`
	Domain              string            `yaml:"domain"`
	Version             string            `yaml:"version"`
	Sha256              string            `yaml:"sha256"`
	PackageSubVersion   string            `yaml:"packageSubVersion,omitempty"`
	SkipVersionOverride bool              `yaml:"skipVersionOverride,omitempty"`
	Env                 map[string]string `yaml:"env,omitempty"`
}
