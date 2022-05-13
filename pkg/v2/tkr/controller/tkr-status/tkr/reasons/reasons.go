// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package reasons provides error types useful for calculating the Valid status condition.
package reasons

import "reflect"

type HasReason interface {
	error
	Reason() string
}

type TKRVersionMismatch string

func (e TKRVersionMismatch) Error() string {
	return string(e)
}

func (e TKRVersionMismatch) Reason() string {
	return reflect.TypeOf(e).Name()
}

type OSImageVersionMismatch string

func (e OSImageVersionMismatch) Error() string {
	return string(e)
}

func (e OSImageVersionMismatch) Reason() string {
	return reflect.TypeOf(e).Name()
}

type MissingOSImage string

func (e MissingOSImage) Error() string {
	return string(e)
}

func (e MissingOSImage) Reason() string {
	return reflect.TypeOf(e).Name()
}

type MissingBootstrapPackage string

func (e MissingBootstrapPackage) Error() string {
	return string(e)
}

func (e MissingBootstrapPackage) Reason() string {
	return reflect.TypeOf(e).Name()
}

type MissingClusterBootstrapTemplate string

func (e MissingClusterBootstrapTemplate) Error() string {
	return string(e)
}

func (e MissingClusterBootstrapTemplate) Reason() string {
	return reflect.TypeOf(e).Name()
}
