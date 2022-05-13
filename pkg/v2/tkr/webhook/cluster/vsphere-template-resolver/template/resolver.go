// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package template provides the vSphere template Resolver mutating webhook on CAPI Cluster.
package template

import (
	"context"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/tkr/resolver"
)

const VarTKRData = "TKR_DATA"

type Webhook struct {
	TKRResolver resolver.CachingResolver
	Log         logr.Logger
	Client      client.Client
	decoder     *admission.Decoder
}

func (cw *Webhook) InjectDecoder(decoder *admission.Decoder) error {
	cw.decoder = decoder
	return nil
}

func (cw *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response { // nolint:gocritic // suppress linter error: hugeParam: req is heavy (400 bytes); consider passing by pointer (gocritic)
	//TODO: Check if this Cluster is  CC cluster and then get the variables from Cluster and resolve the template
	return admission.Allowed("")
}
