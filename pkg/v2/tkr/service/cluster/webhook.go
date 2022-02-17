// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package cluster

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type Webhook struct {
}

func (cw Webhook) Handle(_ context.Context, _ admission.Request) admission.Response {
	return admission.Allowed("Everything is fine :)")
}
