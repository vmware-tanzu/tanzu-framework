// Code generated by go-swagger; DO NOT EDIT.

package azure

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// GenerateTKGConfigForAzureReader is a Reader for the GenerateTKGConfigForAzure structure.
type GenerateTKGConfigForAzureReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GenerateTKGConfigForAzureReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGenerateTKGConfigForAzureOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGenerateTKGConfigForAzureBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGenerateTKGConfigForAzureUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGenerateTKGConfigForAzureInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGenerateTKGConfigForAzureOK creates a GenerateTKGConfigForAzureOK with default headers values
func NewGenerateTKGConfigForAzureOK() *GenerateTKGConfigForAzureOK {
	return &GenerateTKGConfigForAzureOK{}
}

/*GenerateTKGConfigForAzureOK handles this case with default header values.

Generated TKG configuration successfully
*/
type GenerateTKGConfigForAzureOK struct {
	Payload string
}

func (o *GenerateTKGConfigForAzureOK) Error() string {
	return fmt.Sprintf("[POST /api/providers/azure/generate][%d] generateTKGConfigForAzureOK  %+v", 200, o.Payload)
}

func (o *GenerateTKGConfigForAzureOK) GetPayload() string {
	return o.Payload
}

func (o *GenerateTKGConfigForAzureOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGenerateTKGConfigForAzureBadRequest creates a GenerateTKGConfigForAzureBadRequest with default headers values
func NewGenerateTKGConfigForAzureBadRequest() *GenerateTKGConfigForAzureBadRequest {
	return &GenerateTKGConfigForAzureBadRequest{}
}

/*GenerateTKGConfigForAzureBadRequest handles this case with default header values.

Bad request
*/
type GenerateTKGConfigForAzureBadRequest struct {
	Payload *models.Error
}

func (o *GenerateTKGConfigForAzureBadRequest) Error() string {
	return fmt.Sprintf("[POST /api/providers/azure/generate][%d] generateTKGConfigForAzureBadRequest  %+v", 400, o.Payload)
}

func (o *GenerateTKGConfigForAzureBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GenerateTKGConfigForAzureBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGenerateTKGConfigForAzureUnauthorized creates a GenerateTKGConfigForAzureUnauthorized with default headers values
func NewGenerateTKGConfigForAzureUnauthorized() *GenerateTKGConfigForAzureUnauthorized {
	return &GenerateTKGConfigForAzureUnauthorized{}
}

/*GenerateTKGConfigForAzureUnauthorized handles this case with default header values.

Incorrect credentials
*/
type GenerateTKGConfigForAzureUnauthorized struct {
	Payload *models.Error
}

func (o *GenerateTKGConfigForAzureUnauthorized) Error() string {
	return fmt.Sprintf("[POST /api/providers/azure/generate][%d] generateTKGConfigForAzureUnauthorized  %+v", 401, o.Payload)
}

func (o *GenerateTKGConfigForAzureUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GenerateTKGConfigForAzureUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGenerateTKGConfigForAzureInternalServerError creates a GenerateTKGConfigForAzureInternalServerError with default headers values
func NewGenerateTKGConfigForAzureInternalServerError() *GenerateTKGConfigForAzureInternalServerError {
	return &GenerateTKGConfigForAzureInternalServerError{}
}

/*GenerateTKGConfigForAzureInternalServerError handles this case with default header values.

Internal server error
*/
type GenerateTKGConfigForAzureInternalServerError struct {
	Payload *models.Error
}

func (o *GenerateTKGConfigForAzureInternalServerError) Error() string {
	return fmt.Sprintf("[POST /api/providers/azure/generate][%d] generateTKGConfigForAzureInternalServerError  %+v", 500, o.Payload)
}

func (o *GenerateTKGConfigForAzureInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GenerateTKGConfigForAzureInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
