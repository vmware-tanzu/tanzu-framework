// Code generated by go-swagger; DO NOT EDIT.

package features

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// GetFeatureFlagsReader is a Reader for the GetFeatureFlags structure.
type GetFeatureFlagsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetFeatureFlagsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetFeatureFlagsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetFeatureFlagsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetFeatureFlagsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetFeatureFlagsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetFeatureFlagsOK creates a GetFeatureFlagsOK with default headers values
func NewGetFeatureFlagsOK() *GetFeatureFlagsOK {
	return &GetFeatureFlagsOK{}
}

/*
GetFeatureFlagsOK handles this case with default header values.

Successful retrieval of feature flags
*/
type GetFeatureFlagsOK struct {
	Payload models.Features
}

func (o *GetFeatureFlagsOK) Error() string {
	return fmt.Sprintf("[GET /api/features][%d] getFeatureFlagsOK  %+v", 200, o.Payload)
}

func (o *GetFeatureFlagsOK) GetPayload() models.Features {
	return o.Payload
}

func (o *GetFeatureFlagsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFeatureFlagsBadRequest creates a GetFeatureFlagsBadRequest with default headers values
func NewGetFeatureFlagsBadRequest() *GetFeatureFlagsBadRequest {
	return &GetFeatureFlagsBadRequest{}
}

/*
GetFeatureFlagsBadRequest handles this case with default header values.

Bad Request
*/
type GetFeatureFlagsBadRequest struct {
	Payload *models.Error
}

func (o *GetFeatureFlagsBadRequest) Error() string {
	return fmt.Sprintf("[GET /api/features][%d] getFeatureFlagsBadRequest  %+v", 400, o.Payload)
}

func (o *GetFeatureFlagsBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetFeatureFlagsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFeatureFlagsUnauthorized creates a GetFeatureFlagsUnauthorized with default headers values
func NewGetFeatureFlagsUnauthorized() *GetFeatureFlagsUnauthorized {
	return &GetFeatureFlagsUnauthorized{}
}

/*
GetFeatureFlagsUnauthorized handles this case with default header values.

Incorrect credentials
*/
type GetFeatureFlagsUnauthorized struct {
	Payload *models.Error
}

func (o *GetFeatureFlagsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/features][%d] getFeatureFlagsUnauthorized  %+v", 401, o.Payload)
}

func (o *GetFeatureFlagsUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetFeatureFlagsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFeatureFlagsInternalServerError creates a GetFeatureFlagsInternalServerError with default headers values
func NewGetFeatureFlagsInternalServerError() *GetFeatureFlagsInternalServerError {
	return &GetFeatureFlagsInternalServerError{}
}

/*
GetFeatureFlagsInternalServerError handles this case with default header values.

Internal server error
*/
type GetFeatureFlagsInternalServerError struct {
	Payload *models.Error
}

func (o *GetFeatureFlagsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /api/features][%d] getFeatureFlagsInternalServerError  %+v", 500, o.Payload)
}

func (o *GetFeatureFlagsInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetFeatureFlagsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
