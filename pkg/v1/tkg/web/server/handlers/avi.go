// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"

	aviclient "github.com/vmware-tanzu-private/core/pkg/v1/tkg/avi"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/web/server/restapi/operations/avi"
)

// VerifyAccount validates avi credentials and sets the avi client into web app
func (app *App) VerifyAccount(params avi.VerifyAccountParams) middleware.Responder {
	aviControllerParams := &models.AviControllerParams{
		Username: params.Credentials.Username,
		Password: params.Credentials.Password,
		Host:     params.Credentials.Host,
		Tenant:   params.Credentials.Tenant,
		CAData:   params.Credentials.CAData,
	}

	app.aviClient = aviclient.New()
	authed, err := app.aviClient.VerifyAccount(aviControllerParams)
	if err != nil {
		return avi.NewVerifyAccountInternalServerError().WithPayload(Err(err))
	}

	if !authed {
		return avi.NewVerifyAccountInternalServerError().WithPayload(Err(errors.Errorf("unable to authenticate due to incorrect credentials")))
	}

	return avi.NewVerifyAccountCreated()
}

// GetAviClouds handles requests to GET avi clouds
func (app *App) GetAviClouds(params avi.GetAviCloudsParams) middleware.Responder {
	if app.aviClient == nil {
		return avi.NewGetAviCloudsInternalServerError().WithPayload(Err(errors.New("avi client is not initialized properly")))
	}

	aviClouds, err := app.aviClient.GetClouds()
	if err != nil {
		return avi.NewGetAviCloudsInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get avi clouds")))
	}

	return avi.NewGetAviCloudsOK().WithPayload(aviClouds)
}

// GetAviServiceEngineGroups handles requests to GET avi service engine groups
func (app *App) GetAviServiceEngineGroups(params avi.GetAviServiceEngineGroupsParams) middleware.Responder {
	if app.aviClient == nil {
		return avi.NewGetAviServiceEngineGroupsInternalServerError().WithPayload(Err(errors.New("avi client is not initialized properly")))
	}

	aviServiceEngineGroups, err := app.aviClient.GetServiceEngineGroups()
	if err != nil {
		return avi.NewGetAviServiceEngineGroupsInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get avi service engine groups")))
	}

	return avi.NewGetAviServiceEngineGroupsOK().WithPayload(aviServiceEngineGroups)
}

// GetAviVipNetworks handles requests to GET avi VIP networks
func (app *App) GetAviVipNetworks(params avi.GetAviVipNetworksParams) middleware.Responder {
	if app.aviClient == nil {
		return avi.NewGetAviVipNetworksInternalServerError().WithPayload(Err(errors.New("avi client is not initialized properly")))
	}

	aviVipNetworks, err := app.aviClient.GetVipNetworks()
	if err != nil {
		return avi.NewGetAviVipNetworksInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get avi VIP networks")))
	}

	return avi.NewGetAviVipNetworksOK().WithPayload(aviVipNetworks)
}
