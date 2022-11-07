// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	corev1 "k8s.io/api/core/v1"
	capibootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

const (
	admissionPodSecurityConfigFilePath = "/etc/kubernetes/kube-apiserver-admission-pss.yaml"
	admissionPodSecurityConfigFileData = `apiVersion: apiserver.config.k8s.io/v1
kind: AdmissionConfiguration
plugins:
- name: PodSecurity
  configuration:
    apiVersion: pod-security.admission.config.k8s.io/v1beta1
    kind: PodSecurityConfiguration
    defaults:
      enforce: ""
      enforce-version: ""
      audit: "baseline"
      audit-version: "v1.24"
      warn: "baseline"
      warn-version: "v1.24"
    exemptions:
      usernames: []
      runtimeClasses: []
      namespaces: ["kube-system","tkg-system"]
`
	admissionPodSecurityConfigFlagName = "admission-control-config-file"
)

func (c *TkgClient) configurePodSecurityStandard(old *controlplanev1.KubeadmControlPlane) *controlplanev1.KubeadmControlPlane {

	fileExists := false
	for _, f := range old.Spec.KubeadmConfigSpec.Files {
		if f.Path == admissionPodSecurityConfigFilePath {
			fileExists = true
			break
		}
	}

	volumeMountExists := false
	if old.Spec.KubeadmConfigSpec.ClusterConfiguration != nil {
		for _, m := range old.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraVolumes {
			if m.HostPath == admissionPodSecurityConfigFilePath {
				volumeMountExists = true
				break
			}
		}
	}

	if fileExists && volumeMountExists && old.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs[admissionPodSecurityConfigFlagName] == admissionPodSecurityConfigFilePath {
		return old
	}

	kcp := old.DeepCopy()

	if !fileExists {
		kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files,
			capibootstrapv1.File{
				Path:    admissionPodSecurityConfigFilePath,
				Content: admissionPodSecurityConfigFileData,
			},
		)
	}

	if !volumeMountExists {
		if kcp.Spec.KubeadmConfigSpec.ClusterConfiguration == nil {
			kcp.Spec.KubeadmConfigSpec.ClusterConfiguration = &capibootstrapv1.ClusterConfiguration{}
		}

		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraVolumes = append(kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraVolumes,
			capibootstrapv1.HostPathMount{
				Name:      "admission-pss",
				HostPath:  admissionPodSecurityConfigFilePath,
				MountPath: admissionPodSecurityConfigFilePath,
				ReadOnly:  true,
				PathType:  corev1.HostPathFile,
			},
		)
	}

	if kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs == nil {
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs = map[string]string{}
	}
	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs[admissionPodSecurityConfigFlagName] = admissionPodSecurityConfigFilePath

	return kcp
}
