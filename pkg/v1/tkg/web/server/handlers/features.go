// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"

	featuresclient "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/features"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/edition"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/features"
)

// GetFeatureFlags handles requests to GET features
func (app *App) GetFeatureFlags(params features.GetFeatureFlagsParams) middleware.Responder {
	featuresClient, err := featuresclient.New(app.AppConfig.TKGConfigDir, "")
	if err != nil {
		return features.NewGetFeatureFlagsInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get feature flags client")))
	}

	featureFlags, err := featuresClient.GetFeatureFlags()
	if err != nil {
		return features.NewGetFeatureFlagsInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get feature flags")))
	}

	return features.NewGetFeatureFlagsOK().WithPayload(featureFlags)
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
