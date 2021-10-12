// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	yaml "gopkg.in/yaml.v3"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/aws"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/azure"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/docker"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/vsphere"
)

func (app *App) ImportVSphereConfig(params vsphere.ImportTKGConfigForVsphereParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.VSphereConfig{}
	err := populateVsphereConfigFromString(fileContent, configObject)

	configPayload, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).CreateVSphereParams(configObject)

	if err == nil {
		return vsphere.NewImportTKGConfigForVsphereOK().WithPayload(configPayload)
	}
	return vsphere.NewImportTKGConfigForVsphereInternalServerError().WithPayload(Err(err))
}

func (app *App) ImportAzureConfig(params azure.ImportTKGConfigForAzureParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.AzureConfig{}
	err := populateAzureConfigFromString(fileContent, configObject)

	configPayload, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).CreateAzureParams(configObject)

	if err == nil {
		return azure.NewImportTKGConfigForAzureOK().WithPayload(configPayload)
	}
	return azure.NewImportTKGConfigForAzureInternalServerError().WithPayload(Err(err))
}

func (app *App) ImportAwsConfig(params aws.ImportTKGConfigForAWSParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.AWSConfig{}
	err := populateAWSConfigFromString(fileContent, configObject)

	configPayload, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).CreateAWSParams(configObject)

	if err == nil {
		return aws.NewImportTKGConfigForAWSOK().WithPayload(configPayload)
	}
	return aws.NewImportTKGConfigForAWSInternalServerError().WithPayload(Err(err))
}

func (app *App) ImportDockerConfig(params docker.ImportTKGConfigForDockerParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.DockerConfig{}
	err := populateDockerConfigFromString(fileContent, configObject)

	configPayload, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).CreateDockerParams(configObject)

	if err == nil {
		return docker.NewImportTKGConfigForDockerOK().WithPayload(configPayload)
	}
	return docker.NewImportTKGConfigForDockerInternalServerError().WithPayload(Err(err))
}

// TODO SHIMON: can these populate methods be generic?
func populateVsphereConfigFromString(input string, config *tkgconfigproviders.VSphereConfig) error {
	// turn string into byte array and unmarshal the byteArray into the config object
	return yaml.Unmarshal([]byte(input), &config)
}

func populateAWSConfigFromString(input string, config *tkgconfigproviders.AWSConfig) error {
	// turn string into byte array and unmarshal the byteArray into the config object
	return yaml.Unmarshal([]byte(input), &config)
}

func populateAzureConfigFromString(input string, config *tkgconfigproviders.AzureConfig) error {
	// turn string into byte array and unmarshal the byteArray into the config object
	return yaml.Unmarshal([]byte(input), &config)
}

func populateDockerConfigFromString(input string, config *tkgconfigproviders.DockerConfig) error {
	// turn string into byte array and unmarshal the byteArray into the config object
	return yaml.Unmarshal([]byte(input), &config)
}
