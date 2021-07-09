// Code generated by go-swagger; DO NOT EDIT.

package aws

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// GetAWSOSImagesReader is a Reader for the GetAWSOSImages structure.
type GetAWSOSImagesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAWSOSImagesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAWSOSImagesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetAWSOSImagesBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetAWSOSImagesUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAWSOSImagesInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetAWSOSImagesOK creates a GetAWSOSImagesOK with default headers values
func NewGetAWSOSImagesOK() *GetAWSOSImagesOK {
	return &GetAWSOSImagesOK{}
}

/*GetAWSOSImagesOK handles this case with default header values.

Successful retrieval of AWS supported os images
*/
type GetAWSOSImagesOK struct {
	Payload []*models.AWSVirtualMachine
}

func (o *GetAWSOSImagesOK) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/osimages][%d] getAWSOSImagesOK  %+v", 200, o.Payload)
}

func (o *GetAWSOSImagesOK) GetPayload() []*models.AWSVirtualMachine {
	return o.Payload
}

func (o *GetAWSOSImagesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSOSImagesBadRequest creates a GetAWSOSImagesBadRequest with default headers values
func NewGetAWSOSImagesBadRequest() *GetAWSOSImagesBadRequest {
	return &GetAWSOSImagesBadRequest{}
}

/*GetAWSOSImagesBadRequest handles this case with default header values.

Bad request
*/
type GetAWSOSImagesBadRequest struct {
	Payload *models.Error
}

func (o *GetAWSOSImagesBadRequest) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/osimages][%d] getAWSOSImagesBadRequest  %+v", 400, o.Payload)
}

func (o *GetAWSOSImagesBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSOSImagesBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSOSImagesUnauthorized creates a GetAWSOSImagesUnauthorized with default headers values
func NewGetAWSOSImagesUnauthorized() *GetAWSOSImagesUnauthorized {
	return &GetAWSOSImagesUnauthorized{}
}

/*GetAWSOSImagesUnauthorized handles this case with default header values.

Incorrect credentials
*/
type GetAWSOSImagesUnauthorized struct {
	Payload *models.Error
}

func (o *GetAWSOSImagesUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/osimages][%d] getAWSOSImagesUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAWSOSImagesUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSOSImagesUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSOSImagesInternalServerError creates a GetAWSOSImagesInternalServerError with default headers values
func NewGetAWSOSImagesInternalServerError() *GetAWSOSImagesInternalServerError {
	return &GetAWSOSImagesInternalServerError{}
}

/*GetAWSOSImagesInternalServerError handles this case with default header values.

Internal server error
*/
type GetAWSOSImagesInternalServerError struct {
	Payload *models.Error
}

func (o *GetAWSOSImagesInternalServerError) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/osimages][%d] getAWSOSImagesInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAWSOSImagesInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSOSImagesInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
