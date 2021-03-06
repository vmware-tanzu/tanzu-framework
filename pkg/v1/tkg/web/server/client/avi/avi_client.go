// Code generated by go-swagger; DO NOT EDIT.

package avi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"
)

// New creates a new avi API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for avi API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
GetAviClouds retrieves avi load balancer clouds
*/
func (a *Client) GetAviClouds(params *GetAviCloudsParams) (*GetAviCloudsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAviCloudsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAviClouds",
		Method:             "GET",
		PathPattern:        "/api/avi/clouds",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAviCloudsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAviCloudsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAviClouds: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAviServiceEngineGroups retrieves avi load balancer service engine groups
*/
func (a *Client) GetAviServiceEngineGroups(params *GetAviServiceEngineGroupsParams) (*GetAviServiceEngineGroupsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAviServiceEngineGroupsParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAviServiceEngineGroups",
		Method:             "GET",
		PathPattern:        "/api/avi/serviceenginegroups",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAviServiceEngineGroupsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAviServiceEngineGroupsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAviServiceEngineGroups: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
GetAviVipNetworks retrieves all avi networks
*/
func (a *Client) GetAviVipNetworks(params *GetAviVipNetworksParams) (*GetAviVipNetworksOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAviVipNetworksParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "getAviVipNetworks",
		Method:             "GET",
		PathPattern:        "/api/avi/vipnetworks",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &GetAviVipNetworksReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAviVipNetworksOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getAviVipNetworks: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
VerifyAccount validates avi controller credentials
*/
func (a *Client) VerifyAccount(params *VerifyAccountParams) (*VerifyAccountCreated, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewVerifyAccountParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "verifyAccount",
		Method:             "POST",
		PathPattern:        "/api/avi",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &VerifyAccountReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*VerifyAccountCreated)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for verifyAccount: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
