// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clientcreator"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/docker"
)

// IsDockerDaemonAvailable validates docker daemon availability
func (app *App) IsDockerDaemonAvailable(params docker.CheckIfDockerDaemonAvailableParams) middleware.Responder {
	allClients, err := clientcreator.CreateAllClients(app.AppConfig, app.TKGConfigReaderWriter)
	if err != nil {
		return docker.NewCheckIfDockerDaemonAvailableInternalServerError().WithPayload(Err(err))
	}

	c, err := client.New(client.Options{
		ClusterCtlClient:         allClients.ClusterCtlClient,
		ReaderWriterConfigClient: allClients.ConfigClient,
		RegionManager:            allClients.RegionManager,
		TKGConfigDir:             app.AppConfig.TKGConfigDir,
		Timeout:                  app.TKGTimeout,
		FeaturesClient:           allClients.FeaturesClient,
		TKGConfigProvidersClient: allClients.TKGConfigProvidersClient,
		TKGBomClient:             allClients.TKGBomClient,
		TKGConfigUpdater:         allClients.TKGConfigUpdaterClient,
		TKGPathsClient:           allClients.TKGConfigPathsClient,
		ClusterClientFactory:     clusterclient.NewClusterClientFactory(),
	})
	if err != nil {
		return docker.NewCheckIfDockerDaemonAvailableInternalServerError().WithPayload(Err(err))
	}

	if err := c.ValidatePrerequisites(true, false, true); err != nil {
		return docker.NewCheckIfDockerDaemonAvailableBadRequest().WithPayload(Err(err))
	}

	status := models.DockerDaemonStatus{Status: true}

	return docker.NewCheckIfDockerDaemonAvailableOK().WithPayload(&status)
}
