// Code generated by go-swagger; DO NOT EDIT.

package azure

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// NewCreateAzureVirtualNetworkParams creates a new CreateAzureVirtualNetworkParams object
// with the default values initialized.
func NewCreateAzureVirtualNetworkParams() *CreateAzureVirtualNetworkParams {
	var ()
	return &CreateAzureVirtualNetworkParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewCreateAzureVirtualNetworkParamsWithTimeout creates a new CreateAzureVirtualNetworkParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewCreateAzureVirtualNetworkParamsWithTimeout(timeout time.Duration) *CreateAzureVirtualNetworkParams {
	var ()
	return &CreateAzureVirtualNetworkParams{

		timeout: timeout,
	}
}

// NewCreateAzureVirtualNetworkParamsWithContext creates a new CreateAzureVirtualNetworkParams object
// with the default values initialized, and the ability to set a context for a request
func NewCreateAzureVirtualNetworkParamsWithContext(ctx context.Context) *CreateAzureVirtualNetworkParams {
	var ()
	return &CreateAzureVirtualNetworkParams{

		Context: ctx,
	}
}

// NewCreateAzureVirtualNetworkParamsWithHTTPClient creates a new CreateAzureVirtualNetworkParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewCreateAzureVirtualNetworkParamsWithHTTPClient(client *http.Client) *CreateAzureVirtualNetworkParams {
	var ()
	return &CreateAzureVirtualNetworkParams{
		HTTPClient: client,
	}
}

/*
CreateAzureVirtualNetworkParams contains all the parameters to send to the API endpoint
for the create azure virtual network operation typically these are written to a http.Request
*/
type CreateAzureVirtualNetworkParams struct {

	/*Params
	  parameters to create a new Azure Virtual network

	*/
	Params *models.AzureVirtualNetwork
	/*ResourceGroupName
	  Name of the Azure resource group

	*/
	ResourceGroupName string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) WithTimeout(timeout time.Duration) *CreateAzureVirtualNetworkParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) WithContext(ctx context.Context) *CreateAzureVirtualNetworkParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) WithHTTPClient(client *http.Client) *CreateAzureVirtualNetworkParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithParams adds the params to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) WithParams(params *models.AzureVirtualNetwork) *CreateAzureVirtualNetworkParams {
	o.SetParams(params)
	return o
}

// SetParams adds the params to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) SetParams(params *models.AzureVirtualNetwork) {
	o.Params = params
}

// WithResourceGroupName adds the resourceGroupName to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) WithResourceGroupName(resourceGroupName string) *CreateAzureVirtualNetworkParams {
	o.SetResourceGroupName(resourceGroupName)
	return o
}

// SetResourceGroupName adds the resourceGroupName to the create azure virtual network params
func (o *CreateAzureVirtualNetworkParams) SetResourceGroupName(resourceGroupName string) {
	o.ResourceGroupName = resourceGroupName
}

// WriteToRequest writes these params to a swagger request
func (o *CreateAzureVirtualNetworkParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Params != nil {
		if err := r.SetBodyParam(o.Params); err != nil {
			return err
		}
	}

	// path param resourceGroupName
	if err := r.SetPathParam("resourceGroupName", o.ResourceGroupName); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
