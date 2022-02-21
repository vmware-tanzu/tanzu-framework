// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cluster provides the TKR Resolver mutating webhook on CAPI Cluster.
package cluster

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
)

type Webhook struct {
	TKRResolver resolver.Resolver
}

func (cw *Webhook) Handle(_ context.Context, req admission.Request) admission.Response { // nolint:gocritic // suppress linter error: hugeParam: req is heavy (400 bytes); consider passing by pointer (gocritic)
	return admission.Allowed("Everything is fine :) req.Name: '" + req.Name + "'")
}
