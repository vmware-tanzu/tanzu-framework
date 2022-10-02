// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package plugintemplates provides templates for CLI doc generation
package plugintemplates

import _ "embed" // Import files for plugin templates

// CoreREADME contains the cobra cli docs readme template
//
//go:embed README.md.tmpl
var CoreREADME string
