// Code generated by go-swagger; DO NOT EDIT.

package ldap

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// VerifyLdapUserSearchOKCode is the HTTP code returned for type VerifyLdapUserSearchOK
const VerifyLdapUserSearchOKCode int = 200

/*VerifyLdapUserSearchOK Verified LDAP credentials successfully

swagger:response verifyLdapUserSearchOK
*/
type VerifyLdapUserSearchOK struct {

	/*
	  In: Body
	*/
	Payload *models.LdapTestResult `json:"body,omitempty"`
}

// NewVerifyLdapUserSearchOK creates VerifyLdapUserSearchOK with default headers values
func NewVerifyLdapUserSearchOK() *VerifyLdapUserSearchOK {

	return &VerifyLdapUserSearchOK{}
}

// WithPayload adds the payload to the verify ldap user search o k response
func (o *VerifyLdapUserSearchOK) WithPayload(payload *models.LdapTestResult) *VerifyLdapUserSearchOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap user search o k response
func (o *VerifyLdapUserSearchOK) SetPayload(payload *models.LdapTestResult) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapUserSearchOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// VerifyLdapUserSearchBadRequestCode is the HTTP code returned for type VerifyLdapUserSearchBadRequest
const VerifyLdapUserSearchBadRequestCode int = 400

/*VerifyLdapUserSearchBadRequest Bad request

swagger:response verifyLdapUserSearchBadRequest
*/
type VerifyLdapUserSearchBadRequest struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewVerifyLdapUserSearchBadRequest creates VerifyLdapUserSearchBadRequest with default headers values
func NewVerifyLdapUserSearchBadRequest() *VerifyLdapUserSearchBadRequest {

	return &VerifyLdapUserSearchBadRequest{}
}

// WithPayload adds the payload to the verify ldap user search bad request response
func (o *VerifyLdapUserSearchBadRequest) WithPayload(payload *models.Error) *VerifyLdapUserSearchBadRequest {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap user search bad request response
func (o *VerifyLdapUserSearchBadRequest) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapUserSearchBadRequest) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(400)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// VerifyLdapUserSearchUnauthorizedCode is the HTTP code returned for type VerifyLdapUserSearchUnauthorized
const VerifyLdapUserSearchUnauthorizedCode int = 401

/*VerifyLdapUserSearchUnauthorized Incorrect credentials

swagger:response verifyLdapUserSearchUnauthorized
*/
type VerifyLdapUserSearchUnauthorized struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewVerifyLdapUserSearchUnauthorized creates VerifyLdapUserSearchUnauthorized with default headers values
func NewVerifyLdapUserSearchUnauthorized() *VerifyLdapUserSearchUnauthorized {

	return &VerifyLdapUserSearchUnauthorized{}
}

// WithPayload adds the payload to the verify ldap user search unauthorized response
func (o *VerifyLdapUserSearchUnauthorized) WithPayload(payload *models.Error) *VerifyLdapUserSearchUnauthorized {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap user search unauthorized response
func (o *VerifyLdapUserSearchUnauthorized) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapUserSearchUnauthorized) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(401)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

// VerifyLdapUserSearchInternalServerErrorCode is the HTTP code returned for type VerifyLdapUserSearchInternalServerError
const VerifyLdapUserSearchInternalServerErrorCode int = 500

/*VerifyLdapUserSearchInternalServerError Internal server error

swagger:response verifyLdapUserSearchInternalServerError
*/
type VerifyLdapUserSearchInternalServerError struct {

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewVerifyLdapUserSearchInternalServerError creates VerifyLdapUserSearchInternalServerError with default headers values
func NewVerifyLdapUserSearchInternalServerError() *VerifyLdapUserSearchInternalServerError {

	return &VerifyLdapUserSearchInternalServerError{}
}

// WithPayload adds the payload to the verify ldap user search internal server error response
func (o *VerifyLdapUserSearchInternalServerError) WithPayload(payload *models.Error) *VerifyLdapUserSearchInternalServerError {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the verify ldap user search internal server error response
func (o *VerifyLdapUserSearchInternalServerError) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *VerifyLdapUserSearchInternalServerError) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(500)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
