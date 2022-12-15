// Code generated by go-swagger; DO NOT EDIT.

package aws

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// GetAWSKeyPairsReader is a Reader for the GetAWSKeyPairs structure.
type GetAWSKeyPairsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAWSKeyPairsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAWSKeyPairsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetAWSKeyPairsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetAWSKeyPairsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAWSKeyPairsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetAWSKeyPairsOK creates a GetAWSKeyPairsOK with default headers values
func NewGetAWSKeyPairsOK() *GetAWSKeyPairsOK {
	return &GetAWSKeyPairsOK{}
}

/*
GetAWSKeyPairsOK handles this case with default header values.

Successful retrieval of AWS key pairs
*/
type GetAWSKeyPairsOK struct {
	Payload []*models.AWSKeyPair
}

func (o *GetAWSKeyPairsOK) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/keypair][%d] getAWSKeyPairsOK  %+v", 200, o.Payload)
}

func (o *GetAWSKeyPairsOK) GetPayload() []*models.AWSKeyPair {
	return o.Payload
}

func (o *GetAWSKeyPairsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSKeyPairsBadRequest creates a GetAWSKeyPairsBadRequest with default headers values
func NewGetAWSKeyPairsBadRequest() *GetAWSKeyPairsBadRequest {
	return &GetAWSKeyPairsBadRequest{}
}

/*
GetAWSKeyPairsBadRequest handles this case with default header values.

Bad request
*/
type GetAWSKeyPairsBadRequest struct {
	Payload *models.Error
}

func (o *GetAWSKeyPairsBadRequest) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/keypair][%d] getAWSKeyPairsBadRequest  %+v", 400, o.Payload)
}

func (o *GetAWSKeyPairsBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSKeyPairsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSKeyPairsUnauthorized creates a GetAWSKeyPairsUnauthorized with default headers values
func NewGetAWSKeyPairsUnauthorized() *GetAWSKeyPairsUnauthorized {
	return &GetAWSKeyPairsUnauthorized{}
}

/*
GetAWSKeyPairsUnauthorized handles this case with default header values.

Incorrect credentials
*/
type GetAWSKeyPairsUnauthorized struct {
	Payload *models.Error
}

func (o *GetAWSKeyPairsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/keypair][%d] getAWSKeyPairsUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAWSKeyPairsUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSKeyPairsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAWSKeyPairsInternalServerError creates a GetAWSKeyPairsInternalServerError with default headers values
func NewGetAWSKeyPairsInternalServerError() *GetAWSKeyPairsInternalServerError {
	return &GetAWSKeyPairsInternalServerError{}
}

/*
GetAWSKeyPairsInternalServerError handles this case with default header values.

Internal server error
*/
type GetAWSKeyPairsInternalServerError struct {
	Payload *models.Error
}

func (o *GetAWSKeyPairsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /api/providers/aws/keypair][%d] getAWSKeyPairsInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAWSKeyPairsInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAWSKeyPairsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
