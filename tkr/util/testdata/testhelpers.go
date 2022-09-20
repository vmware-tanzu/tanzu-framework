// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package testdata

import (
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Request(o client.Object) ctrl.Request {
	return ctrl.Request{NamespacedName: types.NamespacedName{Namespace: o.GetNamespace(), Name: o.GetName()}}
}
