// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"github.com/aunum/log"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/discovery"
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

func populateDefaultStandaloneDiscovery(c *configapi.ClientConfig) bool {
	if c.ClientOptions == nil {
		c.ClientOptions = &configapi.ClientOptions{}
	}
	if c.ClientOptions.CLI == nil {
		c.ClientOptions.CLI = &configapi.CLIOptions{}
	}
	if c.ClientOptions.CLI.DiscoverySources == nil {
		c.ClientOptions.CLI.DiscoverySources = make([]configapi.PluginDiscovery, 0)
	}

	defaultDiscovery := getDefaultStandaloneDiscoverySource(GetDefaultStandaloneDiscoveryType())
	if defaultDiscovery == nil {
		return false
	}

	matchIdx := findDiscoverySourceIndex(c.ClientOptions.CLI.DiscoverySources, func(pd configapi.PluginDiscovery) bool {
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
	c.ClientOptions.CLI.DiscoverySources = append([]configapi.PluginDiscovery{*defaultDiscovery}, c.ClientOptions.CLI.DiscoverySources...)
	return true
}

func findDiscoverySourceIndex(discoverySources []configapi.PluginDiscovery, matcherFunc func(pd configapi.PluginDiscovery) bool) int {
	for i := range discoverySources {
		if matcherFunc(discoverySources[i]) {
			return i
		}
	}
	return -1 // haven't found a match
}

func getDefaultStandaloneDiscoverySource(dsType string) *configapi.PluginDiscovery {
	switch dsType {
	case common.DiscoveryTypeLocal:
		return getDefaultStandaloneDiscoverySourceLocal()
	case common.DiscoveryTypeOCI:
		return getDefaultStandaloneDiscoverySourceOCI()
	}
	log.Warning("unsupported default standalone discovery configuration")
	return nil
}

func getDefaultStandaloneDiscoverySourceOCI() *configapi.PluginDiscovery {
	return &configapi.PluginDiscovery{
		OCI: &configapi.OCIDiscovery{
			Name:  DefaultStandaloneDiscoveryName,
			Image: GetDefaultStandaloneDiscoveryImage(),
		},
	}
}

func getDefaultStandaloneDiscoverySourceLocal() *configapi.PluginDiscovery {
	return &configapi.PluginDiscovery{
		Local: &configapi.LocalDiscovery{
			Name: DefaultStandaloneDiscoveryNameLocal,
			Path: GetDefaultStandaloneDiscoveryLocalPath(),
		},
	}
}
