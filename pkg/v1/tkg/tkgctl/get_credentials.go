// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

// GetWorkloadClusterCredentialsOptions options that can be passed while getting workload cluster credentials
type GetWorkloadClusterCredentialsOptions struct {
	ClusterName string
	Namespace   string
	ExportFile  string
}

// GetCredentials saves cluster credentials to a file
func (t *tkgctl) GetCredentials(options GetWorkloadClusterCredentialsOptions) error {
	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	getWorkloadClusterCredentialsOptions := client.GetWorkloadClusterCredentialsOptions{
		ClusterName: options.ClusterName,
		Namespace:   options.Namespace,
		ExportFile:  options.ExportFile,
	}

	context, path, err := t.tkgClient.GetWorkloadClusterCredentials(getWorkloadClusterCredentialsOptions)
	if err != nil {
		return err
	}

	log.Infof("Credentials of cluster '%s' have been saved \n", options.ClusterName)
	if path == "" {
		log.Infof("You can now access the cluster by running 'kubectl config use-context %s'\n", context)
	} else {
		log.Infof("You can now access the cluster by running 'kubectl config use-context %s' under path '%s' \n", context, path)
	}

	return nil
}
