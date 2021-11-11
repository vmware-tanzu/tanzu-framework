// Code generated by go-swagger; DO NOT EDIT.

package docker

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new docker API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for docker API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
ApplyTKGConfigForDocker applies the changes to t k g configuration file for docker
*/
func (a *Client) ApplyTKGConfigForDocker(params *ApplyTKGConfigForDockerParams) (*ApplyTKGConfigForDockerOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewApplyTKGConfigForDockerParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "applyTKGConfigForDocker",
		Method:             "POST",
		PathPattern:        "/api/providers/docker/tkgconfig",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ApplyTKGConfigForDockerReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ApplyTKGConfigForDockerOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for applyTKGConfigForDocker: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
CheckIfDockerDaemonAvailable checks if docker deamon is available
*/
func (a *Client) CheckIfDockerDaemonAvailable(params *CheckIfDockerDaemonAvailableParams) (*CheckIfDockerDaemonAvailableOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCheckIfDockerDaemonAvailableParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "checkIfDockerDaemonAvailable",
		Method:             "GET",
		PathPattern:        "/api/providers/docker/daemon",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CheckIfDockerDaemonAvailableReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CheckIfDockerDaemonAvailableOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for checkIfDockerDaemonAvailable: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
CreateDockerRegionalCluster creates docker regional cluster
*/
func (a *Client) CreateDockerRegionalCluster(params *CreateDockerRegionalClusterParams) (*CreateDockerRegionalClusterOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateDockerRegionalClusterParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "createDockerRegionalCluster",
		Method:             "POST",
		PathPattern:        "/api/providers/docker/create",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &CreateDockerRegionalClusterReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateDockerRegionalClusterOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createDockerRegionalCluster: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
ExportTKGConfigForDocker generates t k g configuration file for docker
*/
func (a *Client) ExportTKGConfigForDocker(params *ExportTKGConfigForDockerParams) (*ExportTKGConfigForDockerOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewExportTKGConfigForDockerParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "exportTKGConfigForDocker",
		Method:             "POST",
		PathPattern:        "/api/providers/docker/config/export",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &ExportTKGConfigForDockerReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ExportTKGConfigForDockerOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for exportTKGConfigForDocker: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
