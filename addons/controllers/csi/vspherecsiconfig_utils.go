// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/csi/v1alpha1"
)

// mapVSphereCSIConfigToDataValues maps VSphereCSIConfig CR to data values
func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValues(ctx context.Context, vcsiConfig *csiv1alpha1.VSphereCSIConfig) (*DataValues, error) {
	switch vcsiConfig.Spec.VSphereCSI.Mode {
	case VSphereCSINonParavirtualMode:
		return r.mapVSphereCSIConfigToDataValuesNonParavirtual(ctx, vcsiConfig)
	case VSphereCSIParavirtualMode:
		return r.mapVSphereCSIConfigToDataValuesParavirtual(ctx, vcsiConfig)
	default:
		break
	}
	return nil, errors.Errorf("Invalid CSI mode '%s', must either be '%s' or '%s'",
		vcsiConfig.Spec.VSphereCSI.Mode, VSphereCSIParavirtualMode, VSphereCSINonParavirtualMode)
}

func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValuesParavirtual(ctx context.Context,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) (*DataValues, error) {
	dvs := &DataValues{}
	pvconfig := vcsiConfig.Spec.VSphereCSI.ParavirtualConfig

	if pvconfig == nil {
		return nil, errors.Errorf("ParavirtualConfig missing in Spec.VSphereCSI")
	}

	dvs.VSpherePVCSI = &DataValuesVSpherePVCSI{ClusterName: pvconfig.ClusterName,
		ClusterUID:                       pvconfig.ClusterUID,
		Namespace:                        pvconfig.Namespace,
		SupervisorMasterEndpointHostname: pvconfig.SupervisorMasterEndpointHostname,
		SupervisorMasterPort:             pvconfig.SupervisorMasterPort}

	return dvs, nil
}

func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValuesNonParavirtual(ctx context.Context, vcsiConfig *csiv1alpha1.VSphereCSIConfig) (*DataValues, error) {
	dvs := &DataValues{}
	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig

	if config == nil {
		return nil, errors.Errorf("NonParavirtualConfig missing in Spec.VSphereCSI")
	}

	dvs.VSphereCSI = &DataValuesVSphereCSI{
		Namespace:             config.Namespace,
		ClusterName:           config.ClusterName,
		Server:                config.Server,
		Datacenter:            config.Datacenter,
		PublicNetwork:         config.PublicNetwork,
		Username:              config.Username,
		Password:              config.Password,
		Region:                config.Region,
		Zone:                  config.Zone,
		UseTopologyCategories: config.UseTopologyCategories,
		ProvisionTimeout:      config.ProvisionTimeout,
		AttachTimeout:         config.AttachTimeout,
		ResizerTimeout:        config.ResizerTimeout,
		VSphereVersion:        config.VSphereVersion,
		HttpProxy:             config.HttpProxy,
		HttpsProxy:            config.HttpsProxy,
		NoProxy:               config.NoProxy,
		DeploymentReplicas:    config.DeploymentReplicas,
		WindowsSupport:        config.WindowsSupport,
	}
	return dvs, nil
}

// getOwnerCluster verifies that the VSphereCSIConfig has a cluster as its owner reference,
// and returns the cluster. It tries to read the cluster name from the VSphereCSIConfig's owner reference objects.
// If not there, we assume the owner cluster and VSphereCSIConfig always has the same name.
func (r *VSphereCSIConfigReconciler) getOwnerCluster(ctx context.Context,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) (*clusterapiv1beta1.Cluster, error) {

	logger := log.FromContext(ctx)
	cluster := &clusterapiv1beta1.Cluster{}
	clusterName := vcsiConfig.Name

	// retrieve the owner cluster for the VSphereCSIConfig object
	for _, ownerRef := range vcsiConfig.GetOwnerReferences() {
		if strings.EqualFold(ownerRef.Kind, constants.ClusterKind) {
			clusterName = ownerRef.Name
			break
		}
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: vcsiConfig.Namespace, Name: clusterName}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Cluster resource '%s/%s' not found", vcsiConfig.Namespace, clusterName))
			return nil, nil
		}
		logger.Error(err, fmt.Sprintf("Unable to fetch cluster '%s/%s'", vcsiConfig.Namespace, clusterName))
		return nil, err
	}

	return cluster, nil
}
