// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-logr/logr"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
)

// secretIsType returns true if the secret type matches the given expectedType
func secretIsType(secret *corev1.Secret, expectedType corev1.SecretType) bool {
	return secret.Type == expectedType
}

// containsPackageName returns true if the `tkg.tanzu.vmware.com/package-name` label contains the package name we pass in
func containsPackageName(secret *corev1.Secret, expectedName string) bool {
	name, labelExists := secret.Labels[packageNameLabel]
	return labelExists && strings.Contains(name, expectedName)
}

// hasLabel returns true if the given object has the provided label
func hasLabel(o client.Object, label string) bool {
	_, labelExists := o.GetLabels()[label]

	return labelExists
}

// matchesLabelValue returns true if the value for the given labelKey matches the labelValue we provide
func matchesLabelValue(secret *corev1.Secret, labelKey, labelValue string) bool {
	if !hasLabel(secret, labelKey) {
		return false
	}
	return secret.Labels[labelKey] == labelValue
}

// getClusterNameFromSecret gets the cluster name from data in a secret and returns the cluster name
func getClusterNameFromSecret(secret *corev1.Secret) (string, error) {
	for _, ownerRef := range secret.OwnerReferences {
		if ownerRef.Kind == reflect.TypeOf(clusterapiv1beta1.Cluster{}).Name() {
			return ownerRef.Name, nil
		}
	}

	if secret.GetLabels() != nil {
		clusterName := secret.GetLabels()[tkgClusterNameLabel]
		if clusterName != "" {
			return clusterName, nil
		}
	}
	return "", fmt.Errorf("could not get cluster name from secret")
}

// isManagementCluster returns true if the cluster has the "cluster-role.tkg.tanzu.vmware.com/management" label
func isManagementCluster(cluster *clusterapiv1beta1.Cluster) bool {
	_, labelExists := cluster.GetLabels()[tkgManagementLabel]
	return labelExists
}

// getInfraProvider get infrastructure kind from cluster spec
func getInfraProvider(cluster *clusterapiv1beta1.Cluster) (string, error) {
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
	case infrastructureRefDocker:
		infraProvider = tkgconstants.InfrastructureProviderDocker
	default:
		return "", fmt.Errorf("unknown cluster.Spec.InfrastructureRef.Kind is set for cluster '%s", cluster.Name)
	}

	return infraProvider, nil
}

// listSecretsContainingPackageName returns true if the value of the "tkg.tanzu.vmware.com/package-name" label includes the given packageName
func listSecretsContainingPackageName(ctx context.Context, c client.Client, packageName string) (*corev1.SecretList, error) {
	packageSecrets := &corev1.SecretList{}
	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: packageNameLabel, Operator: metav1.LabelSelectorOpExists, Values: nil},
		},
	}

	s, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return nil, err
	}
	if err := c.List(ctx, packageSecrets, client.MatchingLabelsSelector{Selector: s}); err != nil {
		return nil, err
	}
	secrets := &corev1.SecretList{}
	// Unfortunately I could not find a LabelSelector that would do value "contains" for us, manually doing this for now
	for i := range packageSecrets.Items {
		secret := &packageSecrets.Items[i]
		if secretIsType(secret, clusterBootstrapManagedSecret) && containsPackageName(secret, packageName) {
			secrets.Items = append(secrets.Items, *secret)
		}
	}

	return secrets, nil
}

// listClustersContainingLabel returns true if the cluster has the given label
func listClustersContainingLabel(ctx context.Context, c client.Client, label string) (*clusterapiv1beta1.ClusterList, error) {
	clusters := &clusterapiv1beta1.ClusterList{}
	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: label, Operator: metav1.LabelSelectorOpExists, Values: nil},
		},
	}

	s, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return nil, err
	}

	if err := c.List(ctx, clusters, client.MatchingLabelsSelector{Selector: s}); err != nil {
		return nil, err
	}

	return clusters, nil
}

// getPinnipedInfoConfigMap returns the "pinniped-info" configmap in the "kube-public" namespace, sets data to nil if configmap not found
func getPinnipedInfoConfigMap(ctx context.Context, c client.Client, log logr.Logger) (*corev1.ConfigMap, error) {
	pinnipedInfoCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: kubePublicNamespace,
			Name:      pinnipedInfoConfigMapName,
		},
	}

	if err := c.Get(ctx, client.ObjectKeyFromObject(pinnipedInfoCM), pinnipedInfoCM); err != nil {
		if !k8serror.IsNotFound(err) {
			return nil, err
		}

		log.V(1).Info("pinniped-info configmap not found, setting Data to nil")
		pinnipedInfoCM.Data = nil
	}

	return pinnipedInfoCM, nil
}

