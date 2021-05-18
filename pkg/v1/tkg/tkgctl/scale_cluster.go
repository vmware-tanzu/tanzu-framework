// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

// ScaleClusterOptions options that can be passed while scaling a cluster
type ScaleClusterOptions struct {
	ClusterName       string
	WorkerCount       int32
	ControlPlaneCount int32
	Namespace         string
}

// ScaleCluster scales cluster
func (t *tkgctl) ScaleCluster(options ScaleClusterOptions) error {
	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	if options.ControlPlaneCount <= 0 && options.WorkerCount <= 0 {
		return errors.New("incorrect machine counts provided. Machine count value for control-plane and workers must be greater than 0")
	}

	scaleClusterOptions := client.ScaleClusterOptions{
		Kubeconfig:        t.kubeconfig,
		Namespace:         options.Namespace,
		ClusterName:       options.ClusterName,
		WorkerCount:       options.WorkerCount,
		ControlPlaneCount: options.ControlPlaneCount,
	}

	err := t.tkgClient.ScaleCluster(scaleClusterOptions)
	if err != nil {
		return err
	}

	log.Infof("Workload cluster '%s' is being scaled\n", scaleClusterOptions.ClusterName)
	return nil
}
