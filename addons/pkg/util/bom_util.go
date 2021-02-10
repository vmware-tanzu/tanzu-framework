// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"

	"github.com/vmware-tanzu-private/core/addons/pkg/constants"
	bomtypes "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

// GetAddonImageRepository returns imageRepository from configMap `tkr-controller-config` in namespace `tkr-system` if exists else use BOM
func GetAddonImageRepository(ctx context.Context, c client.Client, bom *bomtypes.Bom) (string, error) {
	bomConfigMap := &corev1.ConfigMap{}
	err := c.Get(ctx, client.ObjectKey{
		Namespace: constants.TKGBomNamespace,
		Name:      constants.TKRConfigmapName,
	}, bomConfigMap)

	// if the configmap exists, try get repository from the config map
	if err == nil {
		if imgRepo, ok := bomConfigMap.Data[constants.TKRRepoKey]; ok && imgRepo != "" {
			return imgRepo, nil
		}
	} else if !errors.IsNotFound(err) {
		// return the error to controller if err is not IsNotFound
		return "", err
	}
	// if the configmap doesn't exist, or there is no `imageRepository` field, get repository from BOM
	return bom.GetImageRepository()
}
