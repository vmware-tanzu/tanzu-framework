// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/tkg-cli/pkg/region"

	"github.com/vmware-tanzu-private/core/apis/config/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/config"
)

type tanzuRegionManager struct {
}

type tanzuRegionManagerFactory struct {
}

// NewFactory creates a new tanzuRegionManagerFactory which implements
// region.ManagerFactory
func NewFactory() region.ManagerFactory {
	return &tanzuRegionManagerFactory{}
}

func (trm *tanzuRegionManager) ListRegionContexts() ([]region.RegionContext, error) {
	tanzuConfig, err := config.GetConfig()
	if err != nil {
		return []region.RegionContext{}, err
	}

	var regionClusters []region.RegionContext
	for _, server := range tanzuConfig.KnownServers {
		if server.Type == v1alpha1.ManagementClusterServerType {
			regionContext := convertServerToRegionContextFull(server,
				server.Name == tanzuConfig.CurrentServer)

			regionClusters = append(regionClusters, regionContext)
		}
	}

	return regionClusters, nil
}

func (trm *tanzuRegionManager) SaveRegionContext(regionCtxt region.RegionContext) error {
	return config.AddServer(convertRegionContextToServer(regionCtxt), false)
}

func (trm *tanzuRegionManager) UpsertRegionContext(regionCtxt region.RegionContext) error {
	return config.PutServer(convertRegionContextToServer(regionCtxt), false)
}

func (trm *tanzuRegionManager) DeleteRegionContext(clusterName string) error {
	currentServer, err := config.GetCurrentServer()
	if err != nil {
		return err
	}

	if clusterName != "" && clusterName != currentServer.Name {
		return fmt.Errorf("cannot delete cluster %s, it is not the current cluster", clusterName)
	}

	if err := config.RemoveServer(currentServer.Name); err != nil {
		return err
	}

	return nil
}

func (trm *tanzuRegionManager) SetCurrentContext(clusterName, contextName string) error {
	return config.SetCurrentServer(clusterName)
}

func (trm *tanzuRegionManager) GetCurrentContext() (region.RegionContext, error) {
	currentServer, err := config.GetCurrentServer()
	if err != nil {
		return region.RegionContext{}, err
	}

	if !currentServer.IsManagementCluster() {
		return region.RegionContext{}, errors.Errorf("The current server is not a management cluster")
	}

	return convertServerToRegionContext(currentServer), nil
}

func (trmf *tanzuRegionManagerFactory) CreateManager(configFile string) (region.Manager, error) {
	return &tanzuRegionManager{}, nil
}

func convertServerToRegionContext(server *v1alpha1.Server) region.RegionContext {
	return convertServerToRegionContextFull(server, false)
}

func convertServerToRegionContextFull(server *v1alpha1.Server, isCurrentContext bool) region.RegionContext {
	return region.RegionContext{
		ClusterName:      server.Name,
		ContextName:      server.ManagementClusterOpts.Context,
		SourceFilePath:   server.ManagementClusterOpts.Path,
		IsCurrentContext: isCurrentContext,
	}
}

func convertRegionContextToServer(regionContext region.RegionContext) *v1alpha1.Server {
	return &v1alpha1.Server{
		Name: regionContext.ClusterName,
		Type: v1alpha1.ManagementClusterServerType,
		ManagementClusterOpts: &v1alpha1.ManagementClusterServer{
			Path:    regionContext.SourceFilePath,
			Context: regionContext.ContextName,
		},
	}
}
