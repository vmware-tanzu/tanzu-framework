// Code generated by go-swagger; DO NOT EDIT.

package vsphere

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new vsphere API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for vsphere API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
ApplyTKGConfigForVsphere applies changes to t k g configuration file for v sphere
*/
func (a *Client) ApplyTKGConfigForVsphere(params *ApplyTKGConfigForVsphereParams) (*ApplyTKGConfigForVsphereOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewApplyTKGConfigForVsphereParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "applyTKGConfigForVsphere",
		Method:             "POST",
		PathPattern:        "/api/providers/vsphere/tkgconfig",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ApplyTKGConfigForVsphereReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ApplyTKGConfigForVsphereOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for applyTKGConfigForVsphere: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
CreateVSphereRegionalCluster creates v sphere regional cluster
*/
func (a *Client) CreateVSphereRegionalCluster(params *CreateVSphereRegionalClusterParams) (*CreateVSphereRegionalClusterOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateVSphereRegionalClusterParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createVSphereRegionalCluster",
		Method:             "POST",
		PathPattern:        "/api/providers/vsphere/create",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateVSphereRegionalClusterReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateVSphereRegionalClusterOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createVSphereRegionalCluster: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ExportTKGConfigForVsphere generates t k g configuration file for v sphere
*/
func (a *Client) ExportTKGConfigForVsphere(params *ExportTKGConfigForVsphereParams) (*ExportTKGConfigForVsphereOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewExportTKGConfigForVsphereParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "exportTKGConfigForVsphere",
		Method:             "POST",
		PathPattern:        "/api/providers/vsphere/config/export",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ExportTKGConfigForVsphereReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ExportTKGConfigForVsphereOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for exportTKGConfigForVsphere: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereComputeResources retrieves v sphere compute resources
*/
func (a *Client) GetVSphereComputeResources(params *GetVSphereComputeResourcesParams) (*GetVSphereComputeResourcesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereComputeResourcesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereComputeResources",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/compute",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereComputeResourcesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereComputeResourcesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereComputeResources: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereDatacenters retrieves v sphere datacenters
*/
func (a *Client) GetVSphereDatacenters(params *GetVSphereDatacentersParams) (*GetVSphereDatacentersOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereDatacentersParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereDatacenters",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/datacenters",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereDatacentersReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereDatacentersOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereDatacenters: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereDatastores retrieves v sphere datastores
*/
func (a *Client) GetVSphereDatastores(params *GetVSphereDatastoresParams) (*GetVSphereDatastoresOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereDatastoresParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereDatastores",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/datastores",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereDatastoresReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereDatastoresOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereDatastores: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereFolders retrieves v sphere folders
*/
func (a *Client) GetVSphereFolders(params *GetVSphereFoldersParams) (*GetVSphereFoldersOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereFoldersParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereFolders",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/folders",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereFoldersReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereFoldersOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereFolders: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereNetworks retrieves networks associated with the datacenter in v sphere
*/
func (a *Client) GetVSphereNetworks(params *GetVSphereNetworksParams) (*GetVSphereNetworksOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereNetworksParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereNetworks",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/networks",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereNetworksReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereNetworksOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereNetworks: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereNodeTypes retrieves v sphere supported kubernetes control plane node types
*/
func (a *Client) GetVSphereNodeTypes(params *GetVSphereNodeTypesParams) (*GetVSphereNodeTypesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereNodeTypesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereNodeTypes",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/nodetypes",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereNodeTypesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereNodeTypesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereNodeTypes: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereOSImages retrieves v sphere supported node os images
*/
func (a *Client) GetVSphereOSImages(params *GetVSphereOSImagesParams) (*GetVSphereOSImagesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereOSImagesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereOSImages",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/osimages",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereOSImagesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereOSImagesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereOSImages: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVSphereResourcePools retrieves v sphere resource pools
*/
func (a *Client) GetVSphereResourcePools(params *GetVSphereResourcePoolsParams) (*GetVSphereResourcePoolsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVSphereResourcePoolsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVSphereResourcePools",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/resourcepools",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVSphereResourcePoolsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVSphereResourcePoolsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVSphereResourcePools: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVsphereThumbprint gets v sphere thumbprint
*/
func (a *Client) GetVsphereThumbprint(params *GetVsphereThumbprintParams) (*GetVsphereThumbprintOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVsphereThumbprintParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVsphereThumbprint",
		Method:             "GET",
		PathPattern:        "/api/providers/vsphere/thumbprint",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVsphereThumbprintReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVsphereThumbprintOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVsphereThumbprint: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ImportTKGConfigForVsphere generates t k g configuration object for v sphere
*/
func (a *Client) ImportTKGConfigForVsphere(params *ImportTKGConfigForVsphereParams) (*ImportTKGConfigForVsphereOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewImportTKGConfigForVsphereParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "importTKGConfigForVsphere",
		Method:             "POST",
		PathPattern:        "/api/providers/vsphere/config/import",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ImportTKGConfigForVsphereReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ImportTKGConfigForVsphereOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for importTKGConfigForVsphere: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
SetVSphereEndpoint validates and set v sphere credentials
*/
func (a *Client) SetVSphereEndpoint(params *SetVSphereEndpointParams) (*SetVSphereEndpointCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewSetVSphereEndpointParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "setVSphereEndpoint",
		Method:             "POST",
		PathPattern:        "/api/providers/vsphere",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &SetVSphereEndpointReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*SetVSphereEndpointCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for setVSphereEndpoint: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
