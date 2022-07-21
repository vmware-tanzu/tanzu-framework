// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package avi

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/avinetworks/sdk/go/clients"
	"github.com/avinetworks/sdk/go/models"
	"github.com/avinetworks/sdk/go/session"
	"github.com/pkg/errors"

	avi_models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// AviSessionTimeout is timeout for avi session
const AviSessionTimeout = 60
const pageSizeMax = "200"
const aviDefaultTenant = "admin" // Per TKG-5862

type client struct {
	ControllerParams   *avi_models.AviControllerParams
	Cloud              MiniCloudClient
	ServiceEngineGroup MiniServiceEngineGroupClient
	Network            MiniNetworkClient
}

// New creates an AVI controller REST API client
func New() Client {
	return &client{
		ControllerParams:   nil,
		Cloud:              nil,
		ServiceEngineGroup: nil,
		Network:            nil,
	}
}

// VerifyAccount verifies if the credentials are correct
// It also setup the Cloud and SerivceEngineGroup services for later use
// upon a successful authentication
func (c *client) VerifyAccount(params *avi_models.AviControllerParams) (bool, error) {
	var aviClient *clients.AviClient
	var err error

	if params.CAData == "" {
		aviClient, err = clients.NewAviClient(params.Host, params.Username,
			session.SetPassword(params.Password),
			session.SetTenant(aviDefaultTenant))
	} else {
		var transport *http.Transport
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM([]byte(params.CAData))
		transport = &http.Transport{
			TLSClientConfig: &tls.Config{ //nolint:gosec
				RootCAs: caCertPool,
			},
		}

		options := []func(*session.AviSession) error{
			session.SetPassword(params.Password),
			session.SetTenant(aviDefaultTenant),
			session.SetControllerStatusCheckLimits(1, 1),
			session.DisableControllerStatusCheckOnFailure(true),
			session.SetTimeout(AviSessionTimeout * time.Second),
			session.SetTransport(transport),
		}

		aviClient, err = clients.NewAviClient(params.Host, params.Username, options...)
	}

	if err != nil {
		return false, errors.Wrap(err, "unable to create API client using the credentials provided")
	}

	apiVersion, err := aviClient.AviSession.GetControllerVersion()
	if err != nil {
		return false, errors.Wrap(err, "unable to get API version")
	}

	SetVersion := session.SetVersion(apiVersion)
	if err := SetVersion(aviClient.AviSession); err != nil {
		return false, errors.Wrap(err, "unable to set API version")
	}

	c.ControllerParams = params
	c.Cloud = aviClient.Cloud
	c.ServiceEngineGroup = aviClient.ServiceEngineGroup
	c.Network = aviClient.Network

	return true, nil
}

// GetClouds retrieves a cloud list from AVI controller through the REST API
// This function depends on the presence of "Cloud" service that is
// made available upon authentication with a Avi controller.
func (c *client) GetClouds() ([]*avi_models.AviCloud, error) {
	if c.Cloud == nil {
		return nil, errors.Errorf("unable to make API calls before authentication")
	}

	var page = 1
	clouds := make([]*avi_models.AviCloud, 0)
	for {
		all, err := c.Cloud.GetAll(session.SetParams(map[string]string{"fields": "name,uuid", "page": strconv.Itoa(page), "page_size": pageSizeMax}))
		if err != nil {
			if page == 1 {
				return nil, errors.Wrap(err, "unable to get all clouds from avi controller due to error")
			}
			break // end of result set reached
		}

		for _, c := range all {
			clouds = append(clouds, &avi_models.AviCloud{
				UUID:     *c.UUID,
				Name:     *c.Name,
				Location: *c.URL,
			})
		}

		page++
	}

	return clouds, nil
}

// GetServiceEngineGroups retrieves a Service Engine Group list from AVI controller through the REST API
// This function depends on the presence of "ServiceEngineGroup" service that is
// made available upon authentication with a Avi controller.
func (c *client) GetServiceEngineGroups() ([]*avi_models.AviServiceEngineGroup, error) {
	if c.ServiceEngineGroup == nil {
		return nil, errors.Errorf("unable to make API calls before authentication")
	}

	var page = 1
	serviceEngineGroups := make([]*avi_models.AviServiceEngineGroup, 0)

	for {
		all, err := c.ServiceEngineGroup.GetAll(session.SetParams(map[string]string{"fields": "name,uuid,cloud_ref", "page": strconv.Itoa(page), "page_size": pageSizeMax}))
		if err != nil {
			if page == 1 {
				return nil, errors.Wrap(err, "unable to get all Service Engine Groups from avi controller due to error")
			}
			break
		}

		for _, seg := range all {
			serviceEngineGroups = append(serviceEngineGroups, &avi_models.AviServiceEngineGroup{
				UUID:     *seg.UUID,
				Name:     *seg.Name,
				Location: c.getCloudID(*seg.CloudRef),
			})
		}

		page++
	}
	return serviceEngineGroups, nil
}

// GetVipNetworks retrieves a Service Engine Group list from AVI controller through the REST API
// This function depends on the presence of "ServiceEngineGroup" service that is
// made available upon authentication with a Avi controller.
func (c *client) GetVipNetworks() ([]*avi_models.AviVipNetwork, error) {
	if c.Network == nil {
		return nil, errors.Errorf("unable to make API calls before authentication")
	}

	var page = 1
	networks := make([]*avi_models.AviVipNetwork, 0)
	for {
		all, err := c.Network.GetAll(session.SetParams(map[string]string{"fields": "name,uuid,cloud_ref,configured_subnets", "page": strconv.Itoa(page), "page_size": pageSizeMax}))
		if err != nil {
			if page == 1 {
				return nil, errors.Wrap(err, "unable to get all Networks from avi controller due to error")
			}
			break
		}

		for _, seg := range all {
			subnets := make([]*avi_models.AviSubnet, 0)
			for _, temp := range seg.ConfiguredSubnets {
				subnets = append(subnets, &avi_models.AviSubnet{
					Subnet: *temp.Prefix.IPAddr.Addr + "/" + strconv.Itoa(int(*temp.Prefix.Mask)),
					Family: *temp.Prefix.IPAddr.Type,
				})
			}

			network := &avi_models.AviVipNetwork{
				UUID:            *seg.UUID,
				Name:            *seg.Name,
				Cloud:           c.getCloudID(*seg.CloudRef),
				ConfigedSubnets: subnets,
			}

			networks = append(networks, network)
		}

		page++
	}
	return networks, nil
}

func (c *client) GetCloudByName(name string) (*models.Cloud, error) {
	return c.Cloud.GetByName(name)
}

func (c *client) GetServiceEngineGroupByName(name string) (*models.ServiceEngineGroup, error) {
	return c.ServiceEngineGroup.GetByName(name)
}

func (c *client) GetVipNetworkByName(name string) (*models.Network, error) {
	return c.Network.GetByName(name)
}

// getCloudID extracts the cloud UUID from the cloudRef string,
// which would be in the format of "https://10.78.115.106/api/cloud/cloud-e8b6d313-ecac-40f8-b38c-440e84c6d731",
// whereas the last part of the URL is the cloud UUID that the Service Engine Group belongs to.
// cloudRef is guarrantted to to be a valid string as indicated below when called in the context.
func (c *client) getCloudID(cloudRef string) string {
	parts := strings.Split(cloudRef, "/")
	return parts[len(parts)-1]
}
