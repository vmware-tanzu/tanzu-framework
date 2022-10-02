// Code generated by go-swagger; DO NOT EDIT.

package aws

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// ExportTKGConfigForAWSHandlerFunc turns a function with the right signature into a export t k g config for a w s handler
type ExportTKGConfigForAWSHandlerFunc func(ExportTKGConfigForAWSParams) middleware.Responder

// Handle executing the request and returning a response
func (fn ExportTKGConfigForAWSHandlerFunc) Handle(params ExportTKGConfigForAWSParams) middleware.Responder {
	return fn(params)
}

// ExportTKGConfigForAWSHandler interface for that can handle valid export t k g config for a w s params
type ExportTKGConfigForAWSHandler interface {
	Handle(ExportTKGConfigForAWSParams) middleware.Responder
}

// NewExportTKGConfigForAWS creates a new http.Handler for the export t k g config for a w s operation
func NewExportTKGConfigForAWS(ctx *middleware.Context, handler ExportTKGConfigForAWSHandler) *ExportTKGConfigForAWS {
	return &ExportTKGConfigForAWS{Context: ctx, Handler: handler}
}

/*
ExportTKGConfigForAWS swagger:route POST /api/providers/aws/config/export aws exportTKGConfigForAWS

Generate TKG configuration file for AWS"
*/
type ExportTKGConfigForAWS struct {
	Context *middleware.Context
	Handler ExportTKGConfigForAWSHandler
}

func (o *ExportTKGConfigForAWS) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewExportTKGConfigForAWSParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
