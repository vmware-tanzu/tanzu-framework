// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package resolver provides the implementation of the TKR resolver.
package resolver

import (
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkr/resolver/data"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkr/resolver/internal"
)

// Resolver resolves TKRs and OSImages.
type Resolver interface {
	// Resolve returns TKRs and OSImages satisfying query constraints.
	Resolve(query data.Query) data.Result
}

// Cache holds TKRs and OSImages to be used by the Resolver.
type Cache interface {
	// Add TanzuKubernetesRelease or OSImage objects to the resolver cache.
	Add(objects ...interface{})

	// Remove TanzuKubernetesRelease or OSImage objects from the resolver cache.
	Remove(objects ...interface{})

	// Get an object by name and obj type.
	Get(name string, obj interface{}) interface{}
}

// CachingResolver combines Resolver and Cache (for convenience).
type CachingResolver interface {
	Resolver
	Cache
}

// New returns a newly created instance of the TKR CachingResolver implementation. It is safe for concurrent use.
func New() CachingResolver {
	return internal.NewResolver()
}
