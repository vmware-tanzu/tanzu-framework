// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kapppkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
)

// pkgiToCluster returns a list of Requests with Cluster ObjectKey
func (r *PackageInstallStatusReconciler) pkgiToCluster(o client.Object) []ctrl.Request {
	pkgi, ok := o.(*kapppkgiv1alpha1.PackageInstall)
	if !ok {
		r.Log.Error(errors.New("invalid type"),
			"Expected to receive PackageInstall resource",
			"actualType", fmt.Sprintf("%T", o))
		return nil
	}

	r.Log.WithValues("pkgi-name", pkgi.Name).Info("Mapping PackageInstalls to cluster")

	clusterObjKey := r.getClusterNamespacedName(pkgi)
	if clusterObjKey == nil {
		return nil
	}

	return []ctrl.Request{{NamespacedName: client.ObjectKey{Namespace: clusterObjKey.Namespace, Name: clusterObjKey.Name}}}
}
