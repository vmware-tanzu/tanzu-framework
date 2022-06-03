// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capictrlpkubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	cutil "github.com/vmware-tanzu/tanzu-framework/addons/controllers/utils"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	pkgtypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
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
		return r.mapVSphereCSIConfigToDataValuesParavirtual(ctx, cluster)
	default:
		break
	}
	//TODO: implement validation webhook to prevent this https://github.com/vmware-tanzu/tanzu-framework/issues/2087
	return nil, errors.Errorf("Invalid CSI mode '%s', must either be '%s' or '%s'",
		vcsiConfig.Spec.VSphereCSI.Mode, VSphereCSIParavirtualMode, VSphereCSINonParavirtualMode)
}

func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValuesParavirtual(_ context.Context,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {

	dvs := &DataValues{}
	dvs.VSpherePVCSI = &DataValuesVSpherePVCSI{}
	dvs.VSpherePVCSI.ClusterName = cluster.Name
	dvs.VSpherePVCSI.ClusterUID = string(cluster.UID)
	// default values from https://github.com/vmware-tanzu/community-edition/blob/main/addons/packages/vsphere-pv-csi/2.4.1/bundle/config/values.yaml
	dvs.VSpherePVCSI.Namespace = "vmware-system-csi"
	dvs.VSpherePVCSI.SupervisorMasterEndpointHostname = "supervisor.default.svc"
	dvs.VSpherePVCSI.SupervisorMasterPort = 6443

	return dvs, nil
}

func (r *VSphereCSIConfigReconciler) mapVSphereCSIConfigToDataValuesNonParavirtual(ctx context.Context,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {

	dvs := &DataValues{}

	// populate derived values
	dvs.VSphereCSI = &DataValuesVSphereCSI{}
	dvs.VSphereCSI.ClusterName = cluster.Name

	vsphereCluster, err := cutil.GetVSphereCluster(ctx, r.Client, cluster)
	if err != nil {
		return nil, err
	}
	dvs.VSphereCSI.Server = vsphereCluster.Spec.Server
	dvs.VSphereCSI.TLSThumbprint = vsphereCluster.Spec.Thumbprint

	cpMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cutil.ControlPlaneName(cluster.Name),
	}, cpMachineTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("VSphereMachineTemplate %s/%s not found", cluster.Namespace, cutil.ControlPlaneName(cluster.Name))
		}
		return nil, errors.Errorf("VSphereMachineTemplate %s/%s could not be fetched, error %v", cluster.Namespace, cutil.ControlPlaneName(cluster.Name), err)
	}

	dvs.VSphereCSI.Datacenter = cpMachineTemplate.Spec.Template.Spec.Datacenter
	dvs.VSphereCSI.PublicNetwork = cutil.TryParseClusterVariableString(ctx, cluster, VSphereNetworkVarName)

	// derive vSphere username and password from the <cluster name> secret
	clusterSecret, err := cutil.GetSecret(ctx, r.Client, cluster.Namespace, cluster.Name)
	if err != nil {
		return nil, err
	}
	dvs.VSphereCSI.Username, dvs.VSphereCSI.Password, err = cutil.GetUsernameAndPasswordFromSecret(clusterSecret)
	if err != nil {
		return nil, err
	}

	dvs.VSphereCSI.Region = cutil.TryParseClusterVariableString(ctx, cluster, VSphereRegionVarName)
	dvs.VSphereCSI.Zone = cutil.TryParseClusterVariableString(ctx, cluster, VSphereZoneVarName)
	dvs.VSphereCSI.VSphereVersion = cutil.TryParseClusterVariableString(ctx, cluster, VSphereVersionVarName)
	dvs.VSphereCSI.WindowsSupport = cutil.TryParseClusterVariableBool(ctx, cluster, IsWindowsWorkloadClusterVarName)

	if cluster.Annotations != nil {
		dvs.VSphereCSI.HTTPProxy = cluster.Annotations[pkgtypes.HTTPProxyConfigAnnotation]
		dvs.VSphereCSI.HTTPSProxy = cluster.Annotations[pkgtypes.HTTPSProxyConfigAnnotation]
		dvs.VSphereCSI.NoProxy = cluster.Annotations[pkgtypes.NoProxyConfigAnnotation]
	}

	// populated from default value in https://github.com/vmware-tanzu/community-edition/blob/main/addons/packages/vsphere-csi/2.4.1/bundle/config/values.yaml
	dvs.VSphereCSI.Namespace = VSphereCSINamespace
	dvs.VSphereCSI.UseTopologyCategories = false
	dvs.VSphereCSI.ProvisionTimeout = VSphereCSIProvisionTimeout
	dvs.VSphereCSI.AttachTimeout = VSphereCSIAttachTimeout
	dvs.VSphereCSI.ResizerTimeout = VSphereCSIResizerTimeout
	deploymentReplicas, err := r.computeRecommendedNumberOfDeploymentReplicas(ctx, cluster)
	if err != nil {
		return nil, errors.Errorf("Failed to set number of vsphere csi deployment replicas: '%v'", err)
	}
	dvs.VSphereCSI.DeploymentReplicas = deploymentReplicas
	dvs.VSphereCSI.InsecureFlag = true

	// TODO: implement defaulting webhook for 'vspherecsiconfig' https://github.com/vmware-tanzu/tanzu-framework/issues/2088
	// override derived values IF set in csi configuration by user
	if vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig != nil {
		if err := r.overrideDerivedValues(ctx, dvs.VSphereCSI, vcsiConfig); err != nil {
			return nil, errors.Errorf("Failed to override derived values: '%v'", err)
		}
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
	clusterName := vcsiConfig.Name // usually the corresponding 'cluster' shares the same name

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

// mapCSIConfigToProviderServiceAccount maps CSIConfig and cluster to the corresponding service account
func (r *VSphereCSIConfigReconciler) mapCSIConfigToProviderServiceAccount(vsphereCluster *capvvmwarev1beta1.VSphereCluster) *capvvmwarev1beta1.ProviderServiceAccount {
	serviceAccount := &capvvmwarev1beta1.ProviderServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-%s", vsphereCluster.Name, "pvcsi"),
			Namespace: vsphereCluster.Namespace,
		},
		Spec: capvvmwarev1beta1.ProviderServiceAccountSpec{
			Ref:              &v1.ObjectReference{Name: vsphereCluster.Name, Namespace: vsphereCluster.Namespace},
			Rules:            providerServiceAccountRBACRules,
			TargetNamespace:  "vmware-system-csi",
			TargetSecretName: "pvcsi-provider-creds",
		},
	}
	return serviceAccount
}

