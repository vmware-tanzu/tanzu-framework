// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"os"
	"sort"

	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"
	"sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/credentials"

	awsclient "github.com/vmware-tanzu/tanzu-framework/tkg/aws"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/aws"
)

// SetAWSEndPoint verify and sets AWS account
func (app *App) SetAWSEndPoint(params aws.SetAWSEndpointParams) middleware.Responder {
	var err error
	var creds *credentials.AWSCredentials

	if params.AccountParams.AccessKeyID != "" && params.AccountParams.SecretAccessKey != "" {
		creds = &credentials.AWSCredentials{
			Region:          params.AccountParams.Region,
			AccessKeyID:     params.AccountParams.AccessKeyID,
			SecretAccessKey: params.AccountParams.SecretAccessKey,
			SessionToken:    params.AccountParams.SessionToken,
		}
	} else {
		if params.AccountParams.ProfileName != "" {
			os.Setenv(constants.ConfigVariableAWSProfile, params.AccountParams.ProfileName)
		}
		creds, err = credentials.NewAWSCredentialFromDefaultChain(params.AccountParams.Region)
		if err != nil {
			return aws.NewSetAWSEndpointInternalServerError().WithPayload(Err(err))
		}
	}
	client, err := awsclient.New(*creds)
	if err != nil {
		return aws.NewSetAWSEndpointInternalServerError().WithPayload(Err(err))
	}
	err = client.VerifyAccount()
	if err != nil {
		return aws.NewSetAWSEndpointInternalServerError().WithPayload(Err(err))
	}
	app.awsClient = client
	return aws.NewCreateAWSRegionalClusterOK()
}

// GetVPCs gets all VPCs under current AWS account
func (app *App) GetVPCs(params aws.GetVPCsParams) middleware.Responder {
	if app.awsClient == nil {
		return aws.NewGetVPCsInternalServerError().WithPayload(Err(errors.New("aws client is not initialized properly")))
	}

	vpcs, err := app.awsClient.ListVPCs()
	if err != nil {
		return aws.NewGetVPCsInternalServerError().WithPayload(Err(err))
	}

	return aws.NewGetVPCsOK().WithPayload(vpcs)
}

// GetAWSAvailabilityZones gets availability zones under user-specified region
func (app *App) GetAWSAvailabilityZones(params aws.GetAWSAvailabilityZonesParams) middleware.Responder {
	if app.awsClient == nil {
		return aws.NewGetAWSAvailabilityZonesInternalServerError().WithPayload(Err(errors.New("aws client is not initialized properly")))
	}

	azs, err := app.awsClient.ListAvailabilityZones()
	if err != nil {
		return aws.NewGetAWSAvailabilityZonesInternalServerError().WithPayload(Err(err))
	}

	return aws.NewGetAWSAvailabilityZonesOK().WithPayload(azs)
}

// GetAWSRegions returns list of AWS regions
func (app *App) GetAWSRegions(params aws.GetAWSRegionsParams) middleware.Responder {
	bomConfig, err := tkgconfigbom.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).GetDefaultTkrBOMConfiguration()
	if err != nil {
		return aws.NewGetAWSRegionsInternalServerError().WithPayload(Err(err))
	}
	regions := []string{}
	for region := range bomConfig.AMI {
		regions = append(regions, region)
	}
	sort.Strings(regions)
	return aws.NewGetAWSRegionsOK().WithPayload(regions)
}

// GetAWSSubnets gets all subnets under given vpc ID
func (app *App) GetAWSSubnets(params aws.GetAWSSubnetsParams) middleware.Responder {
	if app.awsClient == nil {
		return aws.NewGetAWSSubnetsInternalServerError().WithPayload(Err(errors.New("aws client is not initialized properly")))
	}

	subnets, err := app.awsClient.ListSubnets(params.VpcID)
	if err != nil {
		return aws.NewGetAWSSubnetsInternalServerError().WithPayload(Err(err))
	}

	return aws.NewGetAWSSubnetsOK().WithPayload(subnets)
}

// GetAWSNodeTypes gets aws node types
func (app *App) GetAWSNodeTypes(params aws.GetAWSNodeTypesParams) middleware.Responder {
	if app.awsClient == nil {
		return aws.NewGetAWSNodeTypesInternalServerError().WithPayload(Err(errors.New("aws client is not initialized properly")))
	}

	var result []string
	var err error
	if params.Az == nil {
		result, err = app.awsClient.ListInstanceTypes("")
	} else {
		result, err = app.awsClient.ListInstanceTypes(*params.Az)
	}
	if err != nil {
		return aws.NewGetAWSNodeTypesInternalServerError().WithPayload(Err(err))
	}
	return aws.NewGetAWSNodeTypesOK().WithPayload(result)
}

// GetAWSCredentialProfiles gets aws credential profile
func (app *App) GetAWSCredentialProfiles(params aws.GetAWSCredentialProfilesParams) middleware.Responder {
	res, err := awsclient.ListCredentialProfiles("")
	if err != nil {
		return aws.NewGetAWSCredentialProfilesInternalServerError().WithPayload(Err(err))
	}

	return aws.NewGetAWSCredentialProfilesOK().WithPayload(res)
}

// GetAWSOSImages gets os information for AWS
func (app *App) GetAWSOSImages(params aws.GetAWSOSImagesParams) middleware.Responder {
	bomConfig, err := tkgconfigbom.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter).GetDefaultTkrBOMConfiguration()
	if err != nil {
		return aws.NewGetAWSOSImagesInternalServerError().WithPayload(Err(err))
	}

	results := []*models.AWSVirtualMachine{}

	amis, exists := bomConfig.AMI[params.Region]
	if !exists {
		return aws.NewGetAWSOSImagesInternalServerError().WithPayload(Err(errors.Errorf("No AMI found for the provided region '%s'", params.Region)))
	}

	for _, ami := range amis {
		displayName := fmt.Sprintf("%s-%s-%s (%s)", ami.OSInfo.Name, ami.OSInfo.Version, ami.OSInfo.Arch, ami.ID)
		results = append(results, &models.AWSVirtualMachine{
			Name: displayName,
			OsInfo: &models.OSInfo{
				Name:    ami.OSInfo.Name,
				Version: ami.OSInfo.Version,
				Arch:    ami.OSInfo.Arch,
			},
		})
	}
	return aws.NewGetAWSOSImagesOK().WithPayload(results)
}
