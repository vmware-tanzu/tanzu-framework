// Code generated by go-swagger; DO NOT EDIT.

package docker

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// ApplyTKGConfigForDockerReader is a Reader for the ApplyTKGConfigForDocker structure.
type ApplyTKGConfigForDockerReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ApplyTKGConfigForDockerReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewApplyTKGConfigForDockerOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 400:
		result := NewApplyTKGConfigForDockerBadRequest()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewApplyTKGConfigForDockerInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewApplyTKGConfigForDockerOK creates a ApplyTKGConfigForDockerOK with default headers values
func NewApplyTKGConfigForDockerOK() *ApplyTKGConfigForDockerOK {
	return &ApplyTKGConfigForDockerOK{}
}

/*ApplyTKGConfigForDockerOK handles this case with default header values.

Apply change to TKG configuration successfully
*/
type ApplyTKGConfigForDockerOK struct {
	Payload *models.ConfigFileInfo
}

func (o *ApplyTKGConfigForDockerOK) Error() string {
	return fmt.Sprintf("[POST /api/providers/docker/tkgconfig][%d] applyTKGConfigForDockerOK  %+v", 200, o.Payload)
}

func (o *ApplyTKGConfigForDockerOK) GetPayload() *models.ConfigFileInfo {
	return o.Payload
}

func (o *ApplyTKGConfigForDockerOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ConfigFileInfo)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewApplyTKGConfigForDockerBadRequest creates a ApplyTKGConfigForDockerBadRequest with default headers values
func NewApplyTKGConfigForDockerBadRequest() *ApplyTKGConfigForDockerBadRequest {
	return &ApplyTKGConfigForDockerBadRequest{}
}

/*ApplyTKGConfigForDockerBadRequest handles this case with default header values.

Bad request
*/
type ApplyTKGConfigForDockerBadRequest struct {
	Payload *models.Error
}

func (o *ApplyTKGConfigForDockerBadRequest) Error() string {
	return fmt.Sprintf("[POST /api/providers/docker/tkgconfig][%d] applyTKGConfigForDockerBadRequest  %+v", 400, o.Payload)
}

func (o *ApplyTKGConfigForDockerBadRequest) GetPayload() *models.Error {
	return o.Payload
}

func (o *ApplyTKGConfigForDockerBadRequest) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewApplyTKGConfigForDockerInternalServerError creates a ApplyTKGConfigForDockerInternalServerError with default headers values
func NewApplyTKGConfigForDockerInternalServerError() *ApplyTKGConfigForDockerInternalServerError {
	return &ApplyTKGConfigForDockerInternalServerError{}
}

/*ApplyTKGConfigForDockerInternalServerError handles this case with default header values.

Internal server error
*/
type ApplyTKGConfigForDockerInternalServerError struct {
	Payload *models.Error
}

func (o *ApplyTKGConfigForDockerInternalServerError) Error() string {
	return fmt.Sprintf("[POST /api/providers/docker/tkgconfig][%d] applyTKGConfigForDockerInternalServerError  %+v", 500, o.Payload)
}

func (o *ApplyTKGConfigForDockerInternalServerError) GetPayload() *models.Error {
	return o.Payload
}

func (o *ApplyTKGConfigForDockerInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
