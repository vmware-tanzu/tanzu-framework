// Code generated by go-swagger; DO NOT EDIT.

package avi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// GetAviServiceEngineGroupsOKCode is the HTTP code returned for type GetAviServiceEngineGroupsOK
const GetAviServiceEngineGroupsOKCode int = 200

/*
GetAviServiceEngineGroupsOK Successful retrieval of Avi load balancer service engine groups

swagger:response getAviServiceEngineGroupsOK
*/
type GetAviServiceEngineGroupsOK struct {

	/*
	  In: Body
	*/
	Payload []*models.AviServiceEngineGroup `json:"body,omitempty"`
}

// NewGetAviServiceEngineGroupsOK creates GetAviServiceEngineGroupsOK with default headers values
func NewGetAviServiceEngineGroupsOK() *GetAviServiceEngineGroupsOK {

	return &GetAviServiceEngineGroupsOK{}
}

// WithPayload adds the payload to the get avi service engine groups o k response
func (o *GetAviServiceEngineGroupsOK) WithPayload(payload []*models.AviServiceEngineGroup) *GetAviServiceEngineGroupsOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi service engine groups o k response
func (o *GetAviServiceEngineGroupsOK) SetPayload(payload []*models.AviServiceEngineGroup) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviServiceEngineGroupsOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.AviServiceEngineGroup, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GetAviServiceEngineGroupsBadRequestCode is the HTTP code returned for type GetAviServiceEngineGroupsBadRequest
const GetAviServiceEngineGroupsBadRequestCode int = 400

/*
GetAviServiceEngineGroupsBadRequest Bad request

swagger:response getAviServiceEngineGroupsBadRequest
*/
type GetAviServiceEngineGroupsBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAviServiceEngineGroupsBadRequest creates GetAviServiceEngineGroupsBadRequest with default headers values
func NewGetAviServiceEngineGroupsBadRequest() *GetAviServiceEngineGroupsBadRequest {

	return &GetAviServiceEngineGroupsBadRequest{}
}

// WithPayload adds the payload to the get avi service engine groups bad request response
func (o *GetAviServiceEngineGroupsBadRequest) WithPayload(payload *models.Error) *GetAviServiceEngineGroupsBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi service engine groups bad request response
func (o *GetAviServiceEngineGroupsBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviServiceEngineGroupsBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetAviServiceEngineGroupsUnauthorizedCode is the HTTP code returned for type GetAviServiceEngineGroupsUnauthorized
const GetAviServiceEngineGroupsUnauthorizedCode int = 401

/*
GetAviServiceEngineGroupsUnauthorized Incorrect credentials

swagger:response getAviServiceEngineGroupsUnauthorized
*/
type GetAviServiceEngineGroupsUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAviServiceEngineGroupsUnauthorized creates GetAviServiceEngineGroupsUnauthorized with default headers values
func NewGetAviServiceEngineGroupsUnauthorized() *GetAviServiceEngineGroupsUnauthorized {

	return &GetAviServiceEngineGroupsUnauthorized{}
}

// WithPayload adds the payload to the get avi service engine groups unauthorized response
func (o *GetAviServiceEngineGroupsUnauthorized) WithPayload(payload *models.Error) *GetAviServiceEngineGroupsUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi service engine groups unauthorized response
func (o *GetAviServiceEngineGroupsUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviServiceEngineGroupsUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetAviServiceEngineGroupsInternalServerErrorCode is the HTTP code returned for type GetAviServiceEngineGroupsInternalServerError
const GetAviServiceEngineGroupsInternalServerErrorCode int = 500

/*
GetAviServiceEngineGroupsInternalServerError Internal server error

swagger:response getAviServiceEngineGroupsInternalServerError
*/
type GetAviServiceEngineGroupsInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAviServiceEngineGroupsInternalServerError creates GetAviServiceEngineGroupsInternalServerError with default headers values
func NewGetAviServiceEngineGroupsInternalServerError() *GetAviServiceEngineGroupsInternalServerError {

	return &GetAviServiceEngineGroupsInternalServerError{}
}

// WithPayload adds the payload to the get avi service engine groups internal server error response
func (o *GetAviServiceEngineGroupsInternalServerError) WithPayload(payload *models.Error) *GetAviServiceEngineGroupsInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi service engine groups internal server error response
func (o *GetAviServiceEngineGroupsInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviServiceEngineGroupsInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
