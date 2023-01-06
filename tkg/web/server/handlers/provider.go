// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/go-openapi/runtime/middleware"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgconfigbom"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
	"github.com/vmware-tanzu/tanzu-framework/tkg/web/server/restapi/operations/provider"
	"k8s.io/klog/v2/klogr"
)

// GetProvider gets provider information
func (app *App) GetProvider(params provider.GetProviderParams) middleware.Responder {
	defaultTKRBom, err := tkgconfigbom.New(app.AppConfig.TKGConfigDir, app.TKGConfigReaderWriter, klogr.New().WithName("provider-handler")).GetDefaultTkrBOMConfiguration()
	if err != nil {
		return provider.NewGetProviderInternalServerError().WithPayload(Err(errors.Wrap(err, "unable to get the default TanzuKubernetesRelease")))
	}

	providerInfo := models.ProviderInfo{
		Provider:   app.InitOptions.InfrastructureProvider,
		TkrVersion: defaultTKRBom.Release.Version,
	}

	if providerInfo.Provider == "" {
		providerInfo.Provider = "unknown"
	}

	return provider.NewGetProviderOK().WithPayload(&providerInfo)
}
