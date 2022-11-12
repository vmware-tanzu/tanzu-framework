// Copyright 2021-2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
)

const (
	defaultDataValueNameSpace          = "tkg-system"
	defaultDataValueDeploymentReplicas = 3
)

// getOwnerCluster verifies that the AzureFileCSIConfig has a cluster as its owner reference,
// and returns the cluster. It tries to read the cluster name from the AzureFileCSIConfig's owner reference objects.
// If not there, we assume the owner cluster and AzureFileCSIConfig always has the same name.
func (r *AzureFileCSIConfigReconciler) getOwnerCluster(ctx context.Context,
	azurefileCSIConfig *csiv1alpha1.AzureFileCSIConfig) (*clusterapiv1beta1.Cluster, error) {

	logger := log.FromContext(ctx)
	cluster := &clusterapiv1beta1.Cluster{}
	clusterName := azurefileCSIConfig.Name // usually the corresponding 'cluster' shares the same name

	// retrieve the owner cluster for the AzureFileCSIConfig object
	for _, ownerRef := range azurefileCSIConfig.GetOwnerReferences() {
		if strings.EqualFold(ownerRef.Kind, constants.ClusterKind) {
			clusterName = ownerRef.Name
			break
		}
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: azurefileCSIConfig.Namespace, Name: clusterName}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Cluster resource '%s/%s' not found", azurefileCSIConfig.Namespace, clusterName))
			return nil, nil
		}
		logger.Error(err, fmt.Sprintf("Unable to fetch cluster '%s/%s'", azurefileCSIConfig.Namespace, clusterName))
		return nil, err
	}

	return cluster, nil
}

// mapVSphereCSIConfigToDataValues maps VSphereCSIConfig CR to data values
func (r *AzureFileCSIConfigReconciler) mapAzureFileCSIConfigToDataValues(ctx context.Context,
	azureFileCSIConfig *csiv1alpha1.AzureFileCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {

	dvs := &DataValues{
		AzureFileCSI: &DataValuesAzureFileCSI{
			Namespace:          defaultDataValueNameSpace,
			DeploymentReplicas: defaultDataValueDeploymentReplicas,
			HTTPProxy:          "",
			HTTPSProxy:         "",
			NoProxy:            "",
		},
	}

	if azureFileCSIConfig.Spec.AzureFileCSI.Namespace != "" {
		dvs.AzureFileCSI.Namespace = azureFileCSIConfig.Spec.AzureFileCSI.Namespace
	}
	dvs.AzureFileCSI.Namespace = azureFileCSIConfig.Spec.AzureFileCSI.Namespace
	dvs.AzureFileCSI.HTTPProxy = azureFileCSIConfig.Spec.AzureFileCSI.HTTPProxy
	dvs.AzureFileCSI.HTTPSProxy = azureFileCSIConfig.Spec.AzureFileCSI.HTTPSProxy
	dvs.AzureFileCSI.NoProxy = azureFileCSIConfig.Spec.AzureFileCSI.NoProxy

	if azureFileCSIConfig.Spec.AzureFileCSI.DeploymentReplicas != nil {
		dvs.AzureFileCSI.DeploymentReplicas = *azureFileCSIConfig.Spec.AzureFileCSI.DeploymentReplicas
	}

	return dvs, nil
}
