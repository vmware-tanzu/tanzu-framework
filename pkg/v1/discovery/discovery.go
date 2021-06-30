// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// NewClusterQueryClientForConfig returns a new cluster query builder for a REST config.
func NewClusterQueryClientForConfig(config *rest.Config) (*ClusterQueryClient, error) {
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return NewClusterQueryClient(dynamicClient, discoveryClient)
}

// NewClusterQueryClient returns a new cluster query builder
func NewClusterQueryClient(dynamicClient dynamic.Interface, discoveryClient discovery.DiscoveryInterface) (*ClusterQueryClient, error) {
	config := &clusterQueryClientConfig{
		dynamicClient:      dynamicClient,
		discoveryClientset: discoveryClient,
	}

	return &ClusterQueryClient{
		config: config,
	}, nil
}

type clusterQueryClientConfig struct {
	dynamicClient      dynamic.Interface
	discoveryClientset discovery.DiscoveryInterface
}

// ClusterQueryClient allows clients to inspect the cluster objects, GVK and schema state of a cluster
type ClusterQueryClient struct {
	config *clusterQueryClientConfig
}

// Query provides a new query object to prepare
func (c *ClusterQueryClient) Query(targets ...QueryTarget) *ClusterQuery {
	return &ClusterQuery{
		targets: targets,
		config:  c.config,
	}
}

// PreparedQuery provides a prepared object
func (c *ClusterQueryClient) PreparedQuery(targets ...QueryTarget) func() (bool, error) {
	q := &ClusterQuery{
		targets: targets,
		config:  c.config,
	}
	return q.Prepare()
}

// QueryTarget implementations: Resource, GVK, Schema
type QueryTarget interface {
	Run(config *clusterQueryClientConfig) (bool, error)
	Reason() string
}

// ClusterQuery provides a means of executing a queries targets to determine results
type ClusterQuery struct {
	queryFailures []string
	targets       []QueryTarget
	config        *clusterQueryClientConfig
}

// Execute actually executes the function
// Normally this function is returned by Prepare() and stored as a constant to re-use
func (c *ClusterQuery) Execute() (bool, error) {
	for _, t := range c.targets {
		ok, err := t.Run(c.config)
		if err != nil {
			return false, err
		}
		if !ok {
			c.queryFailures = append(c.queryFailures, t.Reason())
			continue
		}
	}

	return len(c.queryFailures) == 0, nil
}

// Prepare queries for the discovery API on the resources, GVKs and/or partial schema a cluster has.
func (c *ClusterQuery) Prepare() func() (bool, error) {
	return c.Execute
}

// QueryFailures returns all of the queries failures
func (c *ClusterQuery) QueryFailures() []string {
	return c.queryFailures
}
