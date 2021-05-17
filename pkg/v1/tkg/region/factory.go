// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package region

// ManagerFactory provides interface for region manager factory
type ManagerFactory interface {
	CreateManager(configPath string) (Manager, error)
}

type managerFactory struct{}

// NewFactory creates new manager factory
func NewFactory() ManagerFactory {
	return &managerFactory{}
}

func (mf *managerFactory) CreateManager(configPath string) (Manager, error) {
	return New(configPath)
}
