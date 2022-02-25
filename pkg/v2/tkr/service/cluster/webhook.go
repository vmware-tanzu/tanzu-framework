// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package cluster provides cluster.Webhook which is a mutating webhook handler for CAPI Cluster objects.
package cluster

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Webhook struct {
}

func (cw *Webhook) Handle(_ context.Context, req admission.Request) admission.Response { // nolint:gocritic // suppress linter error: hugeParam: req is heavy (400 bytes); consider passing by pointer (gocritic)
	return admission.Allowed("Everything is fine :) req.Name: '" + req.Name + "'")
}
