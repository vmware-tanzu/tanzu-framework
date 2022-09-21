// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package templateresolver

import (
	"context"

	"github.com/go-logr/logr"

	"github.com/vmware-tanzu/tanzu-framework/tkg/vc"
)

//go:generate counterfeiter -o ../../../../util/fakes/templateresolver.go --fake-name TemplateResolver . TemplateResolver

// TemplateResolver resolves vSphere templates
type TemplateResolver interface {
	// Resolve returns VM template path and MOIDs satisfying query constraints.
	Resolve(ctx context.Context, svrContext VSphereContext, query Query, vcClient vc.Client) Result
	GetVSphereEndpoint(svrContext VSphereContext) (vc.Client, error)
}

// New returns a newly created instance of the vSphere template resolver implementation.
func New(log logr.Logger) TemplateResolver {
	return NewResolver(log)
}
