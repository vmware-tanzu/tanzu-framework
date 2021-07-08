// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package avi defines client to connect to the avi server
package avi

import (
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"

	avi_models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// Client defines methods to access AVI controller via its REST API
type Client interface {
	VerifyAccount(params *avi_models.AviControllerParams) (bool, error)
	GetClouds() ([]*avi_models.AviCloud, error)
	GetServiceEngineGroups() ([]*avi_models.AviServiceEngineGroup, error)
	GetVipNetworks() ([]*avi_models.AviVipNetwork, error)
}

// MiniCloudClient defines a subset of the methods implemented by Cloud
type MiniCloudClient interface {
	GetAll(options ...session.ApiOptionsParams) ([]*models.Cloud, error)
}

// MiniServiceEngineGroupClient defines a subset of the methods implemented by Cloud
type MiniServiceEngineGroupClient interface {
	GetAll(options ...session.ApiOptionsParams) ([]*models.ServiceEngineGroup, error)
}

// MiniNetworkClient defines a subset of the methods implemented by Network
type MiniNetworkClient interface {
	GetAll(options ...session.ApiOptionsParams) ([]*models.Network, error)
}
