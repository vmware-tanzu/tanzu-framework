// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package context

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type TanzuKubernetesReleaseDiscoverOptions struct {
	InitialDiscoveryFrequency    time.Duration
	ContinuousDiscoveryFrequency time.Duration
}

type ControllerManagerContext struct {
	BOMImagePath         string
	BOMMetadataImagePath string
	Context              context.Context
	Client               client.Client
	Logger               logr.Logger
	Scheme               *runtime.Scheme
	VerifyRegistryCert   bool
	RegistryCertPath     string
	TKRDiscoveryOption   TanzuKubernetesReleaseDiscoverOptions
}

func NewTanzuKubernetesReleaseDiscoverOptions(initFreq, continuousFreq float64) TanzuKubernetesReleaseDiscoverOptions {
	return TanzuKubernetesReleaseDiscoverOptions{
		InitialDiscoveryFrequency:    time.Duration(initFreq) * time.Second,
		ContinuousDiscoveryFrequency: time.Duration(continuousFreq) * time.Second,
	}
}
