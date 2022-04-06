// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"

  cutil "github.com/vmware-tanzu/tanzu-framework/addons/controllers/utils"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	pkgtypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
	cpiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cpi/v1alpha1"
)

// mapCPIConfigToDataValuesNonParavirtual generates CPI data values for non-paravirtual modes
func (r *VSphereCPIConfigReconciler) mapCPIConfigToDataValuesNonParavirtual( // nolint
	ctx context.Context,
	cpiConfig *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (*VSphereCPIDataValues, error,
) { // nolint:whitespace
	d := &VSphereCPIDataValues{}
	c := cpiConfig.Spec.VSphereCPI.NonParavirtualConfig
	d.VSphereCPI.Mode = VsphereCPINonParavirtualMode

	// get the vsphere cluster object
	vsphereCluster, err := cutil.GetVSphereCluster(ctx, r.Client, cluster)
	if err != nil {
		return nil, err
	}

	// derive the thumbprint, server from the vsphere cluster object
	d.VSphereCPI.TLSThumbprint = vsphereCluster.Spec.Thumbprint
	d.VSphereCPI.Server = vsphereCluster.Spec.Server

	// derive vSphere username and password from the <cluster name> secret
	clusterSecret, err := cutil.GetSecret(ctx, r.Client, cluster.Namespace, cluster.Name)
	if err != nil {
		return nil, err
	}
	d.VSphereCPI.Username, d.VSphereCPI.Password, err = cutil.GetUsernameAndPasswordFromSecret(clusterSecret)
	if err != nil {
		return nil, err
	}

	// get the control plane machine template
	cpMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
	if err := r.Client.Get(ctx, types.NamespacedName{
		Namespace: cluster.Namespace,
		Name:      cutil.ControlPlaneName(cluster.Name),
	}, cpMachineTemplate); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("VSphereMachineTemplate %s/%s not found", cluster.Namespace,  cutil.ControlPlaneName(cluster.Name))
		}
		return nil, errors.Errorf("VSphereMachineTemplate %s/%s could not be fetched, error %v", cluster.Namespace, controlPlaneName(cluster.Name), err)
	}

	// derive data center information from control plane machine template, if not provided
	d.VSphereCPI.Datacenter = cpMachineTemplate.Spec.Template.Spec.Datacenter

	// derive ClusterCidr from cluster.spec.clusterNetwork
	if cluster.Spec.ClusterNetwork != nil && cluster.Spec.ClusterNetwork.Pods != nil && len(cluster.Spec.ClusterNetwork.Pods.CIDRBlocks) > 0 {
		d.VSphereCPI.Nsxt.Routes.ClusterCidr = cluster.Spec.ClusterNetwork.Pods.CIDRBlocks[0]
	}

	// derive IP family or proxy related settings from cluster annotations
	if cluster.Annotations != nil {
		d.VSphereCPI.IPFamily = cluster.Annotations[pkgtypes.IPFamilyConfigAnnotation]
		d.VSphereCPI.HTTPProxy = cluster.Annotations[pkgtypes.HTTPProxyConfigAnnotation]
		d.VSphereCPI.HTTPSProxy = cluster.Annotations[pkgtypes.HTTPSProxyConfigAnnotation]
		d.VSphereCPI.NoProxy = cluster.Annotations[pkgtypes.NoProxyConfigAnnotation]
	}

	// derive nsxt related configs from cluster variable
	d.VSphereCPI.Nsxt.PodRoutingEnabled = r.tryParseClusterVariableBool(cluster, NsxtPodRoutingEnabledVarName)
	d.VSphereCPI.Nsxt.Routes.RouterPath = r.tryParseClusterVariableString(cluster, NsxtRouterPathVarName)
	d.VSphereCPI.Nsxt.Routes.ClusterCidr = r.tryParseClusterVariableString(cluster, ClusterCIDRVarName)
	d.VSphereCPI.Nsxt.Username = r.tryParseClusterVariableString(cluster, NsxtUsernameVarName)
	d.VSphereCPI.Nsxt.Password = r.tryParseClusterVariableString(cluster, NsxtPasswordVarName)
	d.VSphereCPI.Nsxt.Host = r.tryParseClusterVariableString(cluster, NsxtManagerHostVarName)
	d.VSphereCPI.Nsxt.InsecureFlag = r.tryParseClusterVariableBool(cluster, NsxtAllowUnverifiedSSLVarName)
	d.VSphereCPI.Nsxt.RemoteAuth = r.tryParseClusterVariableBool(cluster, NsxtRemoteAuthVarName)
	d.VSphereCPI.Nsxt.VmcAccessToken = r.tryParseClusterVariableString(cluster, NsxtVmcAccessTokenVarName)
	d.VSphereCPI.Nsxt.VmcAuthHost = r.tryParseClusterVariableString(cluster, NsxtVmcAuthHostVarName)
	d.VSphereCPI.Nsxt.ClientCertKeyData = r.tryParseClusterVariableString(cluster, NsxtClientCertKeyDataVarName)
	d.VSphereCPI.Nsxt.ClientCertData = r.tryParseClusterVariableString(cluster, NsxtClientCertDataVarName)
	d.VSphereCPI.Nsxt.RootCAData = r.tryParseClusterVariableString(cluster, NsxtRootCADataB64VarName)
	d.VSphereCPI.Nsxt.SecretName = r.tryParseClusterVariableString(cluster, NsxtSecretNameVarName)
	d.VSphereCPI.Nsxt.SecretNamespace = r.tryParseClusterVariableString(cluster, NsxtSecretNamespaceVarName)

	// allow API user to override the derived values if he/she specified fields in the VSphereCPIConfig
	if c.TLSThumbprint != "" {
		d.VSphereCPI.TLSThumbprint = c.TLSThumbprint
	}
	if c.Server != "" {
		d.VSphereCPI.Server = c.Server
	}
	if c.Datacenter != "" {
		d.VSphereCPI.Datacenter = c.Datacenter
	}
	if c.VSphereCredentialRef != nil {
		vsphereSecret, err := cutil.GetSecret(ctx,r.Client, c.VSphereCredentialRef.Namespace, c.VSphereCredentialRef.Name)
		if err != nil {
			return nil, err
		}
		d.VSphereCPI.Username, d.VSphereCPI.Password, err = cutil.GetUsernameAndPasswordFromSecret(vsphereSecret)
		if err != nil {
			return nil, err
		}
	}
	d.VSphereCPI.Region = c.Region
	d.VSphereCPI.Zone = c.Zone
	d.VSphereCPI.InsecureFlag = c.InsecureFlag
	d.VSphereCPI.VMInternalNetwork = c.VMInternalNetwork
	d.VSphereCPI.VMExternalNetwork = c.VMExternalNetwork
	d.VSphereCPI.VMExcludeInternalNetworkSubnetCidr = c.VMExcludeInternalNetworkSubnetCidr
	d.VSphereCPI.VMExcludeExternalNetworkSubnetCidr = c.VMExcludeExternalNetworkSubnetCidr
	d.VSphereCPI.CloudProviderExtraArgs.TLSCipherSuites = c.TLSCipherSuites

	if c.NSXT != nil {
		if c.NSXT.PodRoutingEnabled {
			d.VSphereCPI.Nsxt.PodRoutingEnabled = c.NSXT.PodRoutingEnabled
		}

		if c.NSXT.InsecureFlag {
			d.VSphereCPI.Nsxt.InsecureFlag = c.NSXT.InsecureFlag
		}
		if c.NSXT.Routes != nil {
			d.VSphereCPI.Nsxt.Routes.RouterPath = c.NSXT.Routes.RouterPath
			d.VSphereCPI.Nsxt.Routes.ClusterCidr = c.NSXT.Routes.ClusterCidr
		}
		if c.NSXT.NSXTCredentialsRef != nil {
			nsxtSecret, err := cutil.GetSecret(ctx, r.Client, c.NSXT.NSXTCredentialsRef.Namespace, c.NSXT.NSXTCredentialsRef.Name)
			if err != nil {
				return nil, err
			}
			d.VSphereCPI.Nsxt.Username, d.VSphereCPI.Nsxt.Password, err = cutil.GetUsernameAndPasswordFromSecret(nsxtSecret)
			if err != nil {
				return nil, err
			}
		}
		if c.NSXT.Host != "" {
			d.VSphereCPI.Nsxt.Host = c.NSXT.Host
		}
		if c.NSXT.RemoteAuth {
			d.VSphereCPI.Nsxt.RemoteAuth = c.NSXT.RemoteAuth
		}
		if c.NSXT.VMCAccessToken != "" {
			d.VSphereCPI.Nsxt.VmcAccessToken = c.NSXT.VMCAccessToken
		}
		if c.NSXT.VMCAuthHost != "" {
			d.VSphereCPI.Nsxt.VmcAuthHost = c.NSXT.VMCAuthHost
		}
		if c.NSXT.ClientCertKeyData != "" {
			d.VSphereCPI.Nsxt.ClientCertKeyData = c.NSXT.ClientCertKeyData
		}
		if c.NSXT.ClientCertData != "" {
			d.VSphereCPI.Nsxt.ClientCertData = c.NSXT.ClientCertData
		}
		if c.NSXT.RootCAData != "" {
			d.VSphereCPI.Nsxt.RootCAData = c.NSXT.RootCAData
		}
		if c.NSXT.SecretName != "" {
			d.VSphereCPI.Nsxt.SecretName = c.NSXT.SecretName
			d.VSphereCPI.Nsxt.SecretNamespace = c.NSXT.SecretNamespace
		}
	}
	if c.IPFamily != "" {
		d.VSphereCPI.IPFamily = c.IPFamily
	}
	if c.HTTPProxy != "" {
		d.VSphereCPI.HTTPProxy = c.HTTPProxy
	}
	if c.HTTPSProxy != "" {
		d.VSphereCPI.HTTPSProxy = c.HTTPSProxy
	}
	if c.NoProxy != "" {
		d.VSphereCPI.NoProxy = c.NoProxy
	}
	return d, nil
}

