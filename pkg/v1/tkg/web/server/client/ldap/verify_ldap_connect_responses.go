// Code generated by go-swagger; DO NOT EDIT.

package ldap

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// VerifyLdapConnectReader is a Reader for the VerifyLdapConnect structure.
type VerifyLdapConnectReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *VerifyLdapConnectReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewVerifyLdapConnectOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewVerifyLdapConnectBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewVerifyLdapConnectUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewVerifyLdapConnectInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewVerifyLdapConnectOK creates a VerifyLdapConnectOK with default headers values
func NewVerifyLdapConnectOK() *VerifyLdapConnectOK {
	return &VerifyLdapConnectOK{}
}

/*VerifyLdapConnectOK handles this case with default header values.

Verified LDAP credentials successfully
*/
type VerifyLdapConnectOK struct {
	Payload *models.LdapTestResult
}

func (o *VerifyLdapConnectOK) Error() string {
	return fmt.Sprintf("[POST /api/ldap/connect][%d] verifyLdapConnectOK  %+v", 200, o.Payload)
}

func (o *VerifyLdapConnectOK) GetPayload() *models.LdapTestResult {
	return o.Payload
}

func (o *VerifyLdapConnectOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.LdapTestResult)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapConnectBadRequest creates a VerifyLdapConnectBadRequest with default headers values
func NewVerifyLdapConnectBadRequest() *VerifyLdapConnectBadRequest {
	return &VerifyLdapConnectBadRequest{}
}

/*VerifyLdapConnectBadRequest handles this case with default header values.

Bad request
*/
type VerifyLdapConnectBadRequest struct {
	Payload *models.Error
}

func (o *VerifyLdapConnectBadRequest) Error() string {
	return fmt.Sprintf("[POST /api/ldap/connect][%d] verifyLdapConnectBadRequest  %+v", 400, o.Payload)
}

func (o *VerifyLdapConnectBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapConnectBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapConnectUnauthorized creates a VerifyLdapConnectUnauthorized with default headers values
func NewVerifyLdapConnectUnauthorized() *VerifyLdapConnectUnauthorized {
	return &VerifyLdapConnectUnauthorized{}
}

/*VerifyLdapConnectUnauthorized handles this case with default header values.

Incorrect credentials
*/
type VerifyLdapConnectUnauthorized struct {
	Payload *models.Error
}

func (o *VerifyLdapConnectUnauthorized) Error() string {
	return fmt.Sprintf("[POST /api/ldap/connect][%d] verifyLdapConnectUnauthorized  %+v", 401, o.Payload)
}

func (o *VerifyLdapConnectUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapConnectUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewVerifyLdapConnectInternalServerError creates a VerifyLdapConnectInternalServerError with default headers values
func NewVerifyLdapConnectInternalServerError() *VerifyLdapConnectInternalServerError {
	return &VerifyLdapConnectInternalServerError{}
}

/*VerifyLdapConnectInternalServerError handles this case with default header values.

Internal server error
*/
type VerifyLdapConnectInternalServerError struct {
	Payload *models.Error
}

func (o *VerifyLdapConnectInternalServerError) Error() string {
	return fmt.Sprintf("[POST /api/ldap/connect][%d] verifyLdapConnectInternalServerError  %+v", 500, o.Payload)
}

func (o *VerifyLdapConnectInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *VerifyLdapConnectInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
