// Code generated by go-swagger; DO NOT EDIT.

package vsphere

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// GenerateTKGConfigForVsphereOKCode is the HTTP code returned for type GenerateTKGConfigForVsphereOK
const GenerateTKGConfigForVsphereOKCode int = 200

/*GenerateTKGConfigForVsphereOK Generated TKG configuration successfully

swagger:response generateTKGConfigForVsphereOK
*/
type GenerateTKGConfigForVsphereOK struct {

	/*
	  In: Body
	*/
	Payload string `json:"body,omitempty"`
}

// NewGenerateTKGConfigForVsphereOK creates GenerateTKGConfigForVsphereOK with default headers values
func NewGenerateTKGConfigForVsphereOK() *GenerateTKGConfigForVsphereOK {

	return &GenerateTKGConfigForVsphereOK{}
}

// WithPayload adds the payload to the generate t k g config for vsphere o k response
func (o *GenerateTKGConfigForVsphereOK) WithPayload(payload string) *GenerateTKGConfigForVsphereOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the generate t k g config for vsphere o k response
func (o *GenerateTKGConfigForVsphereOK) SetPayload(payload string) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GenerateTKGConfigForVsphereOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GenerateTKGConfigForVsphereBadRequestCode is the HTTP code returned for type GenerateTKGConfigForVsphereBadRequest
const GenerateTKGConfigForVsphereBadRequestCode int = 400

/*GenerateTKGConfigForVsphereBadRequest Bad request

swagger:response generateTKGConfigForVsphereBadRequest
*/
type GenerateTKGConfigForVsphereBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGenerateTKGConfigForVsphereBadRequest creates GenerateTKGConfigForVsphereBadRequest with default headers values
func NewGenerateTKGConfigForVsphereBadRequest() *GenerateTKGConfigForVsphereBadRequest {

	return &GenerateTKGConfigForVsphereBadRequest{}
}

// WithPayload adds the payload to the generate t k g config for vsphere bad request response
func (o *GenerateTKGConfigForVsphereBadRequest) WithPayload(payload *models.Error) *GenerateTKGConfigForVsphereBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the generate t k g config for vsphere bad request response
func (o *GenerateTKGConfigForVsphereBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GenerateTKGConfigForVsphereBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GenerateTKGConfigForVsphereUnauthorizedCode is the HTTP code returned for type GenerateTKGConfigForVsphereUnauthorized
const GenerateTKGConfigForVsphereUnauthorizedCode int = 401

/*GenerateTKGConfigForVsphereUnauthorized Incorrect credentials

swagger:response generateTKGConfigForVsphereUnauthorized
*/
type GenerateTKGConfigForVsphereUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGenerateTKGConfigForVsphereUnauthorized creates GenerateTKGConfigForVsphereUnauthorized with default headers values
func NewGenerateTKGConfigForVsphereUnauthorized() *GenerateTKGConfigForVsphereUnauthorized {

	return &GenerateTKGConfigForVsphereUnauthorized{}
}

// WithPayload adds the payload to the generate t k g config for vsphere unauthorized response
func (o *GenerateTKGConfigForVsphereUnauthorized) WithPayload(payload *models.Error) *GenerateTKGConfigForVsphereUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the generate t k g config for vsphere unauthorized response
func (o *GenerateTKGConfigForVsphereUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GenerateTKGConfigForVsphereUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GenerateTKGConfigForVsphereInternalServerErrorCode is the HTTP code returned for type GenerateTKGConfigForVsphereInternalServerError
const GenerateTKGConfigForVsphereInternalServerErrorCode int = 500

/*GenerateTKGConfigForVsphereInternalServerError Internal server error

swagger:response generateTKGConfigForVsphereInternalServerError
*/
type GenerateTKGConfigForVsphereInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGenerateTKGConfigForVsphereInternalServerError creates GenerateTKGConfigForVsphereInternalServerError with default headers values
func NewGenerateTKGConfigForVsphereInternalServerError() *GenerateTKGConfigForVsphereInternalServerError {

	return &GenerateTKGConfigForVsphereInternalServerError{}
}

// WithPayload adds the payload to the generate t k g config for vsphere internal server error response
func (o *GenerateTKGConfigForVsphereInternalServerError) WithPayload(payload *models.Error) *GenerateTKGConfigForVsphereInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the generate t k g config for vsphere internal server error response
func (o *GenerateTKGConfigForVsphereInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GenerateTKGConfigForVsphereInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
