// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"fmt"

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
		results: Results{},
	}
}

// PreparedQuery provides a prepared object
func (c *ClusterQueryClient) PreparedQuery(targets ...QueryTarget) func() (bool, error) {
	q := &ClusterQuery{
		targets: targets,
		config:  c.config,
		results: Results{},
	}
	return q.Prepare()
}

// QueryTarget implementations: Resource, GVK, Schema
type QueryTarget interface {
	Name() string
	Run(config *clusterQueryClientConfig) (bool, error)
	Reason() string
}

// QueryResult is a result of a single query.
type QueryResult struct {
	// Found indicates whether the entity being checked in the query exists.
	Found bool
	// NotFoundReason indicates the reason why Found was false.
	NotFoundReason string
}

// Results is a map of query names to their corresponding QueryResult.
type Results map[string]*QueryResult

// ForQuery returns the QueryResult for a given query name.
// Return value is nil if the query target with the given name does not exist.
func (r Results) ForQuery(queryName string) *QueryResult {
	return r[queryName]
}

// ClusterQuery provides a means of executing a queries targets to determine results
type ClusterQuery struct {
	targets []QueryTarget
	config  *clusterQueryClientConfig
	results Results
}

// Execute runs all the query targets and returns true only if *all* of them succeed.
// For granular results of each query, use the Results() method after calling this method.
// Normally this function is returned by Prepare() and stored as a constant to re-use
func (c *ClusterQuery) Execute() (bool, error) {
	// Check for duplicate query names.
	m := make(map[string]struct{})
	for _, t := range c.targets {
		if _, ok := m[t.Name()]; ok {
			return false, fmt.Errorf("query target names must be unique")
		}
		m[t.Name()] = struct{}{}
	}

	success := true
	for _, t := range c.targets {
		ok, err := t.Run(c.config)
		if err != nil {
			return false, err
		}
		queryResult := &QueryResult{}
		if !ok {
			queryResult.Found = false
			queryResult.NotFoundReason = t.Reason()
			c.results[t.Name()] = queryResult
			success = false
			continue
		}
		queryResult.Found = true
		c.results[t.Name()] = queryResult
	}
	return success, nil
}

// Prepare queries for the discovery API on the resources, GVKs and/or partial schema a cluster has.
func (c *ClusterQuery) Prepare() func() (bool, error) {
	return c.Execute
}

// Results returns all of the queries failures
func (c *ClusterQuery) Results() Results {
	return c.results
}
