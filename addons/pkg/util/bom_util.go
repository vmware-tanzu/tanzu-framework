// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"

	"github.com/vmware-tanzu-private/core/addons/pkg/constants"
	bomtypes "github.com/vmware-tanzu-private/core/tkr/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GetBOMByTKRName returns the bom associated with the TKR
func GetBOMByTKRName(ctx context.Context, c client.Client, tkrName string) (*bomtypes.Bom, error) {
	configMapList := &corev1.ConfigMapList{}
	var bomConfigMap *corev1.ConfigMap
	if err := c.List(ctx, configMapList, client.InNamespace(constants.TKGBomNamespace), client.MatchingLabels{constants.TKRLabel: tkrName}); err != nil {
		return nil, err
	}

	if len(configMapList.Items) <= 0 {
		return nil, nil
	}

	bomConfigMap = &configMapList.Items[0]
	bomData, ok := bomConfigMap.BinaryData[constants.TKGBomContent]
	if !ok {
		bomDataString, ok := bomConfigMap.Data[constants.TKGBomContent]
		if !ok {
			return nil, nil
		}
		bomData = []byte(bomDataString)
	}

	bom, err := bomtypes.NewBom(bomData)
	if err != nil {
		return nil, err
	}

	return &bom, nil
}

// GetTKRNameFromBOMConfigMap returns tkr name given a bom configmap
func GetTKRNameFromBOMConfigMap(bomConfigMap *corev1.ConfigMap) string {
	return bomConfigMap.Labels[constants.TKRLabel]
}
