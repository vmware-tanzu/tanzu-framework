// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

// RegisterOptions options that can be passed while registering a cluster
type RegisterOptions struct {
	ClusterName        string
	TMCRegistrationURL string
}

// RegisterWithTmc registers management cluster with TMC
func (t *tkgctl) RegisterWithTmc(options RegisterOptions) error {
	log.Infof("Registering management cluster %s to TMC...\n", options.ClusterName)
	if !utils.IsValidURL(options.TMCRegistrationURL) {
		return errors.Errorf("TMC registration URL '%s' is not valid", options.TMCRegistrationURL)
	}
	return t.tkgClient.RegisterManagementClusterToTmc(options.ClusterName, options.TMCRegistrationURL)
}
