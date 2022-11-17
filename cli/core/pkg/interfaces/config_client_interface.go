// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package interfaces is collection of generic interfaces
package interfaces

import (
	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/config"
)

//go:generate counterfeiter -o ../fakes/config_client_fake.go . ConfigClient
type ConfigClientWrapper interface {
	GetEnvConfigurations() map[string]string
	StoreClientConfig(clientConfig *configapi.ClientConfig) error
	AcquireTanzuConfigLock()
	ReleaseTanzuConfigLock()
}

type configClientWrapperImpl struct{}

func NewConfigClient() ConfigClientWrapper {
	return &configClientWrapperImpl{}
}

func (cc *configClientWrapperImpl) GetEnvConfigurations() map[string]string {
	return config.GetEnvConfigurations()
}

func (cc *configClientWrapperImpl) AcquireTanzuConfigLock() {
	config.AcquireTanzuConfigLock()
}

func (cc *configClientWrapperImpl) ReleaseTanzuConfigLock() {
	config.ReleaseTanzuConfigLock()
}

func (cc *configClientWrapperImpl) StoreClientConfig(clientConfig *configapi.ClientConfig) error {
	return config.StoreClientConfig(clientConfig)
}
