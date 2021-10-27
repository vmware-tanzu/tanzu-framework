// Code generated by go-swagger; DO NOT EDIT.

package docker

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

// NewExportTKGConfigForDockerParams creates a new ExportTKGConfigForDockerParams object
// with the default values initialized.
func NewExportTKGConfigForDockerParams() *ExportTKGConfigForDockerParams {
	var ()
	return &ExportTKGConfigForDockerParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewExportTKGConfigForDockerParamsWithTimeout creates a new ExportTKGConfigForDockerParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewExportTKGConfigForDockerParamsWithTimeout(timeout time.Duration) *ExportTKGConfigForDockerParams {
	var ()
	return &ExportTKGConfigForDockerParams{

		timeout: timeout,
	}
}

// NewExportTKGConfigForDockerParamsWithContext creates a new ExportTKGConfigForDockerParams object
// with the default values initialized, and the ability to set a context for a request
func NewExportTKGConfigForDockerParamsWithContext(ctx context.Context) *ExportTKGConfigForDockerParams {
	var ()
	return &ExportTKGConfigForDockerParams{

		Context: ctx,
	}
}

// NewExportTKGConfigForDockerParamsWithHTTPClient creates a new ExportTKGConfigForDockerParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewExportTKGConfigForDockerParamsWithHTTPClient(client *http.Client) *ExportTKGConfigForDockerParams {
	var ()
	return &ExportTKGConfigForDockerParams{
		HTTPClient: client,
	}
}

/*ExportTKGConfigForDockerParams contains all the parameters to send to the API endpoint
for the export t k g config for docker operation typically these are written to a http.Request
*/
type ExportTKGConfigForDockerParams struct {

	/*Params
	  parameters to generate TKG configuration file for Docker

	*/
	Params *models.DockerRegionalClusterParams

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) WithTimeout(timeout time.Duration) *ExportTKGConfigForDockerParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) WithContext(ctx context.Context) *ExportTKGConfigForDockerParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) WithHTTPClient(client *http.Client) *ExportTKGConfigForDockerParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithParams adds the params to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) WithParams(params *models.DockerRegionalClusterParams) *ExportTKGConfigForDockerParams {
	o.SetParams(params)
	return o
}

// SetParams adds the params to the export t k g config for docker params
func (o *ExportTKGConfigForDockerParams) SetParams(params *models.DockerRegionalClusterParams) {
	o.Params = params
}

// WriteToRequest writes these params to a swagger request
func (o *ExportTKGConfigForDockerParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

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
