// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"

	"github.com/pkg/errors"
	v1alpha1 "github.com/vmware-tanzu-private/core/apis/client/v1alpha1"
	"github.com/vmware-tanzu-private/core/pkg/v1/client"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/region"
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
	tanzuConfig, err := client.GetConfig()
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

func (trm *tanzuRegionManager) SaveRegionContext(region region.RegionContext) error {
	return client.AddServer(convertRegionContextToServer(region), false)
}

func (trm *tanzuRegionManager) UpsertRegionContext(region region.RegionContext) error {
	return client.PutServer(convertRegionContextToServer(region), false)
}

func (trm *tanzuRegionManager) DeleteRegionContext(clusterName string) error {
	currentServer, err := client.GetCurrentServer()
	if err != nil {
		return err
	}

	if clusterName != "" && clusterName != currentServer.Name {
		return fmt.Errorf("Cannot delete cluster %s. It is not the current cluster", clusterName)
	}

	if err = client.RemoveServer(currentServer.Name); err != nil {
		return err
	}

	return nil
}

func (trm *tanzuRegionManager) SetCurrentContext(clusterName string, contextName string) error {
	return client.SetCurrentServer(clusterName)
}

func (trm *tanzuRegionManager) GetCurrentContext() (region.RegionContext, error) {
	currentServer, err := client.GetCurrentServer()
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
