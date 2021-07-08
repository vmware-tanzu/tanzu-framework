// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// This file is safe to edit. Once it exists it will not be overwritten

package restapi

import (
	"crypto/tls"
	"net/http"

	errors "github.com/go-openapi/errors"
	runtime "github.com/go-openapi/runtime"
	middleware "github.com/go-openapi/runtime/middleware"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/aws"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/features"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/provider"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/ui"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/restapi/operations/vsphere"
)

func configureFlags(api *operations.KickstartUIAPI) {
	// api.CommandLineOptionsGroups = []swag.CommandLineOptionsGroup{ ... }
}

func configureAPI(api *operations.KickstartUIAPI) http.Handler { // nolint:funlen,gocyclo
	// configure the api here
	api.ServeError = errors.ServeError

	// Set your custom logger if needed. Default one is log.Printf
	// Expected interface func(string, ...interface{})
	//
	// Example:
	// api.Logger = log.Printf

	api.Logger = log.Infof

	api.JSONConsumer = runtime.JSONConsumer()

	api.JSONProducer = runtime.JSONProducer()

	if api.AwsCreateAWSRegionalClusterHandler == nil {
		api.AwsCreateAWSRegionalClusterHandler = aws.CreateAWSRegionalClusterHandlerFunc(func(params aws.CreateAWSRegionalClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation aws.CreateAWSRegionalCluster has not yet been implemented")
		})
	}
	if api.VsphereCreateVSphereRegionalClusterHandler == nil {
		api.VsphereCreateVSphereRegionalClusterHandler = vsphere.CreateVSphereRegionalClusterHandlerFunc(func(params vsphere.CreateVSphereRegionalClusterParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.CreateVSphereRegionalCluster has not yet been implemented")
		})
	}
	if api.AwsGetAWSAvailabilityZonesHandler == nil {
		api.AwsGetAWSAvailabilityZonesHandler = aws.GetAWSAvailabilityZonesHandlerFunc(func(params aws.GetAWSAvailabilityZonesParams) middleware.Responder {
			return middleware.NotImplemented("operation aws.GetAWSAvailabilityZones has not yet been implemented")
		})
	}
	if api.AwsGetAWSNodeTypesHandler == nil {
		api.AwsGetAWSNodeTypesHandler = aws.GetAWSNodeTypesHandlerFunc(func(params aws.GetAWSNodeTypesParams) middleware.Responder {
			return middleware.NotImplemented("operation aws.GetAWSNodeTypes has not yet been implemented")
		})
	}
	if api.AwsGetAWSOSImagesHandler == nil {
		api.AwsGetAWSOSImagesHandler = aws.GetAWSOSImagesHandlerFunc(func(params aws.GetAWSOSImagesParams) middleware.Responder {
			return middleware.NotImplemented("operation aws.GetAWSOSImages has not yet been implemented")
		})
	}
	if api.AwsGetAWSRegionsHandler == nil {
		api.AwsGetAWSRegionsHandler = aws.GetAWSRegionsHandlerFunc(func(params aws.GetAWSRegionsParams) middleware.Responder {
			return middleware.NotImplemented("operation aws.GetAWSRegions has not yet been implemented")
		})
	}
	if api.ProviderGetProviderHandler == nil {
		api.ProviderGetProviderHandler = provider.GetProviderHandlerFunc(func(params provider.GetProviderParams) middleware.Responder {
			return middleware.NotImplemented("operation provider.GetProvider has not yet been implemented")
		})
	}
	if api.UIGetUIHandler == nil {
		api.UIGetUIHandler = ui.GetUIHandlerFunc(func(params ui.GetUIParams) middleware.Responder {
			return middleware.NotImplemented("operation ui.GetUI has not yet been implemented")
		})
	}
	if api.UIGetUIFileHandler == nil {
		api.UIGetUIFileHandler = ui.GetUIFileHandlerFunc(func(params ui.GetUIFileParams) middleware.Responder {
			return middleware.NotImplemented("operation ui.GetUIFile has not yet been implemented")
		})
	}
	if api.AwsGetVPCsHandler == nil {
		api.AwsGetVPCsHandler = aws.GetVPCsHandlerFunc(func(params aws.GetVPCsParams) middleware.Responder {
			return middleware.NotImplemented("operation aws.GetVPCs has not yet been implemented")
		})
	}
	if api.VsphereGetVSphereDatacentersHandler == nil {
		api.VsphereGetVSphereDatacentersHandler = vsphere.GetVSphereDatacentersHandlerFunc(func(params vsphere.GetVSphereDatacentersParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.GetVSphereDatacenters has not yet been implemented")
		})
	}
	if api.VsphereGetVSphereDatastoresHandler == nil {
		api.VsphereGetVSphereDatastoresHandler = vsphere.GetVSphereDatastoresHandlerFunc(func(params vsphere.GetVSphereDatastoresParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.GetVSphereDatastores has not yet been implemented")
		})
	}
	if api.VsphereGetVSphereNetworksHandler == nil {
		api.VsphereGetVSphereNetworksHandler = vsphere.GetVSphereNetworksHandlerFunc(func(params vsphere.GetVSphereNetworksParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.GetVSphereNetworks has not yet been implemented")
		})
	}
	if api.VsphereGetVSphereNodeTypesHandler == nil {
		api.VsphereGetVSphereNodeTypesHandler = vsphere.GetVSphereNodeTypesHandlerFunc(func(params vsphere.GetVSphereNodeTypesParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.GetVSphereNodeTypes has not yet been implemented")
		})
	}
	if api.VsphereGetVSphereOSImagesHandler == nil {
		api.VsphereGetVSphereOSImagesHandler = vsphere.GetVSphereOSImagesHandlerFunc(func(params vsphere.GetVSphereOSImagesParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.GetVSphereOSImages has not yet been implemented")
		})
	}
	if api.VsphereGetVSphereResourcePoolsHandler == nil {
		api.VsphereGetVSphereResourcePoolsHandler = vsphere.GetVSphereResourcePoolsHandlerFunc(func(params vsphere.GetVSphereResourcePoolsParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.GetVSphereResourcePools has not yet been implemented")
		})
	}
	if api.AwsSetAWSEndpointHandler == nil {
		api.AwsSetAWSEndpointHandler = aws.SetAWSEndpointHandlerFunc(func(params aws.SetAWSEndpointParams) middleware.Responder {
			return middleware.NotImplemented("operation aws.SetAWSEndpoint has not yet been implemented")
		})
	}
	if api.VsphereSetVSphereEndpointHandler == nil {
		api.VsphereSetVSphereEndpointHandler = vsphere.SetVSphereEndpointHandlerFunc(func(params vsphere.SetVSphereEndpointParams) middleware.Responder {
			return middleware.NotImplemented("operation vsphere.SetVSphereEndpoint has not yet been implemented")
		})
	}

	if api.FeaturesGetFeatureFlagsHandler == nil {
		api.FeaturesGetFeatureFlagsHandler = features.GetFeatureFlagsHandlerFunc(func(params features.GetFeatureFlagsParams) middleware.Responder {
			return middleware.NotImplemented("operation features.GetFeatureFlags has not yet been implemented")
		})
	}

	api.ServerShutdown = func() {}

	return setupGlobalMiddleware(api.Serve(setupMiddlewares))
}

// The TLS configuration before HTTPS server starts.
func configureTLS(tlsConfig *tls.Config) {
	// Make all necessary changes to the TLS configuration here.
}

// As soon as server is initialized but not run yet, this function will be called.
// If you need to modify a config, store server instance to stop it individually later, this is the place.
// This function can be called multiple times, depending on the number of serving schemes.
// scheme value will be set accordingly: "http", "https" or "unix"
func configureServer(s *http.Server, scheme, addr string) {
}

// The middleware configuration is for the handler executors. These do not apply to the swagger.json document.
// The middleware executes after routing but before authentication, binding and validation
func setupMiddlewares(handler http.Handler) http.Handler {
	return handler
}

// The middleware configuration happens before anything, this middleware also applies to serving the swagger.json document.
// So this is a good place to plug in a panic handling middleware, logging and metrics
func setupGlobalMiddleware(handler http.Handler) http.Handler {
	return handler
}
