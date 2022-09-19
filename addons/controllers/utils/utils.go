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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	capvv1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	capvvmwarev1beta1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/vmware/v1beta1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/util"
)

// VSphereClusterParavirtualForCAPICluster gets the VSphereCluster CR for the cluster object in paravirtual mode
func VSphereClusterParavirtualForCAPICluster(ctx context.Context, clt client.Client, cluster *clusterapiv1beta1.Cluster) (*capvvmwarev1beta1.VSphereCluster, error) {

	vsphereClusters := &capvvmwarev1beta1.VSphereClusterList{}
	labelMatch, err := labels.NewRequirement(clusterapiv1beta1.ClusterLabelName, selection.Equals, []string{cluster.Name})
	if err != nil {
		return nil, err
	}
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*labelMatch)
	if err := clt.List(ctx, vsphereClusters, &client.ListOptions{LabelSelector: labelSelector, Namespace: cluster.Namespace}); err != nil {
		return nil, err
	}
	if len(vsphereClusters.Items) != 1 {
		return nil, fmt.Errorf("expected to find 1 VSphereCluster object for label key %s and value %s but found %d",
			clusterapiv1beta1.ClusterLabelName, cluster.Name, len(vsphereClusters.Items))
	}
	return &vsphereClusters.Items[0], nil
}

// VSphereClusterNonParavirtualForCluster gets the VSphereCluster CR for the cluster object in non-paravirtual mode
func VSphereClusterNonParavirtualForCluster(ctx context.Context, clt client.Client, cluster *clusterapiv1beta1.Cluster) (*capvv1beta1.VSphereCluster, error) {

	vsphereClusters := &capvv1beta1.VSphereClusterList{}
	labelMatch, err := labels.NewRequirement(clusterapiv1beta1.ClusterLabelName, selection.Equals, []string{cluster.Name})
	if err != nil {
		return nil, err
	}
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*labelMatch)
	if err := clt.List(ctx, vsphereClusters, &client.ListOptions{LabelSelector: labelSelector, Namespace: cluster.Namespace}); err != nil {
		return nil, err
	}
	if len(vsphereClusters.Items) != 1 {
		return nil, fmt.Errorf("expected to find 1 VSphereCluster object for label key %s and value %s but found %d",
			clusterapiv1beta1.ClusterLabelName, cluster.Name, len(vsphereClusters.Items))
	}
	return &vsphereClusters.Items[0], nil
}

// ControlPlaneVsphereMachineTemplateForCluster gets the VsphereMachineTemplate CR of control plane
func ControlPlaneVsphereMachineTemplateForCluster(ctx context.Context, clt client.Client, cluster *clusterapiv1beta1.Cluster) (*capvv1beta1.VSphereMachineTemplate, error) {

	vSphereMachineTemplates := &capvv1beta1.VSphereMachineTemplateList{}
	labelMatch, err := labels.NewRequirement(clusterapiv1beta1.ClusterLabelName, selection.Equals, []string{cluster.Name})
	if err != nil {
		return nil, err
	}
	labelNotMatch, err := labels.NewRequirement(clusterapiv1beta1.ClusterTopologyMachineDeploymentLabelName, selection.DoesNotExist, nil)
	if err != nil {
		return nil, err
	}
	labelSelector := labels.NewSelector()
	labelSelector = labelSelector.Add(*labelMatch, *labelNotMatch)
	if err := clt.List(ctx, vSphereMachineTemplates, &client.ListOptions{LabelSelector: labelSelector, Namespace: cluster.Namespace}); err != nil {
		return nil, err
	}
	if len(vSphereMachineTemplates.Items) != 1 {
		return nil, fmt.Errorf("expected to find 1 vSphereMachineTemplate object with requirements %s in namespace %s but found %d",
			labelSelector, cluster.Namespace, len(vSphereMachineTemplates.Items))
	}
	return &vSphereMachineTemplates.Items[0], nil
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
