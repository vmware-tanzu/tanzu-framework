// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package utils contains utilities used throughout the codebase.
package utils

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// IsClusterBootstrapType returns true if the secret is type `tkg.tanzu.vmware.com/addon`
func IsClusterBootstrapType(secret *corev1.Secret) bool {
	return secret.Type == constants.ClusterBootstrapManagedSecret
}

// ContainsPackageName returns true if the `tkg.tanzu.vmware.com/package-name` label contains the package name we pass in
func ContainsPackageName(secret *corev1.Secret, expectedName string) bool {
	name, labelExists := secret.Labels[constants.PackageNameLabel]
	return labelExists && strings.Contains(name, expectedName)
}

func GetClusterNameFromSecret(secret *corev1.Secret) (string, error) {
	for _, ownerRef := range secret.OwnerReferences {
		if ownerRef.Kind == reflect.TypeOf(clusterapiv1beta1.Cluster{}).Name() {
			return ownerRef.Name, nil
		}
	}

	if secret.GetLabels() != nil {
		clusterName := secret.GetLabels()[constants.TKGClusterNameLabel]
		if clusterName != "" {
			return clusterName, nil
		}
	}
	return "", fmt.Errorf("could not get cluster name from secret")
}

// IsManagementCluster returns true if the cluster has the "cluster-role.tkg.tanzu.vmware.com/management" label
func IsManagementCluster(cluster *clusterapiv1beta1.Cluster) bool {
	_, labelExists := cluster.GetLabels()[constants.TKGManagementLabel]
	return labelExists
}

// GetInfraProvider get infrastructure kind from cluster spec
func GetInfraProvider(cluster *clusterapiv1beta1.Cluster) (string, error) {
	var infraProvider string

	infrastructureRef := cluster.Spec.InfrastructureRef
	if infrastructureRef == nil {
		return "", fmt.Errorf("cluster.Spec.InfrastructureRef is not set for cluster '%s", cluster.Name)
	}

	switch infrastructureRef.Kind {
	case tkgconstants.InfrastructureRefVSphere:
		infraProvider = tkgconstants.InfrastructureProviderVSphere
	case tkgconstants.InfrastructureRefAWS:
		infraProvider = tkgconstants.InfrastructureProviderAWS
	case tkgconstants.InfrastructureRefAzure:
		infraProvider = tkgconstants.InfrastructureProviderAzure
	case constants.InfrastructureRefDocker:
		infraProvider = tkgconstants.InfrastructureProviderDocker
	default:
		return "", fmt.Errorf("unknown cluster.Spec.InfrastructureRef.Kind is set for cluster '%s", cluster.Name)
	}

	return infraProvider, nil
}

func ListSecretsContainingPackageName(ctx context.Context, c client.Client, packageName string) (*corev1.SecretList, error) {
	secrets := &corev1.SecretList{}
	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: constants.PackageNameLabel, Operator: metav1.LabelSelectorOpIn, Values: []string{packageName}},
		},
	}

	s, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return nil, err
	}

	if err := c.List(ctx, secrets, client.MatchingLabelsSelector{Selector: s}); err != nil {
		return nil, err
	}

	return secrets, nil
}

func GetPinnipedInfoConfigMap(ctx context.Context, c client.Client, log logr.Logger) (*corev1.ConfigMap, error) {
	pinnipedInfoCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.KubePublicNamespace,
			Name:      constants.PinnipedInfoConfigMapName,
		},
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(pinnipedInfoCM), pinnipedInfoCM); err != nil {
		if !k8serror.IsNotFound(err) {
			return nil, err
		}

		log.V(1).Info("pinniped-info configmap not found, setting value to nil")
		pinnipedInfoCM.Data = nil
	}

	return pinnipedInfoCM, nil
}

func GetClusterFromSecret(ctx context.Context, c client.Client, secret *corev1.Secret) (*clusterapiv1beta1.Cluster, error) {
	secretNamespace := secret.Namespace
	clusterName, err := GetClusterNameFromSecret(secret)
	if err != nil {
		return nil, err
	}

	cluster := &clusterapiv1beta1.Cluster{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: clusterName}, cluster); err != nil {
		return nil, err
	}

	return cluster, nil
}
