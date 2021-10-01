// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"sync"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/restmapper"
)

// clusterQueryClientConfig holds configuration for executing queries.
type clusterQueryClientConfig struct {
	dynamicClient      dynamic.Interface
	discoveryClientset discovery.DiscoveryInterface

	// Cached RESTMapper.
	// Do not use this field directly as it may not be initialized; use restMapper() method instead.
	mapper meta.RESTMapper
	// Cached openAPISchemaHelper.
	// Do not use this field directly as it may not be initialized; use openAPISchemaHelper() method instead.
	schemaHelper *openAPISchemaHelper

	mapperOnce       sync.Once
	schemaHelperOnce sync.Once
}

// newClusterQueryClientConfig returns an instance of clusterQueryClientConfig.
func newClusterQueryClientConfig(dynamicClient dynamic.Interface, discoveryClient discovery.DiscoveryInterface) *clusterQueryClientConfig {
	config := &clusterQueryClientConfig{
		dynamicClient:      dynamicClient,
		discoveryClientset: discoveryClient,
	}
	return config
}

// restMapper lazily fetches a meta.RESTMapper.
func (c *clusterQueryClientConfig) restMapper() meta.RESTMapper {
	c.mapperOnce.Do(func() {
		c.mapper = restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(c.discoveryClientset))
	})
	return c.mapper
}

// openAPISchemaHelper lazily instantiates an openAPISchemaHelper object.
func (c *clusterQueryClientConfig) openAPISchemaHelper() *openAPISchemaHelper {
	c.schemaHelperOnce.Do(func() {
		c.schemaHelper = newOpenAPISchemaHelper(c.discoveryClientset, c.restMapper())
	})
	return c.schemaHelper
}
