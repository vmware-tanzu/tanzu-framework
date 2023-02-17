// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu/tanzu-framework/tkg/client"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

// GetWorkloadClusterCredentialsOptions options that can be passed while getting workload cluster credentials
type GetWorkloadClusterCredentialsOptions struct {
	ClusterName string
	Namespace   string
	ExportFile  string
}

// this is likely the func responsible for creating kubeconfig files,
// either admin kubeconfig files, or pinniped specific kubeconfig files.
// GetCredentials saves cluster credentials to a file
func (t *tkgctl) GetCredentials(options GetWorkloadClusterCredentialsOptions) error {
	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	getWorkloadClusterCredentialsOptions := client.GetWorkloadClusterCredentialsOptions{
		ClusterName: options.ClusterName,
		Namespace:   options.Namespace,
		// this is the file that will be written if provided.
		ExportFile: options.ExportFile,
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
