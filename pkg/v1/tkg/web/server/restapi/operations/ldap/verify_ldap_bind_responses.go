// Code generated by go-swagger; DO NOT EDIT.

package ldap

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// VerifyLdapBindOKCode is the HTTP code returned for type VerifyLdapBindOK
const VerifyLdapBindOKCode int = 200

/*VerifyLdapBindOK Verified LDAP credentials successfully

swagger:response verifyLdapBindOK
*/
type VerifyLdapBindOK struct {

	/*
	  In: Body
	*/
	Payload *models.LdapTestResult `json:"body,omitempty"`
}

// NewVerifyLdapBindOK creates VerifyLdapBindOK with default headers values
func NewVerifyLdapBindOK() *VerifyLdapBindOK {

	return &VerifyLdapBindOK{}
}

// WithPayload adds the payload to the verify ldap bind o k response
func (o *VerifyLdapBindOK) WithPayload(payload *models.LdapTestResult) *VerifyLdapBindOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap bind o k response
func (o *VerifyLdapBindOK) SetPayload(payload *models.LdapTestResult) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapBindOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// VerifyLdapBindBadRequestCode is the HTTP code returned for type VerifyLdapBindBadRequest
const VerifyLdapBindBadRequestCode int = 400

/*VerifyLdapBindBadRequest Bad request

swagger:response verifyLdapBindBadRequest
*/
type VerifyLdapBindBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewVerifyLdapBindBadRequest creates VerifyLdapBindBadRequest with default headers values
func NewVerifyLdapBindBadRequest() *VerifyLdapBindBadRequest {

	return &VerifyLdapBindBadRequest{}
}

// WithPayload adds the payload to the verify ldap bind bad request response
func (o *VerifyLdapBindBadRequest) WithPayload(payload *models.Error) *VerifyLdapBindBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap bind bad request response
func (o *VerifyLdapBindBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapBindBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// VerifyLdapBindUnauthorizedCode is the HTTP code returned for type VerifyLdapBindUnauthorized
const VerifyLdapBindUnauthorizedCode int = 401

/*VerifyLdapBindUnauthorized Incorrect credentials

swagger:response verifyLdapBindUnauthorized
*/
type VerifyLdapBindUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewVerifyLdapBindUnauthorized creates VerifyLdapBindUnauthorized with default headers values
func NewVerifyLdapBindUnauthorized() *VerifyLdapBindUnauthorized {

	return &VerifyLdapBindUnauthorized{}
}

// WithPayload adds the payload to the verify ldap bind unauthorized response
func (o *VerifyLdapBindUnauthorized) WithPayload(payload *models.Error) *VerifyLdapBindUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap bind unauthorized response
func (o *VerifyLdapBindUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapBindUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// VerifyLdapBindInternalServerErrorCode is the HTTP code returned for type VerifyLdapBindInternalServerError
const VerifyLdapBindInternalServerErrorCode int = 500

/*VerifyLdapBindInternalServerError Internal server error

swagger:response verifyLdapBindInternalServerError
*/
type VerifyLdapBindInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewVerifyLdapBindInternalServerError creates VerifyLdapBindInternalServerError with default headers values
func NewVerifyLdapBindInternalServerError() *VerifyLdapBindInternalServerError {

	return &VerifyLdapBindInternalServerError{}
}

// WithPayload adds the payload to the verify ldap bind internal server error response
func (o *VerifyLdapBindInternalServerError) WithPayload(payload *models.Error) *VerifyLdapBindInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap bind internal server error response
func (o *VerifyLdapBindInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapBindInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
