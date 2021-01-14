// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package context

import (
	"context"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ControllerManagerContext struct {
	BOMImagePath         string
	BOMMetadataImagePath string
	Context              context.Context
	Client               client.Client
	Logger               logr.Logger
	Scheme               *runtime.Scheme
}