// mapCPIConfigToDataValuesParavirtual generates CPI data values for paravirtual modes
func (r *VSphereCPIConfigReconciler) mapCPIConfigToDataValuesParavirtual(_ context.Context, _ *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (*VSphereCPIDataValues, error) {
	d := &VSphereCPIDataValues{}
	d.VSphereCPI.Mode = VSphereCPIParavirtualMode

	// derive owner cluster information
	d.VSphereCPI.ClusterAPIVersion = cluster.GroupVersionKind().GroupVersion().String()
	d.VSphereCPI.ClusterKind = cluster.GroupVersionKind().Kind
	d.VSphereCPI.ClusterName = cluster.ObjectMeta.Name
	d.VSphereCPI.ClusterUID = string(cluster.ObjectMeta.UID)

	d.VSphereCPI.SupervisorMasterEndpointIP = SupervisorEndpointHostname
	d.VSphereCPI.SupervisorMasterPort = fmt.Sprint(SupervisorEndpointPort)

	return d, nil
}

// mapCPIConfigToDataValues maps VSphereCPIConfig CR to data values
func (r *VSphereCPIConfigReconciler) mapCPIConfigToDataValues(ctx context.Context, cpiConfig *cpiv1alpha1.VSphereCPIConfig, cluster *clusterapiv1beta1.Cluster) (*VSphereCPIDataValues, error) {
	switch cpiConfig.Spec.VSphereCPI.Mode {
	case VsphereCPINonParavirtualMode:
		return r.mapCPIConfigToDataValuesNonParavirtual(ctx, cpiConfig, cluster)
	case VSphereCPIParavirtualMode:
		return r.mapCPIConfigToDataValuesParavirtual(ctx, cpiConfig, cluster)
	default:
		break
	}
	return nil, errors.Errorf("Invalid CPI mode %s, must either be %s or %s", cpiConfig.Spec.VSphereCPI.Mode, VSphereCPIParavirtualMode, VsphereCPINonParavirtualMode)
}

// mapCPIConfigToProviderServiceAccountSpec maps CPIConfig and cluster to the corresponding service account spec
func (r *VSphereCPIConfigReconciler) mapCPIConfigToProviderServiceAccountSpec(cluster *clusterapiv1beta1.Cluster) capvvmwarev1beta1.ProviderServiceAccountSpec {
	return capvvmwarev1beta1.ProviderServiceAccountSpec{
		Ref: &v1.ObjectReference{Name: cluster.Name, Namespace: cluster.Namespace},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:     []string{"get", "create", "update", "patch", "delete"},
				APIGroups: []string{"vmoperator.vmware.com"},
				Resources: []string{"virtualmachineservices", "virtualmachineservices/status"},
			},
			{
				Verbs:     []string{"get", "list"},
				APIGroups: []string{"vmoperator.vmware.com"},
				Resources: []string{"virtualmachines", "virtualmachines/status"},
			},
			{
				Verbs:     []string{"get", "create", "update", "list", "patch", "delete", "watch"},
				APIGroups: []string{"nsx.vmware.com"},
				Resources: []string{"ippools", "ippools/status"},
			},
			{
				Verbs:     []string{"get", "create", "update", "list", "patch", "delete"},
				APIGroups: []string{"nsx.vmware.com"},
				Resources: []string{"routesets", "routesets/status"},
			},
		},
		TargetNamespace:  ProviderServiceAccountSecretNamespace,
		TargetSecretName: ProviderServiceAccountSecretName,
	}
}

