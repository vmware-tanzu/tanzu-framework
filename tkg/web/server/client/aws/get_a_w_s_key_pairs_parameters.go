// Code generated by go-swagger; DO NOT EDIT.

package aws

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
)

// NewGetAWSKeyPairsParams creates a new GetAWSKeyPairsParams object
// with the default values initialized.
func NewGetAWSKeyPairsParams() *GetAWSKeyPairsParams {

	return &GetAWSKeyPairsParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewGetAWSKeyPairsParamsWithTimeout creates a new GetAWSKeyPairsParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewGetAWSKeyPairsParamsWithTimeout(timeout time.Duration) *GetAWSKeyPairsParams {

	return &GetAWSKeyPairsParams{

		timeout: timeout,
	}
}

// NewGetAWSKeyPairsParamsWithContext creates a new GetAWSKeyPairsParams object
// with the default values initialized, and the ability to set a context for a request
func NewGetAWSKeyPairsParamsWithContext(ctx context.Context) *GetAWSKeyPairsParams {

	return &GetAWSKeyPairsParams{

		Context: ctx,
	}
}

// NewGetAWSKeyPairsParamsWithHTTPClient creates a new GetAWSKeyPairsParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewGetAWSKeyPairsParamsWithHTTPClient(client *http.Client) *GetAWSKeyPairsParams {

	return &GetAWSKeyPairsParams{
		HTTPClient: client,
	}
}

/*
GetAWSKeyPairsParams contains all the parameters to send to the API endpoint
for the get a w s key pairs operation typically these are written to a http.Request
*/
type GetAWSKeyPairsParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the get a w s key pairs params
func (o *GetAWSKeyPairsParams) WithTimeout(timeout time.Duration) *GetAWSKeyPairsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get a w s key pairs params
func (o *GetAWSKeyPairsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get a w s key pairs params
func (o *GetAWSKeyPairsParams) WithContext(ctx context.Context) *GetAWSKeyPairsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get a w s key pairs params
func (o *GetAWSKeyPairsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get a w s key pairs params
func (o *GetAWSKeyPairsParams) WithHTTPClient(client *http.Client) *GetAWSKeyPairsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get a w s key pairs params
func (o *GetAWSKeyPairsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *GetAWSKeyPairsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
