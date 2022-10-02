// Code generated by go-swagger; DO NOT EDIT.

package avi

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetAviCloudsHandlerFunc turns a function with the right signature into a get avi clouds handler
type GetAviCloudsHandlerFunc func(GetAviCloudsParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAviCloudsHandlerFunc) Handle(params GetAviCloudsParams) middleware.Responder {
	return fn(params)
}

// GetAviCloudsHandler interface for that can handle valid get avi clouds params
type GetAviCloudsHandler interface {
	Handle(GetAviCloudsParams) middleware.Responder
}

// NewGetAviClouds creates a new http.Handler for the get avi clouds operation
func NewGetAviClouds(ctx *middleware.Context, handler GetAviCloudsHandler) *GetAviClouds {
	return &GetAviClouds{Context: ctx, Handler: handler}
}

/*
GetAviClouds swagger:route GET /api/avi/clouds avi getAviClouds

Retrieve Avi load balancer clouds
*/
type GetAviClouds struct {
	Context *middleware.Context
	Handler GetAviCloudsHandler
}

func (o *GetAviClouds) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetAviCloudsParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