// getOwnerCluster verifies that the VSphereCPIConfig has a cluster as its owner reference,
// and returns the cluster. It tries to read the cluster name from the VSphereCPIConfig's owner reference objects.
// If not there, we assume the owner cluster and VSphereCPIConfig always has the same name.
func (r *VSphereCPIConfigReconciler) getOwnerCluster(ctx context.Context, cpiConfig *cpiv1alpha1.VSphereCPIConfig) (*clusterapiv1beta1.Cluster, error) {
	cluster := &clusterapiv1beta1.Cluster{}
	clusterName := cpiConfig.Name

	// retrieve the owner cluster for the VSphereCPIConfig object
	for _, ownerRef := range cpiConfig.GetOwnerReferences() {
		if strings.EqualFold(ownerRef.Kind, constants.ClusterKind) {
			clusterName = ownerRef.Name
			break
		}
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: cpiConfig.Namespace, Name: clusterName}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			r.Log.Info(fmt.Sprintf("Cluster resource '%s/%s' not found", cpiConfig.Namespace, clusterName))
			return nil, nil
		}
		r.Log.Error(err, fmt.Sprintf("Unable to fetch cluster '%s/%s'", cpiConfig.Namespace, clusterName))
		return nil, err
	}
	r.Log.Info(fmt.Sprintf("Cluster resource '%s/%s' is successfully found", cpiConfig.Namespace, clusterName))
	return cluster, nil
}

// tryParseClusterVariableBool tries to parse a boolean cluster variable,
// info any error that occurs
func (r *VSphereCPIConfigReconciler) tryParseClusterVariableBool(cluster *clusterapiv1beta1.Cluster, variableName string) bool {
	res, err := util.ParseClusterVariableBool(cluster, variableName)
	if err != nil {
		r.Log.Info(fmt.Sprintf("Cannot parse cluster variable with key %s", variableName))
	}
	return res
}

// tryParseClusterVariableString tries to parse a string cluster variable,
// info any error that occurs
func (r *VSphereCPIConfigReconciler) tryParseClusterVariableString(cluster *clusterapiv1beta1.Cluster, variableName string) string {
	res, err := util.ParseClusterVariableString(cluster, variableName)
	if err != nil {
		r.Log.Info(fmt.Sprintf("cannot parse cluster variable with key %s", variableName))
	}
	return res
}
