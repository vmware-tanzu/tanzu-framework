// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	yaml "gopkg.in/yaml.v3"

	"github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/aws"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/azure"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/docker"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/vsphere"
	"k8s.io/klog/v2/klogr"
)

// ImportVSphereConfig receives a config file as a string (in ImportTKGConfigForVsphereParams)
// and returns a VsphereRegionalClusterParams object
func (app *App) ImportVSphereConfig(params vsphere.ImportTKGConfigForVsphereParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.VSphereConfig{}
	err := populateVsphereConfigFromString(fileContent, configObject)

	var configPayload *models.VsphereRegionalClusterParams
	if err == nil {
		configPayload, err = tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter, klogr.New().WithName("vsphere-import-handler")).CreateVSphereParams(configObject)
	}
	if err == nil {
		return vsphere.NewImportTKGConfigForVsphereOK().WithPayload(configPayload)
	}
	return vsphere.NewImportTKGConfigForVsphereInternalServerError().WithPayload(Err(err))
}

// ImportAzureConfig receives a config file as a string (in ImportTKGConfigForAzureParams)
// and returns a AzureRegionalClusterParams object
func (app *App) ImportAzureConfig(params azure.ImportTKGConfigForAzureParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.AzureConfig{}
	err := populateAzureConfigFromString(fileContent, configObject)

	var configPayload *models.AzureRegionalClusterParams
	if err == nil {
		configPayload, err = tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter, klogr.New().WithName("azure-import-handler")).CreateAzureParams(configObject)
	}
	if err == nil {
		return azure.NewImportTKGConfigForAzureOK().WithPayload(configPayload)
	}
	return azure.NewImportTKGConfigForAzureInternalServerError().WithPayload(Err(err))
}

// ImportAwsConfig receives a config file as a string (in ImportTKGConfigForAWSParams)
// and returns a AWSRegionalClusterParams object
func (app *App) ImportAwsConfig(params aws.ImportTKGConfigForAWSParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.AWSConfig{}
	err := populateAWSConfigFromString(fileContent, configObject)

	var configPayload *models.AWSRegionalClusterParams
	if err == nil {
		configPayload, err = tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter, klogr.New().WithName("aws-import-handler")).CreateAWSParams(configObject)
	}

	if err == nil {
		return aws.NewImportTKGConfigForAWSOK().WithPayload(configPayload)
	}
	return aws.NewImportTKGConfigForAWSInternalServerError().WithPayload(Err(err))
}

// ImportDockerConfig receives a config file as a string (in ImportTKGConfigForDockerParams)
// and returns a DockerRegionalClusterParams object
func (app *App) ImportDockerConfig(params docker.ImportTKGConfigForDockerParams) middleware.Responder {
	var fileContent = params.Params.Filecontents
	configObject := &tkgconfigproviders.DockerConfig{}
	err := populateDockerConfigFromString(fileContent, configObject)

	var configPayload *models.DockerRegionalClusterParams
	if err == nil {
		configPayload, err = tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter, klogr.New().WithName("docker-import-handler")).CreateDockerParams(configObject)
	}

	if err == nil {
		return docker.NewImportTKGConfigForDockerOK().WithPayload(configPayload)
	}
	return docker.NewImportTKGConfigForDockerInternalServerError().WithPayload(Err(err))
}

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
