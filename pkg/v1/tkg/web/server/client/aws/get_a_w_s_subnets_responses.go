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

// GetAWSSubnetsReader is a Reader for the GetAWSSubnets structure.
type GetAWSSubnetsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAWSSubnetsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAWSSubnetsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetAWSSubnetsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetAWSSubnetsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAWSSubnetsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetAWSSubnetsOK creates a GetAWSSubnetsOK with default headers values
func NewGetAWSSubnetsOK() *GetAWSSubnetsOK {
	return &GetAWSSubnetsOK{}
}

/*GetAWSSubnetsOK handles this case with default header values.

Successful retrieval of AWS subnets
*/
type GetAWSSubnetsOK struct {
	Payload []*models.AWSSubnet
}

func (o *GetAWSSubnetsOK) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/subnets][%d] getAWSSubnetsOK  %+v", 200, o.Payload)
}

func (o *GetAWSSubnetsOK) GetPayload() []*models.AWSSubnet {
	return o.Payload
}

func (o *GetAWSSubnetsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSSubnetsBadRequest creates a GetAWSSubnetsBadRequest with default headers values
func NewGetAWSSubnetsBadRequest() *GetAWSSubnetsBadRequest {
	return &GetAWSSubnetsBadRequest{}
}

/*GetAWSSubnetsBadRequest handles this case with default header values.

Bad request
*/
type GetAWSSubnetsBadRequest struct {
	Payload *models.Error
}

func (o *GetAWSSubnetsBadRequest) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/subnets][%d] getAWSSubnetsBadRequest  %+v", 400, o.Payload)
}

func (o *GetAWSSubnetsBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSSubnetsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSSubnetsUnauthorized creates a GetAWSSubnetsUnauthorized with default headers values
func NewGetAWSSubnetsUnauthorized() *GetAWSSubnetsUnauthorized {
	return &GetAWSSubnetsUnauthorized{}
}

/*GetAWSSubnetsUnauthorized handles this case with default header values.

Incorrect credentials
*/
type GetAWSSubnetsUnauthorized struct {
	Payload *models.Error
}

func (o *GetAWSSubnetsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/subnets][%d] getAWSSubnetsUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAWSSubnetsUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSSubnetsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSSubnetsInternalServerError creates a GetAWSSubnetsInternalServerError with default headers values
func NewGetAWSSubnetsInternalServerError() *GetAWSSubnetsInternalServerError {
	return &GetAWSSubnetsInternalServerError{}
}

/*GetAWSSubnetsInternalServerError handles this case with default header values.

Internal server error
*/
type GetAWSSubnetsInternalServerError struct {
	Payload *models.Error
}

func (o *GetAWSSubnetsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/subnets][%d] getAWSSubnetsInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAWSSubnetsInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSSubnetsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
