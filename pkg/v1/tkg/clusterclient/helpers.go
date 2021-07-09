// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package clusterclient

import (
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

var (
	controlPlaneTolerations = []corev1.Toleration{
		{
			Effect: "NoSchedule",
			Key:    "node-role.kubernetes.io/control-plane",
		},
		{
			Effect: "NoSchedule",
			Key:    "node-role.kubernetes.io/master",
		},
	}

	controlPlaneNodeSelectors = []corev1.NodeSelectorTerm{
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      "node-role.kubernetes.io/control-plane",
					Operator: "Exists",
				},
			},
		},
		{
			MatchExpressions: []corev1.NodeSelectorRequirement{
				{
					Key:      "node-role.kubernetes.io/master",
					Operator: "Exists",
				},
			},
		},
	}
)

// yamlToUnstructured reads yaml bytes and converts it to *unstructured.Unstructured.
func yamlToUnstructured(rawYAML []byte) (*unstructured.Unstructured, error) {
	unst := &unstructured.Unstructured{}
	err := yaml.Unmarshal(rawYAML, unst)
	return unst, err
}

func updateFieldInUnstructured(configuration *unstructured.Unstructured, path []string, value string) error {
	currentValue, _, err := unstructured.NestedString(configuration.UnstructuredContent(), path...)
	if err != nil {
		return errors.Wrapf(err, "unable to retrieve %q from unstructured configuration", strings.Join(path, "."))
	}
	if currentValue != value {
		if err := unstructured.SetNestedField(configuration.UnstructuredContent(), value, path...); err != nil {
			return errors.Wrapf(err, "unable to update %q on unstructured configuration", strings.Join(path, "."))
		}
	}
	return nil
}

// UpdateCoreDNSImageRepositoryInKubeadmConfigMap updates coredns imageRepository in kubeadm-config configMap
func UpdateCoreDNSImageRepositoryInKubeadmConfigMap(kubedmconfigmap *corev1.ConfigMap, newImageRepository string) error {
	data, ok := kubedmconfigmap.Data[clusterConfigurationKey]
	if !ok {
		return errors.Errorf("unable to find %q key in kubeadm ConfigMap", clusterConfigurationKey)
	}

	configuration, err := yamlToUnstructured([]byte(data))
	if err != nil {
		return errors.Wrapf(err, "unable to decode kubeadm ConfigMap's %q to Unstructured object", clusterConfigurationKey)
	}

	// Update dns.imageRepository in kubeadm-config ConfigMap
	err = updateFieldInUnstructured(configuration, []string{"dns", "imageRepository"}, newImageRepository)
	if err != nil {
		return errors.Wrap(err, "unable to update kubeadm-config ConfigMap")
	}

	updated, err := yaml.Marshal(configuration)
	if err != nil {
		return errors.Wrapf(err, "unable to encode kubeadm ConfigMap's %q to YAML", clusterConfigurationKey)
	}
	kubedmconfigmap.Data[clusterConfigurationKey] = string(updated)
	return nil
}

// ensurePodSpecControlPlaneAffinity sets tolerations and affinity on a pod spec
// to ensure that it runs on a control plane node.
func ensurePodSpecControlPlaneAffinity(spec *corev1.PodSpec) {
	spec.Tolerations = ensureControlPlaneTolerations(spec.Tolerations)
	spec.Affinity = &corev1.Affinity{
		NodeAffinity: &corev1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
				NodeSelectorTerms: controlPlaneNodeSelectors,
			},
		},
	}
}

// ensureControlPlaneTolerations sets the tolerations to allow a pod to run
// on a control plane node
func ensureControlPlaneTolerations(tolerations []corev1.Toleration) []corev1.Toleration {
	for _, cpToleration := range controlPlaneTolerations {
		if !hasToleration(cpToleration, tolerations) {
			tolerations = append(tolerations, cpToleration)
		}
	}
	return tolerations
}

// hasToleration returns true if a slice of tolerations includes a given toleration
func hasToleration(toleration corev1.Toleration, tolerations []corev1.Toleration) bool {
	for _, t := range tolerations {
		if t == toleration {
			return true
		}
	}
	return false
}
