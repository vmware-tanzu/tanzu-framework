// Code generated by go-swagger; DO NOT EDIT.

package tmc

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// RetrieveTMCInstallYmlHandlerFunc turns a function with the right signature into a retrieve t m c install yml handler
type RetrieveTMCInstallYmlHandlerFunc func(RetrieveTMCInstallYmlParams) middleware.Responder

// Handle executing the request and returning a response
func (fn RetrieveTMCInstallYmlHandlerFunc) Handle(params RetrieveTMCInstallYmlParams) middleware.Responder {
	return fn(params)
}

// RetrieveTMCInstallYmlHandler interface for that can handle valid retrieve t m c install yml params
type RetrieveTMCInstallYmlHandler interface {
	Handle(RetrieveTMCInstallYmlParams) middleware.Responder
}

// NewRetrieveTMCInstallYml creates a new http.Handler for the retrieve t m c install yml operation
func NewRetrieveTMCInstallYml(ctx *middleware.Context, handler RetrieveTMCInstallYmlHandler) *RetrieveTMCInstallYml {
	return &RetrieveTMCInstallYml{Context: ctx, Handler: handler}
}

/*RetrieveTMCInstallYml swagger:route GET /api/integration/tmc tmc retrieveTMCInstallYml

Retrieves TMC install yml from provided URL

*/
type RetrieveTMCInstallYml struct {
	Context *middleware.Context
	Handler RetrieveTMCInstallYmlHandler
}

func (o *RetrieveTMCInstallYml) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewRetrieveTMCInstallYmlParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