// getClusterFromSecret returns the cluster associated with a given secret
func getClusterFromSecret(ctx context.Context, c client.Client, secret *corev1.Secret) (*clusterapiv1beta1.Cluster, error) {
	secretNamespace := secret.Namespace
	clusterName, err := getClusterNameFromSecret(secret)
	if err != nil {
		return nil, err
	}

	cluster := &clusterapiv1beta1.Cluster{}
	if err := c.Get(ctx, types.NamespacedName{Namespace: secretNamespace, Name: clusterName}, cluster); err != nil {
		return nil, err
	}

	return cluster, nil
}

// getMutateFn returns a function that updates the "values.yaml" section of the given secret
func getMutateFn(secret *corev1.Secret, pinnipedInfoCM *corev1.ConfigMap, cluster *clusterapiv1beta1.Cluster, log logr.Logger, isV1 bool) func() error {
	return func() error {
		supervisorAddress := ""
		supervisorCABundle := ""
		identityManagementType := none

		if pinnipedInfoCM.Data != nil {
			supervisorAddress = pinnipedInfoCM.Data[issuerKey]
			supervisorCABundle = pinnipedInfoCM.Data[issuerCABundleKey]
			identityManagementType = oidc
			log.V(1).Info("retrieved data from pinniped-info configmap",
				"supervisorAddress", supervisorAddress,
				"supervisorCABundle", supervisorCABundle)
		}

		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		}

		pinnipedDataValues := &pinnipedDataValues{}
		existingDataValues, labelExists := secret.Data[tkgDataValueFieldName]
		if labelExists {
			if err := yaml.Unmarshal(existingDataValues, pinnipedDataValues); err != nil {
				log.Error(err, "unable to unmarshal existing data values from secret")
			}
		}

		pinnipedDataValues.IdentityManagementType = identityManagementType
		infraProvider, err := getInfraProvider(cluster)
		if err != nil {
			if pinnipedDataValues.Infrastructure != "" {
				infraProvider = pinnipedDataValues.Infrastructure
			} else {
				log.Error(err, "unable to get infrastructure_provider, setting to vSphere")
				infraProvider = tkgconstants.InfrastructureProviderVSphere
			}
		}
		pinnipedDataValues.Infrastructure = infraProvider
		pinnipedDataValues.ClusterRole = "workload"
		pinnipedDataValues.Pinniped.SupervisorEndpoint = supervisorAddress
		pinnipedDataValues.Pinniped.SupervisorCABundle = supervisorCABundle
		// TODO: Do we want to include concierge audience if idmgmttype is none?
		pinnipedDataValues.Pinniped.Concierge.Audience = fmt.Sprintf("%s-%s", cluster.Name, string(cluster.UID))
		dataValueYamlBytes, err := yaml.Marshal(pinnipedDataValues)
		if err != nil {
			log.Error(err, "error marshaling Pinniped Secret values to yaml")
			return err
		}

		if isV1 {
			dataValueYamlBytes = append([]byte(valuesYAMLPrefix), dataValueYamlBytes...)
		}

		secret.Data[tkgDataValueFieldName] = dataValueYamlBytes
		return nil
	}
}

func secretNameFromClusterName(clusterName types.NamespacedName) types.NamespacedName {
	return types.NamespacedName{
		Namespace: clusterName.Namespace,
		Name:      fmt.Sprintf("%s-pinniped-addon", clusterName.Name),
	}
}

type pinnipedDataValues struct {
	IdentityManagementType string   `yaml:"identity_management_type,omitempty"`
	Infrastructure         string   `yaml:"infrastructure_provider,omitempty"`
	ClusterRole            string   `yaml:"tkg_cluster_role,omitempty"`
	Pinniped               pinniped `yaml:"pinniped,omitempty"`
}

type concierge struct {
	Audience string `yaml:"audience,omitempty"`
}

type pinniped struct {
	SupervisorEndpoint string    `yaml:"supervisor_svc_endpoint"`
	SupervisorCABundle string    `yaml:"supervisor_ca_bundle_data"`
	Concierge          concierge `yaml:"concierge"`
}
