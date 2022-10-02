// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"os"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
)

// ConfigCluster prints cluster template to stdout
//
//nolint:gocritic
func (t *tkgctl) ConfigCluster(configClusterOption CreateClusterOptions) error {
	var err error

	configClusterOption.ClusterConfigFile, err = t.ensureClusterConfigFile(configClusterOption.ClusterConfigFile)
	if err != nil {
		return err
	}

	// configures missing create cluster options from config file variables
	err = t.configureCreateClusterOptionsFromConfigFile(&configClusterOption)
	if err != nil {
		return err
	}

	isInputFileClusterClassBased := false
	// TODO (chandrareddyp) need to support Cluster.YAML file with Cluster Object for the dry-run use case, then based on input file type, pass the isInputFileClusterClassBased flag here (https://github.com/vmware-tanzu/tanzu-framework/issues/2516)
	options, err := t.getCreateClusterOptions(configClusterOption.ClusterName, &configClusterOption, isInputFileClusterClassBased)
	if err != nil {
		return err
	}

	isPacific, err := t.tkgClient.IsPacificManagementCluster()
	if err != nil && strings.Contains(configClusterOption.InfrastructureProvider, client.PacificProviderName) {
		isPacific = true
	}

	if isPacific {
		// For TKGS kubernetesVersion will be same as TkrVersion
		options.KubernetesVersion = configClusterOption.TkrVersion
		options.TKRVersion = configClusterOption.TkrVersion
	} else {
		options.TKRVersion, options.KubernetesVersion, err = t.getAndDownloadTkrIfNeeded(configClusterOption.TkrVersion)
		if err != nil {
			return errors.Wrapf(err, "unable to determine the TKr version and kubernetes version based on '%v'", configClusterOption.TkrVersion)
		}
	}

	// Don't skip validation while creating cluster template
	options.SkipValidation = false

	yaml, err := t.tkgClient.GetClusterConfiguration(&options)
	if err != nil {
		return err
	}
	yaml = append(yaml, '\n')

	if _, err := os.Stdout.Write(yaml); err != nil {
		return errors.Wrap(err, "failed to write yaml to Stdout")
	}

	return nil
}
