// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	capibootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

func TestTkgClient_configurePodSecurityStandard(t *testing.T) {
	c := &TkgClient{}

	kcp := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			KubeadmConfigSpec: capibootstrapv1.KubeadmConfigSpec{
				ClusterConfiguration: &capibootstrapv1.ClusterConfiguration{
					APIServer: capibootstrapv1.APIServer{
						ControlPlaneComponent: capibootstrapv1.ControlPlaneComponent{
							ExtraArgs: map[string]string{
								admissionPodSecurityConfigFlagName: admissionPodSecurityConfigFilePath,
							},
							ExtraVolumes: []capibootstrapv1.HostPathMount{
								{
									Name:      "admission-pss",
									HostPath:  admissionPodSecurityConfigFilePath,
									MountPath: admissionPodSecurityConfigFilePath,
									ReadOnly:  true,
									PathType:  corev1.HostPathFile,
								},
							},
						},
					},
				},
				Files: []capibootstrapv1.File{
					{
						Path:    admissionPodSecurityConfigFilePath,
						Content: admissionPodSecurityConfigFileData,
					},
				},
			},
		},
	}

	tests := []struct {
		name string
		old  *controlplanev1.KubeadmControlPlane
		want *controlplanev1.KubeadmControlPlane
	}{
		{
			"nil pointer check",
			&controlplanev1.KubeadmControlPlane{},
			kcp.DeepCopy(),
		},
		{
			"no-op",
			kcp.DeepCopy(),
			kcp.DeepCopy(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.configurePodSecurityStandard(tt.old); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TkgClient.configurePodSecurityStandard() = %v, want %v", got, tt.want)
			}
		})
	}
}
