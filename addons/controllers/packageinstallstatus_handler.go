// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
)

// pkgiToCluster returns a list of Requests with Cluster ObjectKey
func pkgiToCluster(o client.Object) []ctrl.Request {
	pkgi, ok := o.(*kapppkgiv1alpha1.PackageInstall)
	if !ok {
		return nil
	}

	clusterObjKey := getClusterNamespacedName(pkgi)
	if clusterObjKey == nil {
		return nil
	}

	return []ctrl.Request{{NamespacedName: client.ObjectKey{Namespace: clusterObjKey.Namespace, Name: clusterObjKey.Name}}}
}
