// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/web/server/models"
)

// Err Define the error message
func Err(e error) *models.Error {
	err := models.Error{Message: e.Error()}
	return &err
}
