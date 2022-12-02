// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package featuregateclient

import "fmt"

// ErrType is a machine readable value created for facilitating error-matching
// in tests.
type ErrType string

const (
	// ErrType indicates a resource could not be found.
	ErrTypeNotFound ErrType = "NotFound"
	// ErrTypeForbidden indicates an action is not allowed.
	ErrTypeForbidden ErrType = "Forbidden"
	// ErrTypeTooMany indicates there are too many of a resource.
	ErrTypeTooMany ErrType = "TooMany"
)

// Error converts a ErrorType into its corresponding canonical error message.
func (t ErrType) Error() string {
	switch t {
	case ErrTypeNotFound:
		return "Not found"
	case ErrTypeForbidden:
		return "Forbidden"
	case ErrTypeTooMany:
		return "Too many"
	default:
		return fmt.Sprintf("unrecognized validation error: %q", string(t))
	}
}
