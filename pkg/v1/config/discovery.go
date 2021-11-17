// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/aunum/log"

	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/common"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/discovery"
)

func populateDefaultStandaloneDiscovery(c *configv1alpha1.ClientConfig) bool {
	if c.ClientOptions == nil {
		c.ClientOptions = &configv1alpha1.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configv1alpha1.CLIOptions{}
	}
	if c.ClientOptions.CLI.DiscoverySources == nil {
		c.ClientOptions.CLI.DiscoverySources = make([]configv1alpha1.PluginDiscovery, 0)
	}

	defaultDiscovery := getDefaultStandaloneDiscoverySource(GetDefaultStandaloneDiscoveryType())
	if defaultDiscovery == nil {
		return false
	}

	matchIdx := findDiscoverySourceIndex(c.ClientOptions.CLI.DiscoverySources, func(pd configv1alpha1.PluginDiscovery) bool {
		return discovery.CheckDiscoveryName(pd, DefaultStandaloneDiscoveryName) ||
			discovery.CheckDiscoveryName(pd, DefaultStandaloneDiscoveryNameLocal)
	})

	if matchIdx >= 0 {
		if discovery.CompareDiscoverySource(c.ClientOptions.CLI.DiscoverySources[matchIdx], *defaultDiscovery, GetDefaultStandaloneDiscoveryType()) {
			return false
		}
		c.ClientOptions.CLI.DiscoverySources[matchIdx] = *defaultDiscovery
		return true
	}

	// Prepend default discovery to available discovery sources
	c.ClientOptions.CLI.DiscoverySources = append([]configv1alpha1.PluginDiscovery{*defaultDiscovery}, c.ClientOptions.CLI.DiscoverySources...)
	return true
}

func findDiscoverySourceIndex(discoverySources []configv1alpha1.PluginDiscovery, matcherFunc func(pd configv1alpha1.PluginDiscovery) bool) int {
	for i := range discoverySources {
		if matcherFunc(discoverySources[i]) {
			return i
		}
	}
	return -1 // haven't found a match
}

func getDefaultStandaloneDiscoverySource(dsType string) *configv1alpha1.PluginDiscovery {
	switch dsType {
	case common.DiscoveryTypeLocal:
		return getDefaultStandaloneDiscoverySourceLocal()
	case common.DiscoveryTypeOCI:
		return getDefaultStandaloneDiscoverySourceOCI()
	}
	log.Warning("unsupported default standalone discovery configuration")
	return nil
}

func getDefaultStandaloneDiscoverySourceOCI() *configv1alpha1.PluginDiscovery {
	return &configv1alpha1.PluginDiscovery{
		OCI: &configv1alpha1.OCIDiscovery{
			Name:  DefaultStandaloneDiscoveryName,
			Image: GetDefaultStandaloneDiscoveryImage(),
		},
	}
}

func getDefaultStandaloneDiscoverySourceLocal() *configv1alpha1.PluginDiscovery {
	return &configv1alpha1.PluginDiscovery{
		Local: &configv1alpha1.LocalDiscovery{
			Name: DefaultStandaloneDiscoveryNameLocal,
			Path: GetDefaultStandaloneDiscoveryLocalPath(),
		},
	}
}
