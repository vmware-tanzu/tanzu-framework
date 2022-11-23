// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"os"
	"strings"

	"github.com/vmware-tanzu/tanzu-framework/tkg/aws"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"

	"github.com/pkg/errors"

	"sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/cloudformation/bootstrap"
	"sigs.k8s.io/cluster-api-provider-aws/v2/cmd/clusterawsadm/credentials"
)

const awsCredentialError = "failed to gather credentials for operation that requires client-side access to AWS account. Provide credentials for the relevant AWS account using either AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN environment variables, configuration variables (not for delete) or set up a profile using the AWS CLI https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html and use the AWS_PROFILE environment variable when running this command."

// CreateAWSCloudFormationStack create aws cloud formation stack
func (c *TkgClient) CreateAWSCloudFormationStack() error {
	log.SendProgressUpdate(statusRunning, StepConfigPrerequisite, InitRegionSteps)
	log.Info("Creating AWS CloudFormation Stack")
	template, err := c.generateAWSBootstrapTemplate()

	if err != nil {
		return errors.Wrap(err, "failed to generate template")
	}

	creds, err := c.GetAWSCreds()
	if err != nil {
		return errors.Wrap(err, "unable to retrieve AWS credentials")
	}

	awsClient, err := aws.New(*creds)
	if err != nil {
		return errors.Wrap(err, "failed create AWS client")
	}

	err = awsClient.CreateCloudFormationStackWithTemplate(template)
	if err != nil {
		return errors.Wrap(err, "failed to create aws CloudFormation stack")
	}
	return nil
}

// GenerateAWSCloudFormationTemplate produces the CloudFormation template YAML
func (c *TkgClient) GenerateAWSCloudFormationTemplate() (string, error) {
	template, err := c.generateAWSBootstrapTemplate()
	if err != nil {
		return "", err
	}
	cfnTemplate := template.RenderCloudFormation()
	dat, err := cfnTemplate.YAML()
	if err != nil {
		return "", err
	}
	return string(dat), nil
}

func (c *TkgClient) generateAWSBootstrapTemplate() (*bootstrap.Template, error) {
	// Don't need a full client
	awsClient, err := aws.New(credentials.AWSCredentials{})
	if err != nil {
		return nil, errors.Wrap(err, "failed create AWS client")
	}

	tmcDisablement := false

	tmcDisablementVar, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableDisableTMCCloudPermissions)
	if err == nil && strings.EqualFold(tmcDisablementVar, "true") {
		tmcDisablement = true
	}

	template, err := awsClient.GenerateBootstrapTemplate(aws.GenerateBootstrapTemplateInput{
		DisableTanzuMissionControlPermissions: tmcDisablement,
	})
	if err != nil {
		return nil, err
	}
	return template, nil
}

// GetAWSCreds get aws credentials
func (c *TkgClient) GetAWSCreds() (*credentials.AWSCredentials, error) {
	region, err := c.getAWSRegion()
	if err != nil {
		return &credentials.AWSCredentials{}, err
	}

	creds, err := c.getAWSCredFromConfig(region)
	if err == nil {
		return creds, nil
	}

	if profile, err := c.readerwriterConfigClient.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSProfile); err == nil && profile != "" {
		os.Setenv(constants.ConfigVariableAWSProfile, profile)
	}

	// find AWS Credentials in default chain if they are not in tkgconfig or set as environment variables
	creds, err = credentials.NewAWSCredentialFromDefaultChain(region)
	if err != nil {
		log.V(3).Error(err, "AWS SDK unable to resolve credential provider chain")
		return nil, errors.New(awsCredentialError)
	}
	return creds, nil
}

func (c *TkgClient) getAWSCredFromConfig(region string) (*credentials.AWSCredentials, error) {
	var err error

	accessKeyID, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSAccessKeyID)
	if err != nil || accessKeyID == "" {
		return &credentials.AWSCredentials{}, errors.Wrapf(err, "failed to get %s", constants.ConfigVariableAWSAccessKeyID)
	}

	secretAccessKey, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSSecretAccessKey)
	if err != nil || secretAccessKey == "" {
		return &credentials.AWSCredentials{}, errors.Wrapf(err, "failed to get %s", constants.ConfigVariableAWSSecretAccessKey)
	}
	sessionToken, _ := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSSessionToken)

	return &credentials.AWSCredentials{Region: region, AccessKeyID: accessKeyID, SecretAccessKey: secretAccessKey, SessionToken: sessionToken}, nil
}

func (c *TkgClient) getAWSRegion() (string, error) {
	region, err := c.TKGConfigReaderWriter().Get(constants.ConfigVariableAWSRegion)
	if err != nil || region == "" {
		return "", errors.New("cannot find AWS region. If using the CLI, set the AWS_REGION environment variable appropriately")
	}
	return region, nil
}
