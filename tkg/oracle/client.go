// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package oracle

import (
	"context"
	"time"

	"github.com/aunum/log"
	"github.com/pkg/errors"

	oraclecore "github.com/oracle/oci-go-sdk/v49/core"
	oracleidentity "github.com/oracle/oci-go-sdk/v49/identity"
	oracleworkrequests "github.com/oracle/oci-go-sdk/v49/workrequests"

	oraclecommon "github.com/oracle/oci-go-sdk/v49/common"
)

type client struct {
	configProvider    oraclecommon.ConfigurationProvider
	identityClient    oracleidentity.IdentityClient
	computeClient     oraclecore.ComputeClient
	workRequestClient oracleworkrequests.WorkRequestClient
}

// WaitForWorkRequest waits for an OPC work request specified by its OCID
func (c *client) WaitForWorkRequest(ctx context.Context, id *string, interval time.Duration) error {
	for {
		response, err := c.workRequestClient.GetWorkRequest(ctx, oracleworkrequests.GetWorkRequestRequest{
			WorkRequestId: id,
		})
		if err != nil {
			return err
		}
		log.Infof("wait for worker request %s, progress %f", *id, *response.PercentComplete)
		if response.Status == oracleworkrequests.WorkRequestStatusSucceeded {
			return nil
		}

		if response.Status == oracleworkrequests.WorkRequestStatusCanceling ||
			response.Status == oracleworkrequests.WorkRequestStatusCanceled ||
			response.Status == oracleworkrequests.WorkRequestStatusFailed {
			return errors.New("unexpected work request status %s" + string(response.Status))
		}
		time.Sleep(interval)
	}
}

// ImportImageSync initiates the image import from public endpoint and waits for it finishes synchronously
func (c *client) ImportImageSync(ctx context.Context, displayName, compartment, image string) (*oraclecore.Image, error) {
	request := oraclecore.CreateImageRequest{
		CreateImageDetails: oraclecore.CreateImageDetails{
			CompartmentId: &compartment,
			DisplayName:   &displayName,
			ImageSourceDetails: oraclecore.ImageSourceViaObjectStorageUriDetails{
				SourceUri: &image,
			},
		},
	}
	response, err := c.computeClient.CreateImage(ctx, request)
	if err != nil {
		return nil, err
	}
	if err := c.WaitForWorkRequest(ctx, response.OpcWorkRequestId, 10*time.Second); err != nil {
		return nil, err
	}
	return &response.Image, nil
}

// EnsureCompartmentExists ensures the specified compartment exists on the OCI
func (c *client) EnsureCompartmentExists(ctx context.Context, compartment string) (*oracleidentity.Compartment, error) {
	response, err := c.identityClient.GetCompartment(ctx, oracleidentity.GetCompartmentRequest{
		CompartmentId: &compartment,
	})
	if err != nil {
		return nil, err
	}
	log.Debugf("check compartment %s, response %v", compartment, response)
	return &response.Compartment, nil
}

// Region returns the region that the Oracle Client operates on
func (c *client) Region() (string, error) {
	return c.configProvider.Region()
}

// Credentials returns the credentials that the Oracle Client uses
func (c *client) Credentials() oraclecommon.ConfigurationProvider {
	return c.configProvider
}

func (c *client) IsUsingInstancePrincipal() (bool, error) {
	creds := c.Credentials()
	authConfig, err := creds.AuthType()
	if err != nil {
		return false, err
	}
	if authConfig.AuthType == oraclecommon.InstancePrincipal || authConfig.AuthType == oraclecommon.InstancePrincipalDelegationToken {
		return true, nil
	}
	return false, nil
}

// New creates an Oracle client
func New() (_ Client, err error) {
	c := &client{}
	c.configProvider = oraclecommon.DefaultConfigProvider()
	if c.identityClient, err = oracleidentity.NewIdentityClientWithConfigurationProvider(c.configProvider); err != nil {
		return nil, errors.Wrap(err, "unable to create identity client for Oracle Cloud Infrastructure")
	}
	if c.computeClient, err = oraclecore.NewComputeClientWithConfigurationProvider(c.configProvider); err != nil {
		return nil, errors.Wrap(err, "unable to create compute client for Oracle Cloud Infrastructure")
	}
	if c.workRequestClient, err = oracleworkrequests.NewWorkRequestClientWithConfigurationProvider(c.configProvider); err != nil {
		return nil, errors.Wrap(err, "unable to create work request client for Oracle Cloud Infrastructure")
	}
	return c, nil
}
