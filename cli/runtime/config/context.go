// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"fmt"

	configapi "github.com/vmware-tanzu/tanzu-framework/cli/runtime/apis/config/v1alpha1"
)

// GetContext by name.
func GetContext(name string) (*configapi.Context, error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return nil, err
	}

	return cfg.GetContext(name)
}

// ContextExists tells whether the Context by the given name exists.
func ContextExists(name string) (bool, error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return false, err
	}

	return cfg.HasContext(name), nil
}

// AddContext adds a Context to the config.
func AddContext(c *configapi.Context, setCurrent bool) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	cfg, err := GetClientConfigNoLock()
	if err != nil {
		return err
	}

	exists := cfg.HasContext(c.Name)
	if exists {
		return fmt.Errorf("context %q already exists", c.Name)
	}

	cfg.KnownContexts = append(cfg.KnownContexts, c)
	if setCurrent {
		err = cfg.SetCurrentContext(c.Type, c.Name)
		if err != nil {
			return err
		}
	}

	return StoreClientConfig(cfg)
}

// RemoveContext removes a Context from the config.
func RemoveContext(name string) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	cfg, err := GetClientConfigNoLock()
	if err != nil {
		return err
	}

	newContexts := make([]*configapi.Context, 0)
	var ctx *configapi.Context
	for _, c := range cfg.KnownContexts {
		if c.Name == name {
			ctx = c
		} else {
			newContexts = append(newContexts, c)
		}
	}

	if ctx == nil {
		return fmt.Errorf("context %s not found", name)
	}

	cfg.KnownContexts = newContexts
	newServers := []*configapi.Server{}
	for _, s := range cfg.KnownServers {
		if s.Name != name {
			newServers = append(newServers, s)
		}
	}
	cfg.KnownServers = newServers

	if cfg.CurrentContext[ctx.Type] == name {
		delete(cfg.CurrentContext, ctx.Type)
	}
	if cfg.CurrentServer == name {
		cfg.CurrentServer = ""
	}

	return StoreClientConfig(cfg)
}

// SetCurrentContext sets the current Context.
func SetCurrentContext(name string) error {
	// Acquire tanzu config lock
	AcquireTanzuConfigLock()
	defer ReleaseTanzuConfigLock()

	cfg, err := GetClientConfigNoLock()
	if err != nil {
		return err
	}

	ctx, err := cfg.GetContext(name)
	if err != nil {
		return err
	}
	err = cfg.SetCurrentContext(ctx.Type, name)
	if err != nil {
		return err
	}

	return StoreClientConfig(cfg)
}

// GetCurrentContext gets the current context.
func GetCurrentContext(ctxType configapi.ContextType) (c *configapi.Context, err error) {
	cfg, err := GetClientConfig()
	if err != nil {
		return nil, err
	}

	return cfg.GetCurrentContext(ctxType)
}
