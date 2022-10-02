// Code generated by go-swagger; DO NOT EDIT.

package docker

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/tkg/web/server/models"
)

// ExportTKGConfigForDockerReader is a Reader for the ExportTKGConfigForDocker structure.
type ExportTKGConfigForDockerReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ExportTKGConfigForDockerReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewExportTKGConfigForDockerOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewExportTKGConfigForDockerBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewExportTKGConfigForDockerInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewExportTKGConfigForDockerOK creates a ExportTKGConfigForDockerOK with default headers values
func NewExportTKGConfigForDockerOK() *ExportTKGConfigForDockerOK {
	return &ExportTKGConfigForDockerOK{}
}

/*
ExportTKGConfigForDockerOK handles this case with default header values.

Generated TKG configuration successfully
*/
type ExportTKGConfigForDockerOK struct {
	Payload string
}

func (o *ExportTKGConfigForDockerOK) Error() string {
	return fmt.Sprintf("[POST /api/providers/docker/config/export][%d] exportTKGConfigForDockerOK  %+v", 200, o.Payload)
}

func (o *ExportTKGConfigForDockerOK) GetPayload() string {
	return o.Payload
}

func (o *ExportTKGConfigForDockerOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportTKGConfigForDockerBadRequest creates a ExportTKGConfigForDockerBadRequest with default headers values
func NewExportTKGConfigForDockerBadRequest() *ExportTKGConfigForDockerBadRequest {
	return &ExportTKGConfigForDockerBadRequest{}
}

/*
ExportTKGConfigForDockerBadRequest handles this case with default header values.

Bad request
*/
type ExportTKGConfigForDockerBadRequest struct {
	Payload *models.Error
}

func (o *ExportTKGConfigForDockerBadRequest) Error() string {
	return fmt.Sprintf("[POST /api/providers/docker/config/export][%d] exportTKGConfigForDockerBadRequest  %+v", 400, o.Payload)
}

func (o *ExportTKGConfigForDockerBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportTKGConfigForDockerBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewExportTKGConfigForDockerInternalServerError creates a ExportTKGConfigForDockerInternalServerError with default headers values
func NewExportTKGConfigForDockerInternalServerError() *ExportTKGConfigForDockerInternalServerError {
	return &ExportTKGConfigForDockerInternalServerError{}
}

/*
ExportTKGConfigForDockerInternalServerError handles this case with default header values.

Internal server error
*/
type ExportTKGConfigForDockerInternalServerError struct {
	Payload *models.Error
}

func (o *ExportTKGConfigForDockerInternalServerError) Error() string {
	return fmt.Sprintf("[POST /api/providers/docker/config/export][%d] exportTKGConfigForDockerInternalServerError  %+v", 500, o.Payload)
}

func (o *ExportTKGConfigForDockerInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *ExportTKGConfigForDockerInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
