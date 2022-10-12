// Code generated by go-swagger; DO NOT EDIT.

package azure

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// GetAzureVnetsReader is a Reader for the GetAzureVnets structure.
type GetAzureVnetsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetAzureVnetsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetAzureVnetsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetAzureVnetsBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 401:
		result := NewGetAzureVnetsUnauthorized()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetAzureVnetsInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetAzureVnetsOK creates a GetAzureVnetsOK with default headers values
func NewGetAzureVnetsOK() *GetAzureVnetsOK {
	return &GetAzureVnetsOK{}
}

/*
GetAzureVnetsOK handles this case with default header values.

Successful retrieval of Azure virtual networks
*/
type GetAzureVnetsOK struct {
	Payload []*models.AzureVirtualNetwork
}

func (o *GetAzureVnetsOK) Error() string {
	return fmt.Sprintf("[GET /api/providers/azure/resourcegroups/{resourceGroupName}/vnets][%d] getAzureVnetsOK  %+v", 200, o.Payload)
}

func (o *GetAzureVnetsOK) GetPayload() []*models.AzureVirtualNetwork {
	return o.Payload
}

func (o *GetAzureVnetsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAzureVnetsBadRequest creates a GetAzureVnetsBadRequest with default headers values
func NewGetAzureVnetsBadRequest() *GetAzureVnetsBadRequest {
	return &GetAzureVnetsBadRequest{}
}

/*
GetAzureVnetsBadRequest handles this case with default header values.

Bad Request
*/
type GetAzureVnetsBadRequest struct {
	Payload *models.Error
}

func (o *GetAzureVnetsBadRequest) Error() string {
	return fmt.Sprintf("[GET /api/providers/azure/resourcegroups/{resourceGroupName}/vnets][%d] getAzureVnetsBadRequest  %+v", 400, o.Payload)
}

func (o *GetAzureVnetsBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAzureVnetsBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAzureVnetsUnauthorized creates a GetAzureVnetsUnauthorized with default headers values
func NewGetAzureVnetsUnauthorized() *GetAzureVnetsUnauthorized {
	return &GetAzureVnetsUnauthorized{}
}

/*
GetAzureVnetsUnauthorized handles this case with default header values.

Incorrect credentials
*/
type GetAzureVnetsUnauthorized struct {
	Payload *models.Error
}

func (o *GetAzureVnetsUnauthorized) Error() string {
	return fmt.Sprintf("[GET /api/providers/azure/resourcegroups/{resourceGroupName}/vnets][%d] getAzureVnetsUnauthorized  %+v", 401, o.Payload)
}

func (o *GetAzureVnetsUnauthorized) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAzureVnetsUnauthorized) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetAzureVnetsInternalServerError creates a GetAzureVnetsInternalServerError with default headers values
func NewGetAzureVnetsInternalServerError() *GetAzureVnetsInternalServerError {
	return &GetAzureVnetsInternalServerError{}
}

/*
GetAzureVnetsInternalServerError handles this case with default header values.

Internal server error
*/
type GetAzureVnetsInternalServerError struct {
	Payload *models.Error
}

func (o *GetAzureVnetsInternalServerError) Error() string {
	return fmt.Sprintf("[GET /api/providers/azure/resourcegroups/{resourceGroupName}/vnets][%d] getAzureVnetsInternalServerError  %+v", 500, o.Payload)
}

func (o *GetAzureVnetsInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetAzureVnetsInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
