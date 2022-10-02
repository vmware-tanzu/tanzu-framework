// Code generated by go-swagger; DO NOT EDIT.

package ldap

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// VerifyLdapBindReader is a Reader for the VerifyLdapBind structure.
type VerifyLdapBindReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *VerifyLdapBindReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewVerifyLdapBindOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewVerifyLdapBindBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewVerifyLdapBindUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewVerifyLdapBindInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewVerifyLdapBindOK creates a VerifyLdapBindOK with default headers values
func NewVerifyLdapBindOK() *VerifyLdapBindOK {
	return &VerifyLdapBindOK{}
}

/*
VerifyLdapBindOK handles this case with default header values.

Verified LDAP credentials successfully
*/
type VerifyLdapBindOK struct {
	Payload *models.LdapTestResult
}

func (o *VerifyLdapBindOK) Error() string {
	return fmt.Sprintf("[POST /api/ldap/bind][%d] verifyLdapBindOK  %+v", 200, o.Payload)
}

func (o *VerifyLdapBindOK) GetPayload() *models.LdapTestResult {
	return o.Payload
}

func (o *VerifyLdapBindOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.LdapTestResult)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapBindBadRequest creates a VerifyLdapBindBadRequest with default headers values
func NewVerifyLdapBindBadRequest() *VerifyLdapBindBadRequest {
	return &VerifyLdapBindBadRequest{}
}

/*
VerifyLdapBindBadRequest handles this case with default header values.

Bad request
*/
type VerifyLdapBindBadRequest struct {
	Payload *models.Error
}

func (o *VerifyLdapBindBadRequest) Error() string {
	return fmt.Sprintf("[POST /api/ldap/bind][%d] verifyLdapBindBadRequest  %+v", 400, o.Payload)
}

func (o *VerifyLdapBindBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapBindBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapBindUnauthorized creates a VerifyLdapBindUnauthorized with default headers values
func NewVerifyLdapBindUnauthorized() *VerifyLdapBindUnauthorized {
	return &VerifyLdapBindUnauthorized{}
}

/*
VerifyLdapBindUnauthorized handles this case with default header values.

Incorrect credentials
*/
type VerifyLdapBindUnauthorized struct {
	Payload *models.Error
}

func (o *VerifyLdapBindUnauthorized) Error() string {
	return fmt.Sprintf("[POST /api/ldap/bind][%d] verifyLdapBindUnauthorized  %+v", 401, o.Payload)
}

func (o *VerifyLdapBindUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapBindUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapBindInternalServerError creates a VerifyLdapBindInternalServerError with default headers values
func NewVerifyLdapBindInternalServerError() *VerifyLdapBindInternalServerError {
	return &VerifyLdapBindInternalServerError{}
}

/*
VerifyLdapBindInternalServerError handles this case with default header values.

Internal server error
*/
type VerifyLdapBindInternalServerError struct {
	Payload *models.Error
}

func (o *VerifyLdapBindInternalServerError) Error() string {
	return fmt.Sprintf("[POST /api/ldap/bind][%d] verifyLdapBindInternalServerError  %+v", 500, o.Payload)
}

func (o *VerifyLdapBindInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapBindInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
