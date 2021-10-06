// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/config"
	featuresclient "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/features"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/edition"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/features"
)

// GetFeatureFlags handles requests to GET features
func (app *App) GetFeatureFlags(params features.GetFeatureFlagsParams) middleware.Responder {
	cfg, err := config.GetClientConfig()
	if err != nil {
		return features.NewGetFeatureFlagsInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get client configuration")))
	}
	payload := &models.Features{}
	payload.Cli = cfg.ClientOptions.Features
	payload.Plugins = make(map[string]models.FeatureMap)
	for pluginName, featureMap := range cfg.ClientOptions.Plugins {
		payload.Plugins[pluginName] = convertPluginFeatureMap(featureMap)
	}
	return features.NewGetFeatureFlagsOK().WithPayload(payload)
}

// Converts a v1alpha1.FeatureMap (that comes from the config file) to a models.FeatureMap (to be part of the endpoint's response payload)
func convertPluginFeatureMap(featureMap v1alpha1.FeatureMap) models.FeatureMap {
	result := models.FeatureMap{}
	for featureName, enabled := range featureMap.Features {
		(result)[featureName] = enabled
	}
	return result
}

// GetTanzuEdition returns the Tanzu edition
func (app *App) GetTanzuEdition(params edition.GetTanzuEditionParams) middleware.Responder {
	featuresClient, err := featuresclient.New(app.AppConfig.TKGConfigDir, "")
	if err != nil {
		return edition.NewGetTanzuEditionInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get feature flags client")))
	}

	tanzuEdition, err := featuresClient.GetFeatureFlag("edition")
	if err != nil {
		return edition.NewGetTanzuEditionInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get tanzu edition")))
	}

	return edition.NewGetTanzuEditionOK().WithPayload(tanzuEdition)
}
