// Code generated by go-swagger; DO NOT EDIT.

package docker

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// ExportTKGConfigForDockerOKCode is the HTTP code returned for type ExportTKGConfigForDockerOK
const ExportTKGConfigForDockerOKCode int = 200

/*
ExportTKGConfigForDockerOK Generated TKG configuration successfully

swagger:response exportTKGConfigForDockerOK
*/
type ExportTKGConfigForDockerOK struct {

	/*
	  In: Body
	*/
	Payload string `json:"body,omitempty"`
}

// NewExportTKGConfigForDockerOK creates ExportTKGConfigForDockerOK with default headers values
func NewExportTKGConfigForDockerOK() *ExportTKGConfigForDockerOK {

	return &ExportTKGConfigForDockerOK{}
}

// WithPayload adds the payload to the export t k g config for docker o k response
func (o *ExportTKGConfigForDockerOK) WithPayload(payload string) *ExportTKGConfigForDockerOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the export t k g config for docker o k response
func (o *ExportTKGConfigForDockerOK) SetPayload(payload string) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ExportTKGConfigForDockerOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// ExportTKGConfigForDockerBadRequestCode is the HTTP code returned for type ExportTKGConfigForDockerBadRequest
const ExportTKGConfigForDockerBadRequestCode int = 400

/*
ExportTKGConfigForDockerBadRequest Bad request

swagger:response exportTKGConfigForDockerBadRequest
*/
type ExportTKGConfigForDockerBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewExportTKGConfigForDockerBadRequest creates ExportTKGConfigForDockerBadRequest with default headers values
func NewExportTKGConfigForDockerBadRequest() *ExportTKGConfigForDockerBadRequest {

	return &ExportTKGConfigForDockerBadRequest{}
}

// WithPayload adds the payload to the export t k g config for docker bad request response
func (o *ExportTKGConfigForDockerBadRequest) WithPayload(payload *models.Error) *ExportTKGConfigForDockerBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the export t k g config for docker bad request response
func (o *ExportTKGConfigForDockerBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ExportTKGConfigForDockerBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ExportTKGConfigForDockerInternalServerErrorCode is the HTTP code returned for type ExportTKGConfigForDockerInternalServerError
const ExportTKGConfigForDockerInternalServerErrorCode int = 500

/*
ExportTKGConfigForDockerInternalServerError Internal server error

swagger:response exportTKGConfigForDockerInternalServerError
*/
type ExportTKGConfigForDockerInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewExportTKGConfigForDockerInternalServerError creates ExportTKGConfigForDockerInternalServerError with default headers values
func NewExportTKGConfigForDockerInternalServerError() *ExportTKGConfigForDockerInternalServerError {

	return &ExportTKGConfigForDockerInternalServerError{}
}

// WithPayload adds the payload to the export t k g config for docker internal server error response
func (o *ExportTKGConfigForDockerInternalServerError) WithPayload(payload *models.Error) *ExportTKGConfigForDockerInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the export t k g config for docker internal server error response
func (o *ExportTKGConfigForDockerInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ExportTKGConfigForDockerInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
