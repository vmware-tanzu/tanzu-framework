// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	yaml "gopkg.in/yaml.v3"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/aws"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/azure"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/docker"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/vsphere"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigproviders"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigupdater"
)

// ExportVSphereConfig creates return payload of config file string from incoming params object
func (app *App) ExportVSphereConfig(params vsphere.ExportTKGConfigForVsphereParams) middleware.Responder {
	var configString string
	// create the provider object with the configuration data
	config, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewVSphereConfig(params.Params)
	if err == nil {
		configString, err = transformConfigToString(config)
	}
	if err == nil {
		return vsphere.NewExportTKGConfigForVsphereOK().WithPayload(configString)
	}
	return vsphere.NewExportTKGConfigForVsphereInternalServerError().WithPayload(Err(err))
}

// ExportDockerConfig creates return payload of config file string from incoming params object
func (app *App) ExportDockerConfig(params docker.ExportTKGConfigForDockerParams) middleware.Responder {
	var configString string
	// create the provider object with the configuration data
	config, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewDockerConfig(params.Params)
	if err == nil {
		configString, err = transformConfigToString(config)
	}
	if err == nil {
		return docker.NewExportTKGConfigForDockerOK().WithPayload(configString)
	}
	return docker.NewExportTKGConfigForDockerInternalServerError().WithPayload(Err(err))
}

// ExportAzureConfig creates return payload of config file string from incoming params object
func (app *App) ExportAzureConfig(params azure.ExportTKGConfigForAzureParams) middleware.Responder {
	var configString string
	// create the provider object with the configuration data
	config, err := tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewAzureConfig(params.Params)
	if err == nil {
		configString, err = transformConfigToString(config)
	}
	if err == nil {
		return azure.NewExportTKGConfigForAzureOK().WithPayload(configString)
	}
	return azure.NewExportTKGConfigForAzureInternalServerError().WithPayload(Err(err))
}

// ExportAWSConfig creates return payload of config file string from incoming params object
func (app *App) ExportAWSConfig(params aws.ExportTKGConfigForAWSParams) middleware.Responder {
	var config *tkgconfigproviders.AWSConfig
	var configString string

	encodedCreds, err := app.awsClient.EncodeCredentials()
	if err == nil {
		// create the provider object with the configuration data
		config, err = tkgconfigproviders.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).NewAWSConfig(params.Params, encodedCreds)
	}
	if err == nil {
		configString, err = transformConfigToString(config)
	}
	if err == nil {
		return aws.NewExportTKGConfigForAWSOK().WithPayload(configString)
	}
	return aws.NewExportTKGConfigForAWSInternalServerError().WithPayload(Err(err))
}

func transformConfigToString(config interface{}) (out string, err error) {
	var configMap map[string]string
	var configByte []byte

	// turn the configuration object into a map
	configMap, err = tkgconfigupdater.CreateConfigMap(config)
	if err == nil {
		// turn the map into a byte array
		configByte, err = yaml.Marshal(&configMap)
	}
	if err == nil {
		return string(configByte), nil
	}
	return "", err
}