func (r *VSphereCSIConfigReconciler) overrideDerivedValues(ctx context.Context,
	dvscsi *DataValuesVSphereCSI,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) error {

	r.overrideProxyValues(dvscsi, vcsiConfig)
	r.overrideTimeoutValues(dvscsi, vcsiConfig)
	r.overrideTopologyValues(ctx, dvscsi, vcsiConfig)
	r.overrideClusterValues(dvscsi, vcsiConfig)
	r.overrideMiscValues(dvscsi, vcsiConfig)

	return r.overrideCredentialValues(ctx, dvscsi, vcsiConfig)
}

func (r *VSphereCSIConfigReconciler) overrideProxyValues(dvscsi *DataValuesVSphereCSI,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) {

	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig
	if config.HTTPProxy != "" {
		dvscsi.HTTPProxy = config.HTTPProxy
	}
	if config.HTTPSProxy != "" {
		dvscsi.HTTPSProxy = config.HTTPSProxy
	}
	if config.NoProxy != "" {
		dvscsi.NoProxy = config.NoProxy
	}
}

func (r *VSphereCSIConfigReconciler) overrideTimeoutValues(dvscsi *DataValuesVSphereCSI,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) {

	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig
	if config.ProvisionTimeout != "" {
		dvscsi.ProvisionTimeout = config.ProvisionTimeout
	}
	if config.AttachTimeout != "" {
		dvscsi.AttachTimeout = config.AttachTimeout
	}
	if config.ResizerTimeout != "" {
		dvscsi.ResizerTimeout = config.ResizerTimeout
	}
}

