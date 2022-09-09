// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package tkgconfigproviders helps setup and modify configs for TKG supported providers
package tkgconfigproviders

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigreaderwriter"
)

type client struct {
	configDir             string
	tkgBomClient          tkgconfigbom.Client
	tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter
}

// New creates new tkgconfig providers client
func New(configDir string, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) Client {
	tkgConfigProvidersClient := &client{
		configDir:             configDir,
		tkgBomClient:          tkgconfigbom.New(configDir, tkgConfigReaderWriter),
		tkgConfigReaderWriter: tkgConfigReaderWriter,
	}
	return tkgConfigProvidersClient
}

// Client implements TKG provider configuration related functions
type Client interface {
	NewAWSConfig(params *models.AWSRegionalClusterParams, encodedCredentials string) (*AWSConfig, error)
	GetAzureVMImageInfo(tkrVersion string) (*tkgconfigbom.AzureInfo, error)
	GetAWSAMIInfo(bomConfiguration *tkgconfigbom.BOMConfiguration, awsRegion string) (*tkgconfigbom.AMIInfo, error)
	NewAzureConfig(params *models.AzureRegionalClusterParams) (*AzureConfig, error)
	NewVSphereConfig(params *models.VsphereRegionalClusterParams) (*VSphereConfig, error)
	NewDockerConfig(params *models.DockerRegionalClusterParams) (*DockerConfig, error)
	CreateAWSParams(res *AWSConfig) (params *models.AWSRegionalClusterParams, err error)
	CreateAzureParams(res *AzureConfig) (params *models.AzureRegionalClusterParams, err error)
	CreateDockerParams(res *DockerConfig) (params *models.DockerRegionalClusterParams, err error)
	CreateVSphereParams(res *VSphereConfig) (params *models.VsphereRegionalClusterParams, err error)
}

// TKGConfigReaderWriter returns tkgConfigReaderWriter client
func (c *client) TKGConfigReaderWriter() tkgconfigreaderwriter.TKGConfigReaderWriter {
	return c.tkgConfigReaderWriter
}
