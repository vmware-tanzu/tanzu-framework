// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/cpi/v1alpha1"
)

// ClusterToOracleCPIConfig returns a list of Requests with OracleCPIConfig ObjectKey based on Cluster events
func (r *OracleCPIConfigReconciler) ClusterToOracleCPIConfig(o client.Object) []ctrl.Request {
	cluster, ok := o.(*clusterapiv1beta1.Cluster)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive Cluster resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	r.Log.V(4).Info("Mapping Cluster to OracleCPIConfig")

	cs := &cpiv1alpha1.OracleCPIConfigList{}
	_ = r.List(context.Background(), cs)
	var requests []ctrl.Request

	for _, cpiConfig := range cs.Items {
		requests = forClusterMappedByCPIConfigEnqueueRequest(&cpiConfig, cluster, r.Config.ConfigControllerConfig, r.Log, requests)
	}

	return requests
}
