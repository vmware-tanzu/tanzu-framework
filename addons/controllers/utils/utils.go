// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package utils contains utility functions shared by controllers
package utils

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	kubeadmv1beta1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
)

// VSphereClusterParavirtualForCAPICluster gets the VSphereCluster CR for the cluster object in paravirtual mode
func VSphereClusterParavirtualForCAPICluster(ctx context.Context, clt client.Client, cluster *clusterapiv1beta1.Cluster) (*capvvmwarev1beta1.VSphereCluster, error) {
	vSphereClusterRef := cluster.Spec.InfrastructureRef
	if vSphereClusterRef == nil {
		return nil, fmt.Errorf("cluster %s 's infrastructure reference is not set yet", cluster.Name)
	}
	if vSphereClusterRef.Kind != constants.InfrastructureRefVSphere || vSphereClusterRef.APIVersion != capvvmwarev1beta1.GroupVersion.String() {
		return nil, fmt.Errorf("cluster %s 's infrastructure reference is not a paravirt vSphereCluster: %v", cluster.Name, vSphereClusterRef)
	}
	vsphereCluster := &capvvmwarev1beta1.VSphereCluster{}
	if err := clt.Get(ctx, types.NamespacedName{Namespace: vSphereClusterRef.Namespace, Name: vSphereClusterRef.Name}, vsphereCluster); err != nil {
		return nil, err
	}
	return vsphereCluster, nil
}

// VSphereClusterNonParavirtualForCluster gets the VSphereCluster CR for the cluster object in non-paravirtual mode
func VSphereClusterNonParavirtualForCluster(ctx context.Context, clt client.Client, cluster *clusterapiv1beta1.Cluster) (*capvv1beta1.VSphereCluster, error) {
	vSphereClusterRef := cluster.Spec.InfrastructureRef
	if vSphereClusterRef == nil {
		return nil, fmt.Errorf("cluster %s 's infrastructure reference is not set yet", cluster.Name)
	}
	vsphereCluster := &capvv1beta1.VSphereCluster{}
	if vSphereClusterRef.Kind != constants.InfrastructureRefVSphere || vSphereClusterRef.APIVersion != capvv1beta1.GroupVersion.String() {
		return nil, fmt.Errorf("cluster %s 's infrastructure reference is not a non-paravirt vSphereCluster: %v", cluster.Name, vSphereClusterRef)
	}
	if err := clt.Get(ctx, types.NamespacedName{Namespace: vSphereClusterRef.Namespace, Name: vSphereClusterRef.Name}, vsphereCluster); err != nil {
		return nil, err
	}
	return vsphereCluster, nil
}

// ControlPlaneVsphereMachineTemplateNonParavirtualForCluster gets the VsphereMachineTemplate CR of control plane
func ControlPlaneVsphereMachineTemplateNonParavirtualForCluster(ctx context.Context, clt client.Client, cluster *clusterapiv1beta1.Cluster) (*capvv1beta1.VSphereMachineTemplate, error) {
	controlPlaneRef := cluster.Spec.ControlPlaneRef
	if controlPlaneRef == nil {
		return nil, fmt.Errorf("cluster %s 's controlplane reference is not set yet", cluster.Name)
	}
	controlPlane := &kubeadmv1beta1.KubeadmControlPlane{}
	if err := clt.Get(ctx, types.NamespacedName{Namespace: controlPlaneRef.Namespace, Name: controlPlaneRef.Name}, controlPlane); err != nil {
		return nil, err
	}
	vSphereMachineTemplateRef := controlPlane.Spec.MachineTemplate.InfrastructureRef
	vSphereMachineTemplate := &capvv1beta1.VSphereMachineTemplate{}
	if vSphereMachineTemplateRef.Kind != constants.InfrastructureRefVSphereMachineTemplate || vSphereMachineTemplateRef.APIVersion != capvv1beta1.GroupVersion.String() {
		return nil, fmt.Errorf("cluster %s 's controlplane machinetemplate reference is not a non-paravirt vSphereMachineTemplate: %v", cluster.Name, vSphereMachineTemplateRef)
	}
	if err := clt.Get(ctx, types.NamespacedName{Namespace: vSphereMachineTemplateRef.Namespace, Name: vSphereMachineTemplateRef.Name}, vSphereMachineTemplate); err != nil {
		return nil, err
	}
	return vSphereMachineTemplate, nil
}

// ControlPlaneName returns the control plane name for a cluster name
func ControlPlaneName(clusterName string) string {
	return fmt.Sprintf("%s-control-plane", clusterName)
}

// GetSecret gets the secret object given its name and namespace
func GetSecret(ctx context.Context, clt client.Client, namespace, name string) (*v1.Secret, error) {
	secret := &v1.Secret{}
	if err := clt.Get(ctx, types.NamespacedName{Namespace: namespace, Name: name}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, errors.Errorf("Secret %s/%s not found", namespace, name)
		}
		return nil, errors.Errorf("Secret %s/%s could not be fetched, error %v", namespace, name, err)
	}
	return secret, nil
}

// GetUsernameAndPasswordFromSecret extracts a username and password from secret
func GetUsernameAndPasswordFromSecret(s *v1.Secret) (string, string, error) {
	username, exists := s.Data["username"]
	if !exists {
		return "", "", errors.Errorf("Secret %s/%s doesn't have string data with username", s.Namespace, s.Name)
	}
	password, exists := s.Data["password"]
	if !exists {
		return "", "", errors.Errorf("Secret %s/%s doesn't have string data with password", s.Namespace, s.Name)
	}
	return string(username), string(password), nil
}

// GetCCMName returns the name of cloud control manager for a cluster
func GetCCMName(cluster *clusterapiv1beta1.Cluster) string {
	return fmt.Sprintf("%s-%s", cluster.Name, "ccm")
}

// TryParseClusterVariableBool tries to parse a boolean cluster variable,
// info any error that occurs
func TryParseClusterVariableBool(ctx context.Context, cluster *clusterapiv1beta1.Cluster,
	variableName string) bool {

	res, err := util.ParseClusterVariableBool(cluster, variableName)
	if err != nil {
		log.FromContext(ctx).Error(err, "cannot parse cluster variable", "key", variableName)
	}
	return res
}

// TryParseClusterVariableString tries to parse a string cluster variable,
// info any error that occurs
func TryParseClusterVariableString(ctx context.Context, cluster *clusterapiv1beta1.Cluster,
	variableName string) string {

	res, err := util.ParseClusterVariableString(cluster, variableName)
	if err != nil {
		log.FromContext(ctx).Error(err, "cannot parse cluster variable", "key", variableName)
	}
	return res
}
