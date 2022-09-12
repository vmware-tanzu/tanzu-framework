// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"fmt"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

var (
	azureVarsToBeValidated = []string{constants.ConfigVariableAzureClientSecret}
)

// ValidateEnvVariables validates the presence of required environment variables for a given IaaS
func (c *TkgClient) ValidateEnvVariables(iaas string) error {
	// This function only contains a validator for Azure environment variables for now
	// This function can be used to add validators for environment variables pertaining to other IaaSes in the future
	// as and when required
	if iaas == AzureProviderName {
		return c.validateAzureEnvVariables(azureVarsToBeValidated)
	}
	return nil
}

func (c *TkgClient) validateAzureEnvVariables(azureEnvVars []string) error {
	for _, envVar := range azureEnvVars {
		_, err := c.TKGConfigReaderWriter().Get(envVar)
		if err != nil {
			return fmt.Errorf("config Variable %s not set", envVar)
		}
	}
	return nil
}
