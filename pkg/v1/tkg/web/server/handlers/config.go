// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/aws"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/azure"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/docker"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/vsphere"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
)

// ApplyTKGConfigForVsphere applies TKG configuration for vSphere
func (app *App) ApplyTKGConfigForVsphere(params vsphere.ApplyTKGConfigForVsphereParams) middleware.Responder { // nolint:dupl
	config, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewVSphereConfig(params.Params)
	if err != nil {
		return vsphere.NewApplyTKGConfigForVsphereInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, config)
	if err != nil {
		return vsphere.NewApplyTKGConfigForVsphereInternalServerError().WithPayload(Err(err))
	}
	return vsphere.NewApplyTKGConfigForVsphereOK().WithPayload(&models.ConfigFileInfo{Path: app.getFilePathForSavingConfig()})
}

// ApplyTKGConfigForAWS applies TKG configuration for AWS
func (app *App) ApplyTKGConfigForAWS(params aws.ApplyTKGConfigForAWSParams) middleware.Responder {
	if app.awsClient == nil {
		return aws.NewApplyTKGConfigForAWSInternalServerError().WithPayload(Err(errors.New("aws client is not initialized properly")))
	}
	encodedCreds, err := app.awsClient.EncodeCredentials()
	if err != nil {
		return aws.NewApplyTKGConfigForAWSInternalServerError().WithPayload(Err(err))
	}

	config, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewAWSConfig(params.Params, encodedCreds)
	if err != nil {
		return aws.NewApplyTKGConfigForAWSInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, config)
	if err != nil {
		return aws.NewApplyTKGConfigForAWSInternalServerError().WithPayload(Err(err))
	}

	return aws.NewApplyTKGConfigForAWSOK().WithPayload(&models.ConfigFileInfo{Path: app.getFilePathForSavingConfig()})
}

// ApplyTKGConfigForAzure applies TKG configuration for Azure
func (app *App) ApplyTKGConfigForAzure(params azure.ApplyTKGConfigForAzureParams) middleware.Responder {
	if app.azureClient == nil {
		return azure.NewApplyTKGConfigForAzureInternalServerError().WithPayload(Err(errors.New("azure client is not initialized properly")))
	}

	config, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewAzureConfig(params.Params)
	if err != nil {
		return azure.NewApplyTKGConfigForAzureInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, config)
	if err != nil {
		return azure.NewApplyTKGConfigForAzureInternalServerError().WithPayload(Err(err))
	}

	return azure.NewApplyTKGConfigForAzureOK().WithPayload(&models.ConfigFileInfo{Path: app.getFilePathForSavingConfig()})
}

// ApplyTKGConfigForDocker applies TKG configuration for Docker
func (app *App) ApplyTKGConfigForDocker(params docker.ApplyTKGConfigForDockerParams) middleware.Responder { // nolint:dupl
	config, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewDockerConfig(params.Params)
	if err != nil {
		return docker.NewApplyTKGConfigForDockerInternalServerError().WithPayload(Err(err))
	}

	err = tkgconfigupdater.SaveConfig(app.getFilePathForSavingConfig(), app.TKGConfigReaderWriter, config)
	if err != nil {
		return docker.NewApplyTKGConfigForDockerInternalServerError().WithPayload(Err(err))
	}

	return docker.NewApplyTKGConfigForDockerOK().WithPayload(&models.ConfigFileInfo{Path: app.getFilePathForSavingConfig()})
}
