// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
)

// GetTKRByName gets TKR object given a TKR name
func GetTKRByName(ctx context.Context, c client.Client, tkrName string) (*runtanzuv1alpha1.TanzuKubernetesRelease, error) {
	if tkrName == "" {
		return nil, nil
	}

	tkr := &runtanzuv1alpha1.TanzuKubernetesRelease{}

	tkrNamespaceName := client.ObjectKey{
		Name: tkrName,
	}

	if err := c.Get(ctx, tkrNamespaceName, tkr); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return tkr, nil
}
