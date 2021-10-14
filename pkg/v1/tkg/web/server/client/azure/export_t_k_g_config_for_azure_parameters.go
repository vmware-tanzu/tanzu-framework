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

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// NewExportTKGConfigForAzureParams creates a new ExportTKGConfigForAzureParams object
// with the default values initialized.
func NewExportTKGConfigForAzureParams() *ExportTKGConfigForAzureParams {
	var ()
	return &ExportTKGConfigForAzureParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewExportTKGConfigForAzureParamsWithTimeout creates a new ExportTKGConfigForAzureParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewExportTKGConfigForAzureParamsWithTimeout(timeout time.Duration) *ExportTKGConfigForAzureParams {
	var ()
	return &ExportTKGConfigForAzureParams{

		timeout: timeout,
	}
}

// NewExportTKGConfigForAzureParamsWithContext creates a new ExportTKGConfigForAzureParams object
// with the default values initialized, and the ability to set a context for a request
func NewExportTKGConfigForAzureParamsWithContext(ctx context.Context) *ExportTKGConfigForAzureParams {
	var ()
	return &ExportTKGConfigForAzureParams{

		Context: ctx,
	}
}

// NewExportTKGConfigForAzureParamsWithHTTPClient creates a new ExportTKGConfigForAzureParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewExportTKGConfigForAzureParamsWithHTTPClient(client *http.Client) *ExportTKGConfigForAzureParams {
	var ()
	return &ExportTKGConfigForAzureParams{
		HTTPClient: client,
	}
}

/*ExportTKGConfigForAzureParams contains all the parameters to send to the API endpoint
for the export t k g config for azure operation typically these are written to a http.Request
*/
type ExportTKGConfigForAzureParams struct {

	/*Params
	  parameters to generate TKG configuration file for Azure

	*/
	Params *models.AzureRegionalClusterParams

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) WithTimeout(timeout time.Duration) *ExportTKGConfigForAzureParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) WithContext(ctx context.Context) *ExportTKGConfigForAzureParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) WithHTTPClient(client *http.Client) *ExportTKGConfigForAzureParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithParams adds the params to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) WithParams(params *models.AzureRegionalClusterParams) *ExportTKGConfigForAzureParams {
	o.SetParams(params)
	return o
}

// SetParams adds the params to the export t k g config for azure params
func (o *ExportTKGConfigForAzureParams) SetParams(params *models.AzureRegionalClusterParams) {
	o.Params = params
}

// WriteToRequest writes these params to a swagger request
func (o *ExportTKGConfigForAzureParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if o.Params != nil {
		if err := r.SetBodyParam(o.Params); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
