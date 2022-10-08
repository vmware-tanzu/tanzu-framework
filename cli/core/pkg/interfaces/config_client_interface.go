// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package interfaces is collection of generic interfaces
package interfaces

import (
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

//go:generate counterfeiter -o ../fakes/config_client_fake.go . ConfigClient
type ConfigClient interface {
	GetEnvConfigurations() map[string]string
}

type configClientImpl struct{}

func NewConfigClient() ConfigClient {
	return &configClientImpl{}
}

func (cc *configClientImpl) GetEnvConfigurations() map[string]string {
	return config.GetEnvConfigurations()
}
