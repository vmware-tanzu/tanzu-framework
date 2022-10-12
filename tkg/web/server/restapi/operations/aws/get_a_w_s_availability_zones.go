// Code generated by go-swagger; DO NOT EDIT.

package aws

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	middleware "github.com/go-openapi/runtime/middleware"
)

// GetAWSAvailabilityZonesHandlerFunc turns a function with the right signature into a get a w s availability zones handler
type GetAWSAvailabilityZonesHandlerFunc func(GetAWSAvailabilityZonesParams) middleware.Responder

// Handle executing the request and returning a response
func (fn GetAWSAvailabilityZonesHandlerFunc) Handle(params GetAWSAvailabilityZonesParams) middleware.Responder {
	return fn(params)
}

// GetAWSAvailabilityZonesHandler interface for that can handle valid get a w s availability zones params
type GetAWSAvailabilityZonesHandler interface {
	Handle(GetAWSAvailabilityZonesParams) middleware.Responder
}

// NewGetAWSAvailabilityZones creates a new http.Handler for the get a w s availability zones operation
func NewGetAWSAvailabilityZones(ctx *middleware.Context, handler GetAWSAvailabilityZonesHandler) *GetAWSAvailabilityZones {
	return &GetAWSAvailabilityZones{Context: ctx, Handler: handler}
}

/*
GetAWSAvailabilityZones swagger:route GET /api/providers/aws/AvailabilityZones aws getAWSAvailabilityZones

Retrieve AWS availability zones of current region
*/
type GetAWSAvailabilityZones struct {
	Context *middleware.Context
	Handler GetAWSAvailabilityZonesHandler
}

func (o *GetAWSAvailabilityZones) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		r = rCtx
	}
	var Params = NewGetAWSAvailabilityZonesParams()

	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request

	o.Context.Respond(rw, r, route.Produces, route, res)

}
