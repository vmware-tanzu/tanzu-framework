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

// VerifyLdapGroupSearchReader is a Reader for the VerifyLdapGroupSearch structure.
type VerifyLdapGroupSearchReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *VerifyLdapGroupSearchReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewVerifyLdapGroupSearchOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewVerifyLdapGroupSearchBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewVerifyLdapGroupSearchUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewVerifyLdapGroupSearchInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewVerifyLdapGroupSearchOK creates a VerifyLdapGroupSearchOK with default headers values
func NewVerifyLdapGroupSearchOK() *VerifyLdapGroupSearchOK {
	return &VerifyLdapGroupSearchOK{}
}

/*
VerifyLdapGroupSearchOK handles this case with default header values.

Verified LDAP credentials successfully
*/
type VerifyLdapGroupSearchOK struct {
	Payload *models.LdapTestResult
}

func (o *VerifyLdapGroupSearchOK) Error() string {
	return fmt.Sprintf("[POST /api/ldap/groups/search][%d] verifyLdapGroupSearchOK  %+v", 200, o.Payload)
}

func (o *VerifyLdapGroupSearchOK) GetPayload() *models.LdapTestResult {
	return o.Payload
}

func (o *VerifyLdapGroupSearchOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.LdapTestResult)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapGroupSearchBadRequest creates a VerifyLdapGroupSearchBadRequest with default headers values
func NewVerifyLdapGroupSearchBadRequest() *VerifyLdapGroupSearchBadRequest {
	return &VerifyLdapGroupSearchBadRequest{}
}

/*
VerifyLdapGroupSearchBadRequest handles this case with default header values.

Bad request
*/
type VerifyLdapGroupSearchBadRequest struct {
	Payload *models.Error
}

func (o *VerifyLdapGroupSearchBadRequest) Error() string {
	return fmt.Sprintf("[POST /api/ldap/groups/search][%d] verifyLdapGroupSearchBadRequest  %+v", 400, o.Payload)
}

func (o *VerifyLdapGroupSearchBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapGroupSearchBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapGroupSearchUnauthorized creates a VerifyLdapGroupSearchUnauthorized with default headers values
func NewVerifyLdapGroupSearchUnauthorized() *VerifyLdapGroupSearchUnauthorized {
	return &VerifyLdapGroupSearchUnauthorized{}
}

/*
VerifyLdapGroupSearchUnauthorized handles this case with default header values.

Incorrect credentials
*/
type VerifyLdapGroupSearchUnauthorized struct {
	Payload *models.Error
}

func (o *VerifyLdapGroupSearchUnauthorized) Error() string {
	return fmt.Sprintf("[POST /api/ldap/groups/search][%d] verifyLdapGroupSearchUnauthorized  %+v", 401, o.Payload)
}

func (o *VerifyLdapGroupSearchUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapGroupSearchUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapGroupSearchInternalServerError creates a VerifyLdapGroupSearchInternalServerError with default headers values
func NewVerifyLdapGroupSearchInternalServerError() *VerifyLdapGroupSearchInternalServerError {
	return &VerifyLdapGroupSearchInternalServerError{}
}

/*
VerifyLdapGroupSearchInternalServerError handles this case with default header values.

Internal server error
*/
type VerifyLdapGroupSearchInternalServerError struct {
	Payload *models.Error
}

func (o *VerifyLdapGroupSearchInternalServerError) Error() string {
	return fmt.Sprintf("[POST /api/ldap/groups/search][%d] verifyLdapGroupSearchInternalServerError  %+v", 500, o.Payload)
}

func (o *VerifyLdapGroupSearchInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapGroupSearchInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
