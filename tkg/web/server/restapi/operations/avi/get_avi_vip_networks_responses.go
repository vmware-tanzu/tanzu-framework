// Code generated by go-swagger; DO NOT EDIT.

package avi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// GetAviVipNetworksOKCode is the HTTP code returned for type GetAviVipNetworksOK
const GetAviVipNetworksOKCode int = 200

/*
GetAviVipNetworksOK Successful retrieval of Avi load balancer service engine groups

swagger:response getAviVipNetworksOK
*/
type GetAviVipNetworksOK struct {

	/*
	  In: Body
	*/
	Payload []*models.AviVipNetwork `json:"body,omitempty"`
}

// NewGetAviVipNetworksOK creates GetAviVipNetworksOK with default headers values
func NewGetAviVipNetworksOK() *GetAviVipNetworksOK {

	return &GetAviVipNetworksOK{}
}

// WithPayload adds the payload to the get avi vip networks o k response
func (o *GetAviVipNetworksOK) WithPayload(payload []*models.AviVipNetwork) *GetAviVipNetworksOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi vip networks o k response
func (o *GetAviVipNetworksOK) SetPayload(payload []*models.AviVipNetwork) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviVipNetworksOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty array
		payload = make([]*models.AviVipNetwork, 0, 50)
	}

	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// GetAviVipNetworksBadRequestCode is the HTTP code returned for type GetAviVipNetworksBadRequest
const GetAviVipNetworksBadRequestCode int = 400

/*
GetAviVipNetworksBadRequest Bad request

swagger:response getAviVipNetworksBadRequest
*/
type GetAviVipNetworksBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAviVipNetworksBadRequest creates GetAviVipNetworksBadRequest with default headers values
func NewGetAviVipNetworksBadRequest() *GetAviVipNetworksBadRequest {

	return &GetAviVipNetworksBadRequest{}
}

// WithPayload adds the payload to the get avi vip networks bad request response
func (o *GetAviVipNetworksBadRequest) WithPayload(payload *models.Error) *GetAviVipNetworksBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi vip networks bad request response
func (o *GetAviVipNetworksBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviVipNetworksBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetAviVipNetworksUnauthorizedCode is the HTTP code returned for type GetAviVipNetworksUnauthorized
const GetAviVipNetworksUnauthorizedCode int = 401

/*
GetAviVipNetworksUnauthorized Incorrect credentials

swagger:response getAviVipNetworksUnauthorized
*/
type GetAviVipNetworksUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAviVipNetworksUnauthorized creates GetAviVipNetworksUnauthorized with default headers values
func NewGetAviVipNetworksUnauthorized() *GetAviVipNetworksUnauthorized {

	return &GetAviVipNetworksUnauthorized{}
}

// WithPayload adds the payload to the get avi vip networks unauthorized response
func (o *GetAviVipNetworksUnauthorized) WithPayload(payload *models.Error) *GetAviVipNetworksUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi vip networks unauthorized response
func (o *GetAviVipNetworksUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviVipNetworksUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// GetAviVipNetworksInternalServerErrorCode is the HTTP code returned for type GetAviVipNetworksInternalServerError
const GetAviVipNetworksInternalServerErrorCode int = 500

/*
GetAviVipNetworksInternalServerError Internal server error

swagger:response getAviVipNetworksInternalServerError
*/
type GetAviVipNetworksInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewGetAviVipNetworksInternalServerError creates GetAviVipNetworksInternalServerError with default headers values
func NewGetAviVipNetworksInternalServerError() *GetAviVipNetworksInternalServerError {

	return &GetAviVipNetworksInternalServerError{}
}

// WithPayload adds the payload to the get avi vip networks internal server error response
func (o *GetAviVipNetworksInternalServerError) WithPayload(payload *models.Error) *GetAviVipNetworksInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the get avi vip networks internal server error response
func (o *GetAviVipNetworksInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *GetAviVipNetworksInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
