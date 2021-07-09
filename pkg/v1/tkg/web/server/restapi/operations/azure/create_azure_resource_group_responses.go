// Code generated by go-swagger; DO NOT EDIT.

package azure

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// CreateAzureResourceGroupCreatedCode is the HTTP code returned for type CreateAzureResourceGroupCreated
const CreateAzureResourceGroupCreatedCode int = 201

/*CreateAzureResourceGroupCreated Successfully created Azure resource group

swagger:response createAzureResourceGroupCreated
*/
type CreateAzureResourceGroupCreated struct {

	/*
	  In: Body
	*/
	Payload string `json:"body,omitempty"`
}

// NewCreateAzureResourceGroupCreated creates CreateAzureResourceGroupCreated with default headers values
func NewCreateAzureResourceGroupCreated() *CreateAzureResourceGroupCreated {

	return &CreateAzureResourceGroupCreated{}
}

// WithPayload adds the payload to the create azure resource group created response
func (o *CreateAzureResourceGroupCreated) WithPayload(payload string) *CreateAzureResourceGroupCreated {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create azure resource group created response
func (o *CreateAzureResourceGroupCreated) SetPayload(payload string) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAzureResourceGroupCreated) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(201)
	payload := o.Payload
	if err := producer.Produce(rw, payload); err != nil {
		panic(err) // let the recovery middleware deal with this
	}
}

// CreateAzureResourceGroupBadRequestCode is the HTTP code returned for type CreateAzureResourceGroupBadRequest
const CreateAzureResourceGroupBadRequestCode int = 400

/*CreateAzureResourceGroupBadRequest Bad request

swagger:response createAzureResourceGroupBadRequest
*/
type CreateAzureResourceGroupBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateAzureResourceGroupBadRequest creates CreateAzureResourceGroupBadRequest with default headers values
func NewCreateAzureResourceGroupBadRequest() *CreateAzureResourceGroupBadRequest {

	return &CreateAzureResourceGroupBadRequest{}
}

// WithPayload adds the payload to the create azure resource group bad request response
func (o *CreateAzureResourceGroupBadRequest) WithPayload(payload *models.Error) *CreateAzureResourceGroupBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create azure resource group bad request response
func (o *CreateAzureResourceGroupBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAzureResourceGroupBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreateAzureResourceGroupUnauthorizedCode is the HTTP code returned for type CreateAzureResourceGroupUnauthorized
const CreateAzureResourceGroupUnauthorizedCode int = 401

/*CreateAzureResourceGroupUnauthorized Incorrect credentials

swagger:response createAzureResourceGroupUnauthorized
*/
type CreateAzureResourceGroupUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateAzureResourceGroupUnauthorized creates CreateAzureResourceGroupUnauthorized with default headers values
func NewCreateAzureResourceGroupUnauthorized() *CreateAzureResourceGroupUnauthorized {

	return &CreateAzureResourceGroupUnauthorized{}
}

// WithPayload adds the payload to the create azure resource group unauthorized response
func (o *CreateAzureResourceGroupUnauthorized) WithPayload(payload *models.Error) *CreateAzureResourceGroupUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create azure resource group unauthorized response
func (o *CreateAzureResourceGroupUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAzureResourceGroupUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// CreateAzureResourceGroupInternalServerErrorCode is the HTTP code returned for type CreateAzureResourceGroupInternalServerError
const CreateAzureResourceGroupInternalServerErrorCode int = 500

/*CreateAzureResourceGroupInternalServerError Internal server error

swagger:response createAzureResourceGroupInternalServerError
*/
type CreateAzureResourceGroupInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewCreateAzureResourceGroupInternalServerError creates CreateAzureResourceGroupInternalServerError with default headers values
func NewCreateAzureResourceGroupInternalServerError() *CreateAzureResourceGroupInternalServerError {

	return &CreateAzureResourceGroupInternalServerError{}
}

// WithPayload adds the payload to the create azure resource group internal server error response
func (o *CreateAzureResourceGroupInternalServerError) WithPayload(payload *models.Error) *CreateAzureResourceGroupInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the create azure resource group internal server error response
func (o *CreateAzureResourceGroupInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *CreateAzureResourceGroupInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
