// Code generated by go-swagger; DO NOT EDIT.

package avi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// GetAviCloudsReader is a Reader for the GetAviClouds structure.
type GetAviCloudsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAviCloudsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAviCloudsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetAviCloudsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetAviCloudsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAviCloudsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetAviCloudsOK creates a GetAviCloudsOK with default headers values
func NewGetAviCloudsOK() *GetAviCloudsOK {
	return &GetAviCloudsOK{}
}

/*GetAviCloudsOK handles this case with default header values.

Successful retrieval of Avi load balancer clouds
*/
type GetAviCloudsOK struct {
	Payload []*models.AviCloud
}

func (o *GetAviCloudsOK) Error() string {
	return fmt.Sprintf("[GET /api/avi/clouds][%d] getAviCloudsOK  %+v", 200, o.Payload)
}

func (o *GetAviCloudsOK) GetPayload() []*models.AviCloud {
	return o.Payload
}

func (o *GetAviCloudsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAviCloudsBadRequest creates a GetAviCloudsBadRequest with default headers values
func NewGetAviCloudsBadRequest() *GetAviCloudsBadRequest {
	return &GetAviCloudsBadRequest{}
}

/*GetAviCloudsBadRequest handles this case with default header values.

Bad request
*/
type GetAviCloudsBadRequest struct {
	Payload *models.Error
}

func (o *GetAviCloudsBadRequest) Error() string {
	return fmt.Sprintf("[GET /api/avi/clouds][%d] getAviCloudsBadRequest  %+v", 400, o.Payload)
}

func (o *GetAviCloudsBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAviCloudsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAviCloudsUnauthorized creates a GetAviCloudsUnauthorized with default headers values
func NewGetAviCloudsUnauthorized() *GetAviCloudsUnauthorized {
	return &GetAviCloudsUnauthorized{}
}

/*GetAviCloudsUnauthorized handles this case with default header values.

Incorrect credentials
*/
type GetAviCloudsUnauthorized struct {
	Payload *models.Error
}

func (o *GetAviCloudsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/avi/clouds][%d] getAviCloudsUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAviCloudsUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAviCloudsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAviCloudsInternalServerError creates a GetAviCloudsInternalServerError with default headers values
func NewGetAviCloudsInternalServerError() *GetAviCloudsInternalServerError {
	return &GetAviCloudsInternalServerError{}
}

/*GetAviCloudsInternalServerError handles this case with default header values.

Internal server error
*/
type GetAviCloudsInternalServerError struct {
	Payload *models.Error
}

func (o *GetAviCloudsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /api/avi/clouds][%d] getAviCloudsInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAviCloudsInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAviCloudsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
