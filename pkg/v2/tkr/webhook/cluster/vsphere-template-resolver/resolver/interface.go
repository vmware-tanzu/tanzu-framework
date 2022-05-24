// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package templateresolver

// TemplateResolver resolves vSphere templates
type TemplateResolver interface {
	// Resolve returns VM template path and MOIDs satisfying query constraints.
	Resolve(svrContext VSphereContext, query Query) Result
}

// New returns a newly created instance of the vSphere template resolver implementation.
func New() TemplateResolver {
	return NewResolver()
}
