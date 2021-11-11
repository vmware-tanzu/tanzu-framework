// Code generated by go-swagger; DO NOT EDIT.

package aws

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new aws API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for aws API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
ApplyTKGConfigForAWS applies the changes to t k g configuration file for a w s
*/
func (a *Client) ApplyTKGConfigForAWS(params *ApplyTKGConfigForAWSParams) (*ApplyTKGConfigForAWSOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewApplyTKGConfigForAWSParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "applyTKGConfigForAWS",
		Method:             "POST",
		PathPattern:        "/api/providers/aws/tkgconfig",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ApplyTKGConfigForAWSReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ApplyTKGConfigForAWSOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for applyTKGConfigForAWS: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
CreateAWSRegionalCluster creates a w s regional cluster
*/
func (a *Client) CreateAWSRegionalCluster(params *CreateAWSRegionalClusterParams) (*CreateAWSRegionalClusterOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateAWSRegionalClusterParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createAWSRegionalCluster",
		Method:             "POST",
		PathPattern:        "/api/providers/aws/create",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateAWSRegionalClusterReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateAWSRegionalClusterOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createAWSRegionalCluster: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ExportTKGConfigForAWS generates t k g configuration file for a w s
*/
func (a *Client) ExportTKGConfigForAWS(params *ExportTKGConfigForAWSParams) (*ExportTKGConfigForAWSOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewExportTKGConfigForAWSParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "exportTKGConfigForAWS",
		Method:             "POST",
		PathPattern:        "/api/providers/aws/config/export",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ExportTKGConfigForAWSReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ExportTKGConfigForAWSOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for exportTKGConfigForAWS: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAWSAvailabilityZones retrieves a w s availability zones of current region
*/
func (a *Client) GetAWSAvailabilityZones(params *GetAWSAvailabilityZonesParams) (*GetAWSAvailabilityZonesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAWSAvailabilityZonesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAWSAvailabilityZones",
		Method:             "GET",
		PathPattern:        "/api/providers/aws/AvailabilityZones",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAWSAvailabilityZonesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAWSAvailabilityZonesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAWSAvailabilityZones: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAWSCredentialProfiles retrieves a w s credential profiles
*/
func (a *Client) GetAWSCredentialProfiles(params *GetAWSCredentialProfilesParams) (*GetAWSCredentialProfilesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAWSCredentialProfilesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAWSCredentialProfiles",
		Method:             "GET",
		PathPattern:        "/api/providers/aws/profiles",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAWSCredentialProfilesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAWSCredentialProfilesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAWSCredentialProfiles: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAWSEndpoint retrieves aws account params from environment variables
*/
func (a *Client) GetAWSEndpoint(params *GetAWSEndpointParams) (*GetAWSEndpointOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAWSEndpointParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAWSEndpoint",
		Method:             "GET",
		PathPattern:        "/api/providers/aws",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAWSEndpointReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAWSEndpointOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAWSEndpoint: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAWSNodeTypes retrieves a w s supported node types
*/
func (a *Client) GetAWSNodeTypes(params *GetAWSNodeTypesParams) (*GetAWSNodeTypesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAWSNodeTypesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAWSNodeTypes",
		Method:             "GET",
		PathPattern:        "/api/providers/aws/nodetypes",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAWSNodeTypesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAWSNodeTypesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAWSNodeTypes: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAWSOSImages retrieves a w s supported os images
*/
func (a *Client) GetAWSOSImages(params *GetAWSOSImagesParams) (*GetAWSOSImagesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAWSOSImagesParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAWSOSImages",
		Method:             "GET",
		PathPattern:        "/api/providers/aws/osimages",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAWSOSImagesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAWSOSImagesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAWSOSImages: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAWSRegions retrieves a w s regions
*/
func (a *Client) GetAWSRegions(params *GetAWSRegionsParams) (*GetAWSRegionsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAWSRegionsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAWSRegions",
		Method:             "GET",
		PathPattern:        "/api/providers/aws/regions",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAWSRegionsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAWSRegionsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAWSRegions: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAWSSubnets retrieves a w s subnets info under a v p c
*/
func (a *Client) GetAWSSubnets(params *GetAWSSubnetsParams) (*GetAWSSubnetsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAWSSubnetsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAWSSubnets",
		Method:             "GET",
		PathPattern:        "/api/providers/aws/subnets",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAWSSubnetsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAWSSubnetsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAWSSubnets: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetVPCs retrieves a w s v p cs
*/
func (a *Client) GetVPCs(params *GetVPCsParams) (*GetVPCsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetVPCsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getVPCs",
		Method:             "GET",
		PathPattern:        "/api/providers/aws/vpc",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetVPCsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetVPCsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getVPCs: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
SetAWSEndpoint validates and set aws credentials
*/
func (a *Client) SetAWSEndpoint(params *SetAWSEndpointParams) (*SetAWSEndpointCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewSetAWSEndpointParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "setAWSEndpoint",
		Method:             "POST",
		PathPattern:        "/api/providers/aws",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &SetAWSEndpointReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*SetAWSEndpointCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for setAWSEndpoint: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
