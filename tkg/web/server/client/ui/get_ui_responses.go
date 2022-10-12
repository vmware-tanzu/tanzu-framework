// Code generated by go-swagger; DO NOT EDIT.

package ui

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// GetUIReader is a Reader for the GetUI structure.
type GetUIReader struct {
	formats strfmt.Registry
	writer  io.Writer
}

// ReadResponse reads a server response into the received o.
func (o *GetUIReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetUIOK(o.writer)
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewGetUIBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewGetUIInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewGetUIOK creates a GetUIOK with default headers values
func NewGetUIOK(writer io.Writer) *GetUIOK {
	return &GetUIOK{
		Payload: writer,
	}
}

/*
GetUIOK handles this case with default header values.

Successful operation
*/
type GetUIOK struct {
	Payload io.Writer
}

func (o *GetUIOK) Error() string {
	return fmt.Sprintf("[GET /][%d] getUiOK  %+v", 200, o.Payload)
}

func (o *GetUIOK) GetPayload() io.Writer {
	return o.Payload
}

func (o *GetUIOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetUIBadRequest creates a GetUIBadRequest with default headers values
func NewGetUIBadRequest() *GetUIBadRequest {
	return &GetUIBadRequest{}
}

/*
GetUIBadRequest handles this case with default header values.

Bad request
*/
type GetUIBadRequest struct {
	Payload *models.Error
}

func (o *GetUIBadRequest) Error() string {
	return fmt.Sprintf("[GET /][%d] getUiBadRequest  %+v", 400, o.Payload)
}

func (o *GetUIBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetUIBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetUIInternalServerError creates a GetUIInternalServerError with default headers values
func NewGetUIInternalServerError() *GetUIInternalServerError {
	return &GetUIInternalServerError{}
}

/*
GetUIInternalServerError handles this case with default header values.

Internal server error
*/
type GetUIInternalServerError struct {
	Payload *models.Error
}

func (o *GetUIInternalServerError) Error() string {
	return fmt.Sprintf("[GET /][%d] getUiInternalServerError  %+v", 500, o.Payload)
}

func (o *GetUIInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *GetUIInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
