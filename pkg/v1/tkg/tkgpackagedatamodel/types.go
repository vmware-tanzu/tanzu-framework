// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// nolint
package tkgpackagedatamodel

import (
	"fmt"

	"github.com/pkg/errors"
)

// PackagePluginNonCriticalError is used for non critical package plugin errors which should be treated more like warnings
type PackagePluginNonCriticalError struct {
	Reason string
}

func (e *PackagePluginNonCriticalError) Error() string { return e.Reason }

type ResourceType int

const (
	ResourceTypePackageInstall ResourceType = iota
	ResourceTypePackageRepository
)

type OperationType int

const (
	OperationTypeInstall OperationType = iota
	OperationTypeUpdate
)

// TypeBoolPtr satisfies Value interface defined in "https://github.com/spf13/pflag/blob/master/flag.go"
type TypeBoolPtr struct {
	ExportToAllNamespaces *bool
}

// Type returns the default type for a TypeBoolPtr variable
func (v *TypeBoolPtr) Type() string {
	return ""
}

// Set sets the TypeBoolPtr variable based on the string argument
func (v *TypeBoolPtr) Set(val string) error {
	f := false
	t := true
	if val == "true" || val == "True" {
		v.ExportToAllNamespaces = &t
	} else if val == "false" || val == "False" {
		v.ExportToAllNamespaces = &f
	} else if val != "" {
		return errors.New(fmt.Sprintf("invalid argument '%s'", val))
	}

	return nil
}

// String returns the string representation of a TypeBoolPtr variable
func (v *TypeBoolPtr) String() string {
	if v.ExportToAllNamespaces != nil {
		return fmt.Sprint(*v.ExportToAllNamespaces)
	}
	return ""
}
