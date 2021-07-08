// Code generated by go-swagger; DO NOT EDIT.

package avi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"

	models "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// NewVerifyAccountParams creates a new VerifyAccountParams object
// no default values defined in spec.
func NewVerifyAccountParams() VerifyAccountParams {

	return VerifyAccountParams{}
}

// VerifyAccountParams contains all the bound params for the verify account operation
// typically these are obtained from a http.Request
//
// swagger:parameters verifyAccount
type VerifyAccountParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Avi controller credentials
	  In: body
	*/
	Credentials *models.AviControllerParams
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewVerifyAccountParams() beforehand.
func (o *VerifyAccountParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	if runtime.HasBody(r) {
		defer r.Body.Close()
		var body models.AviControllerParams
		if err := route.Consumer.Consume(r.Body, &body); err != nil {
			res = append(res, errors.NewParseError("credentials", "body", "", err))
		} else {
			// validate body object
			if err := body.Validate(route.Formats); err != nil {
				res = append(res, err)
			}

			if len(res) == 0 {
				o.Credentials = &body
			}
		}
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
