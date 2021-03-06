// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/swag"
)

// VSphereThumbprint v sphere thumbprint
// swagger:model VSphereThumbprint
type VSphereThumbprint struct {

	// insecure
	Insecure *bool `json:"insecure,omitempty"`

	// thumbprint
	Thumbprint string `json:"thumbprint,omitempty"`
}

// Validate validates this v sphere thumbprint
func (m *VSphereThumbprint) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *VSphereThumbprint) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *VSphereThumbprint) UnmarshalBinary(b []byte) error {
	var res VSphereThumbprint
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
