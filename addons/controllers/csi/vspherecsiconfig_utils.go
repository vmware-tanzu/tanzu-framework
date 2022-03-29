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
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	pkgtypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/csi/v1alpha1"
)

// mapVSphereCSIConfigToDataValues maps VSphereCSIConfig CR to data values
func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValues(ctx context.Context,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {
	switch vcsiConfig.Spec.VSphereCSI.Mode {
	case VSphereCSINonParavirtualMode:
		return r.mapVSphereCSIConfigToDataValuesNonParavirtual(ctx, vcsiConfig, cluster)
	case VSphereCSIParavirtualMode:
		return r.mapVSphereCSIConfigToDataValuesParavirtual(ctx, vcsiConfig, cluster)
	default:
		break
	}
	return nil, errors.Errorf("Invalid CSI mode '%s', must either be '%s' or '%s'",
		vcsiConfig.Spec.VSphereCSI.Mode, VSphereCSIParavirtualMode, VSphereCSINonParavirtualMode)
}

func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValuesParavirtual(ctx context.Context,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {
	dvs := &DataValues{}
	config := vcsiConfig.Spec.VSphereCSI.ParavirtualConfig

	if config == nil {
		return nil, errors.Errorf("ParavirtualConfig missing in Spec.VSphereCSI")
	}

	// first populate derived values
	dvs.VSpherePVCSI = &DataValuesVSpherePVCSI{}

	dvs.VSpherePVCSI.Namespace = tryParseClusterVariableString(ctx, cluster, NamespaceVarName)
	dvs.VSpherePVCSI.ClusterName = tryParseClusterVariableString(ctx, cluster, ClusterNameVarName)

	dvs.VSpherePVCSI.ClusterUID = string(cluster.UID) // only place to get UID

	// default values from https://github.com/vmware-tanzu/community-edition/blob/main/addons/packages/vsphere-pv-csi/2.4.1/bundle/config/values.yaml
	dvs.VSpherePVCSI.SupervisorMasterEndpointHostname = "supervisor.default.svc"
	dvs.VSpherePVCSI.SupervisorMasterPort = 6443

	// override derived & default values IF set in csi config
	if config.Namespace != "" {
		dvs.VSpherePVCSI.Namespace = config.Namespace
	}

	if config.ClusterName != "" {
		dvs.VSpherePVCSI.ClusterName = config.ClusterName
	}

	if config.ClusterUID != "" {
		dvs.VSpherePVCSI.ClusterUID = config.ClusterUID
	}

	if config.SupervisorMasterEndpointHostname != "" {
		dvs.VSpherePVCSI.SupervisorMasterEndpointHostname = config.SupervisorMasterEndpointHostname
	}

	if config.SupervisorMasterPort != nil {
		dvs.VSpherePVCSI.SupervisorMasterPort = *config.SupervisorMasterPort
	}

	return dvs, nil
}

func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValuesNonParavirtual(ctx context.Context,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {
	dvs := &DataValues{}
	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig

	if config == nil {
		return nil, errors.Errorf("NonParavirtualConfig missing in Spec.VSphereCSI")
	}

	// populate derived values
	dvs.VSphereCSI = &DataValuesVSphereCSI{}

	dvs.VSphereCSI.Namespace = tryParseClusterVariableString(ctx, cluster, NamespaceVarName)
	dvs.VSphereCSI.ClusterName = tryParseClusterVariableString(ctx, cluster, ClusterNameVarName)
	dvs.VSphereCSI.Server = tryParseClusterVariableString(ctx, cluster, VSphereServerVarName)
	dvs.VSphereCSI.Datacenter = tryParseClusterVariableString(ctx, cluster, VSphereDatacenterVarName)
	dvs.VSphereCSI.PublicNetwork = tryParseClusterVariableString(ctx, cluster, VSphereNetworkVarName) //TODO: IS 'VSPHERE_NETWORK' same as 'public network'?
	dvs.VSphereCSI.Username = tryParseClusterVariableString(ctx, cluster, VSphereUsernameVarName)
	dvs.VSphereCSI.Password = tryParseClusterVariableString(ctx, cluster, VSpherePasswordVarName)
	dvs.VSphereCSI.Region = tryParseClusterVariableString(ctx, cluster, VSphereRegionVarName)
	dvs.VSphereCSI.Zone = tryParseClusterVariableString(ctx, cluster, VSphereZoneVarName)
	dvs.VSphereCSI.VSphereVersion = tryParseClusterVariableString(ctx, cluster, VSphereVersionVarName)
	dvs.VSphereCSI.WindowsSupport = tryParseClusterVariableBool(ctx, cluster, IsWindowsWorkloadClusterVarName)
	dvs.VSphereCSI.TLSThumbprint = tryParseClusterVariableString(ctx, cluster, VSphereTLSThumbprintVarName)

	if cluster.Annotations != nil {
		dvs.VSphereCSI.HttpProxy = cluster.Annotations[pkgtypes.HTTPProxyConfigAnnotation]
		dvs.VSphereCSI.HttpsProxy = cluster.Annotations[pkgtypes.HTTPSProxyConfigAnnotation]
		dvs.VSphereCSI.NoProxy = cluster.Annotations[pkgtypes.NoProxyConfigAnnotation]
	}

	// populated from default value in https://github.com/vmware-tanzu/community-edition/blob/main/addons/packages/vsphere-csi/2.4.1/bundle/config/values.yaml
	dvs.VSphereCSI.UseTopologyCategories = false
	dvs.VSphereCSI.ProvisionTimeout = "300s"
	dvs.VSphereCSI.AttachTimeout = "300s"
	dvs.VSphereCSI.ResizerTimeout = "300s"
	dvs.VSphereCSI.DeploymentReplicas = 3
	dvs.VSphereCSI.InsecureFlag = true

	// override derived values IF set in csi configuration by user
	if config.Namespace != "" {
		dvs.VSphereCSI.Namespace = config.Namespace
	}
	if config.ClusterName != "" {
		dvs.VSphereCSI.ClusterName = config.ClusterName
	}
	if config.Server != "" {
		dvs.VSphereCSI.Server = config.Server
	}
	if config.Datacenter != "" {
		dvs.VSphereCSI.Datacenter = config.Datacenter
	}
	if config.PublicNetwork != "" {
		dvs.VSphereCSI.PublicNetwork = config.PublicNetwork
	}
	if config.Username != "" {
		dvs.VSphereCSI.Username = config.Username
	}
	if config.Password != "" {
		dvs.VSphereCSI.Password = config.Password
	}
	if config.Region != "" {
		dvs.VSphereCSI.Region = config.Region
	}
	if config.Zone != "" {
		dvs.VSphereCSI.Zone = config.Zone
	}
	if config.VSphereVersion != "" {
		dvs.VSphereCSI.VSphereVersion = config.VSphereVersion
	}
	if config.WindowsSupport != nil {
		dvs.VSphereCSI.WindowsSupport = *config.WindowsSupport
	}
	if config.TLSThumbprint != "" {
		dvs.VSphereCSI.TLSThumbprint = config.TLSThumbprint
	}
	if config.HttpProxy != "" {
		dvs.VSphereCSI.HttpProxy = config.HttpProxy
	}
	if config.HttpsProxy != "" {
		dvs.VSphereCSI.HttpsProxy = config.HttpsProxy
	}
	if config.NoProxy != "" {
		dvs.VSphereCSI.NoProxy = config.NoProxy
	}
	if config.UseTopologyCategories != nil {
		dvs.VSphereCSI.UseTopologyCategories = *config.UseTopologyCategories
	}
	if config.ProvisionTimeout != "" {
		dvs.VSphereCSI.ProvisionTimeout = config.ProvisionTimeout
	}
	if config.AttachTimeout != "" {
		dvs.VSphereCSI.AttachTimeout = config.AttachTimeout
	}
	if config.ResizerTimeout != "" {
		dvs.VSphereCSI.ResizerTimeout = config.ResizerTimeout
	}
	if config.DeploymentReplicas != nil {
		dvs.VSphereCSI.DeploymentReplicas = *config.DeploymentReplicas
	}
	if config.InsecureFlag != nil {
		dvs.VSphereCSI.InsecureFlag = *config.InsecureFlag
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

// getVSphereCluster gets the VSphereCluster CR for the cluster object
func (r *VSphereCSIConfigReconciler) getVSphereCluster(ctx context.Context,
	cluster *clusterapiv1beta1.Cluster) (*capvv1beta1.VSphereCluster, error) {
	vsphereCluster := &capvv1beta1.VSphereCluster{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cluster.Name,
	}, vsphereCluster); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("VSphereCluster %s/%s not found", cluster.Namespace, cluster.Name)
		}
		return nil, errors.Errorf("VSphereCluster %s/%s could not be fetched, error %v", cluster.Namespace, cluster.Name, err)
	}
	return vsphereCluster, nil
}

// tryParseClusterVariableBool tries to parse a boolean cluster variable,
// info any error that occurs
func tryParseClusterVariableBool(ctx context.Context, cluster *clusterapiv1beta1.Cluster,
	variableName string) bool {
	res, err := util.ParseClusterVariableBool(cluster, variableName)
	if err != nil {
		log.FromContext(ctx).Error(err, "cannot parse cluster variable", "key", variableName)
	}
	return res
}

// tryParseClusterVariableString tries to parse a string cluster variable,
// info any error that occurs
func tryParseClusterVariableString(ctx context.Context, cluster *clusterapiv1beta1.Cluster,
	variableName string) string {
	res, err := util.ParseClusterVariableString(cluster, variableName)
	if err != nil {
		log.FromContext(ctx).Error(err, "cannot parse cluster variable", "key", variableName)
	}
	return res
}