func (r *VSphereCSIConfigReconciler) overrideTopologyValues(ctx context.Context,
	dvscsi *DataValuesVSphereCSI,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) {

	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig
	if config.Server != "" {
		dvscsi.Server = config.Server
	}
	if config.Datacenter != "" {
		dvscsi.Datacenter = config.Datacenter
	}
	if config.PublicNetwork != "" {
		dvscsi.PublicNetwork = config.PublicNetwork
	}
	if config.UseTopologyCategories != nil {
		dvscsi.UseTopologyCategories = *config.UseTopologyCategories
	}
	if config.DeploymentReplicas != nil {
		dvscsi.DeploymentReplicas = r.constrainNumberOfDeploymentReplicas(ctx, *config.DeploymentReplicas)
	}
}

func (r *VSphereCSIConfigReconciler) overrideCredentialValues(ctx context.Context,
	dvscsi *DataValuesVSphereCSI,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) error {

	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig
	if config.VSphereCredentialLocalObjRef != nil {
		secret, err := cutil.GetSecret(ctx, r.Client, vcsiConfig.Namespace, config.VSphereCredentialLocalObjRef.Name)
		if err != nil {
			return err
		}
		dvscsi.Username, dvscsi.Password, err = cutil.GetUsernameAndPasswordFromSecret(secret)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *VSphereCSIConfigReconciler) overrideClusterValues(dvscsi *DataValuesVSphereCSI,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) {

	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig
	if config.Namespace != "" {
		dvscsi.Namespace = config.Namespace
	}
	if config.ClusterName != "" {
		dvscsi.ClusterName = config.ClusterName
	}

	if config.Region != "" {
		dvscsi.Region = config.Region
	}
	if config.Zone != "" {
		dvscsi.Zone = config.Zone
	}
}

func (r *VSphereCSIConfigReconciler) overrideMiscValues(dvscsi *DataValuesVSphereCSI,
	vcsiConfig *csiv1alpha1.VSphereCSIConfig) {

	config := vcsiConfig.Spec.VSphereCSI.NonParavirtualConfig
	if config.VSphereVersion != "" {
		dvscsi.VSphereVersion = config.VSphereVersion
	}
	if config.WindowsSupport != nil {
		dvscsi.WindowsSupport = *config.WindowsSupport
	}
	if config.TLSThumbprint != "" {
		dvscsi.TLSThumbprint = config.TLSThumbprint
	}
	if config.InsecureFlag != nil {
		dvscsi.InsecureFlag = *config.InsecureFlag
	}
}

func (r *VSphereCSIConfigReconciler) constrainNumberOfDeploymentReplicas(ctx context.Context, proposedCount int32) int32 {
	logger := log.FromContext(ctx)
	if proposedCount < VSphereCSIMinDeploymentReplicas {
		logger.Info(fmt.Sprintf("WARNING: adjusting vsphere csi replica count from '%d' to '%d'",
			proposedCount, VSphereCSIMinDeploymentReplicas))
		return VSphereCSIMinDeploymentReplicas
	}

	if proposedCount > VSphereCSIMaxDeploymentReplicas {
		logger.Info(fmt.Sprintf("WARNING: adjusting vsphere csi replica count from '%d' to '%d'",
			proposedCount, VSphereCSIMaxDeploymentReplicas))
		return VSphereCSIMaxDeploymentReplicas
	}

	return proposedCount
}

func (r *VSphereCSIConfigReconciler) computeRecommendedNumberOfDeploymentReplicas(ctx context.Context,
	cluster *clusterapiv1beta1.Cluster) (int32, error) {

	cpNodeCount, err := r.getNumberOfControlPlaneNodes(ctx, cluster)

	if err != nil {
		return -1, errors.Errorf("Failed to compute number of vsphere csi deployment replicas: '%v'", err)
	}

	return r.constrainNumberOfDeploymentReplicas(ctx, cpNodeCount), nil
}

func (r *VSphereCSIConfigReconciler) getNumberOfControlPlaneNodes(ctx context.Context,
	cluster *clusterapiv1beta1.Cluster) (int32, error) {

	name := cluster.Spec.ControlPlaneRef.Name
	namespace := cluster.Spec.ControlPlaneRef.Namespace

	kcp := &capictrlpkubeadmv1beta1.KubeadmControlPlane{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: namespace,
		Name:      name,
	}, kcp); err != nil {
		return -1, errors.Errorf("KubeadmControlPlane %s/%s could not be fetched, error %v",
			name, namespace, err)
	}
	return *kcp.Spec.Replicas, nil
}
