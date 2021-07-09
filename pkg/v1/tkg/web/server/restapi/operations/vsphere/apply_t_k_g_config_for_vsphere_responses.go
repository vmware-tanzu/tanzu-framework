// Code generated by go-swagger; DO NOT EDIT.

package vsphere

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// ApplyTKGConfigForVsphereOKCode is the HTTP code returned for type ApplyTKGConfigForVsphereOK
const ApplyTKGConfigForVsphereOKCode int = 200

/*ApplyTKGConfigForVsphereOK apply changes to TKG configuration file successfully

swagger:response applyTKGConfigForVsphereOK
*/
type ApplyTKGConfigForVsphereOK struct {

	/*
	  In: Body
	*/
	Payload *models.ConfigFileInfo `json:"body,omitempty"`
}

// NewApplyTKGConfigForVsphereOK creates ApplyTKGConfigForVsphereOK with default headers values
func NewApplyTKGConfigForVsphereOK() *ApplyTKGConfigForVsphereOK {

	return &ApplyTKGConfigForVsphereOK{}
}

// WithPayload adds the payload to the apply t k g config for vsphere o k response
func (o *ApplyTKGConfigForVsphereOK) WithPayload(payload *models.ConfigFileInfo) *ApplyTKGConfigForVsphereOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the apply t k g config for vsphere o k response
func (o *ApplyTKGConfigForVsphereOK) SetPayload(payload *models.ConfigFileInfo) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ApplyTKGConfigForVsphereOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ApplyTKGConfigForVsphereBadRequestCode is the HTTP code returned for type ApplyTKGConfigForVsphereBadRequest
const ApplyTKGConfigForVsphereBadRequestCode int = 400

/*ApplyTKGConfigForVsphereBadRequest Bad request

swagger:response applyTKGConfigForVsphereBadRequest
*/
type ApplyTKGConfigForVsphereBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewApplyTKGConfigForVsphereBadRequest creates ApplyTKGConfigForVsphereBadRequest with default headers values
func NewApplyTKGConfigForVsphereBadRequest() *ApplyTKGConfigForVsphereBadRequest {

	return &ApplyTKGConfigForVsphereBadRequest{}
}

// WithPayload adds the payload to the apply t k g config for vsphere bad request response
func (o *ApplyTKGConfigForVsphereBadRequest) WithPayload(payload *models.Error) *ApplyTKGConfigForVsphereBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the apply t k g config for vsphere bad request response
func (o *ApplyTKGConfigForVsphereBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ApplyTKGConfigForVsphereBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ApplyTKGConfigForVsphereUnauthorizedCode is the HTTP code returned for type ApplyTKGConfigForVsphereUnauthorized
const ApplyTKGConfigForVsphereUnauthorizedCode int = 401

/*ApplyTKGConfigForVsphereUnauthorized Incorrect credentials

swagger:response applyTKGConfigForVsphereUnauthorized
*/
type ApplyTKGConfigForVsphereUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewApplyTKGConfigForVsphereUnauthorized creates ApplyTKGConfigForVsphereUnauthorized with default headers values
func NewApplyTKGConfigForVsphereUnauthorized() *ApplyTKGConfigForVsphereUnauthorized {

	return &ApplyTKGConfigForVsphereUnauthorized{}
}

// WithPayload adds the payload to the apply t k g config for vsphere unauthorized response
func (o *ApplyTKGConfigForVsphereUnauthorized) WithPayload(payload *models.Error) *ApplyTKGConfigForVsphereUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the apply t k g config for vsphere unauthorized response
func (o *ApplyTKGConfigForVsphereUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ApplyTKGConfigForVsphereUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// ApplyTKGConfigForVsphereInternalServerErrorCode is the HTTP code returned for type ApplyTKGConfigForVsphereInternalServerError
const ApplyTKGConfigForVsphereInternalServerErrorCode int = 500

/*ApplyTKGConfigForVsphereInternalServerError Internal server error

swagger:response applyTKGConfigForVsphereInternalServerError
*/
type ApplyTKGConfigForVsphereInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewApplyTKGConfigForVsphereInternalServerError creates ApplyTKGConfigForVsphereInternalServerError with default headers values
func NewApplyTKGConfigForVsphereInternalServerError() *ApplyTKGConfigForVsphereInternalServerError {

	return &ApplyTKGConfigForVsphereInternalServerError{}
}

// WithPayload adds the payload to the apply t k g config for vsphere internal server error response
func (o *ApplyTKGConfigForVsphereInternalServerError) WithPayload(payload *models.Error) *ApplyTKGConfigForVsphereInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the apply t k g config for vsphere internal server error response
func (o *ApplyTKGConfigForVsphereInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *ApplyTKGConfigForVsphereInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
