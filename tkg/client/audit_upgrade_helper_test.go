// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"

	corev1 "k8s.io/api/core/v1"
	capibootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

const (
	auditConfigFileString = `
	apiVersion: %s
	kind: Policy
	rules:
	- level: None
	  users:
	  - system:serviceaccount:kube-system:kube-proxy
	  verbs:
	  - watch
	  resources:
	  - group: ""
		resources:
		- endpoints
		- services
		- services/status
	- level: None
	  userGroups:
	  - system:nodes
	  verbs:
	  - get
	  resources:
	  - group: ""
		resources:
		- nodes
		- nodes/status
	- level: None
	  users:
	  - system:kube-controller-manager
	  - system:kube-scheduler
	  - system:serviceaccount:kube-system:endpoint-controller
	  verbs:
	  - get
	  - update
	  namespaces:
	  - kube-system
	  resources:
	  - group: ""
		resources:
		- endpoints
	- level: None
	  users:
	  - system:apiserver
	  verbs:
	  - get
	  resources:
	  - group: ""
		resources:
		- namespaces
		- namespaces/status
		- namespaces/finalize
	- level: None
	  users:
	  - system:kube-controller-manager
	  verbs:
	  - get
	  - list
	  resources:
	  - group: metrics.k8s.io
	- level: None
	  nonResourceURLs:
	  - /healthz*
	  - /version
	  - /swagger*
	- level: None
	  resources:
	  - group: ""
		resources:
		- events
	- level: None
	  userGroups:
	  - system:serviceaccounts:vmware-system-tmc
	  verbs:
	  - get
	  - list
	  - watch
	- level: None
	  users:
	  - system:serviceaccount:kube-system:generic-garbage-collector
	  verbs:
	  - get
	  - list
	  - watch
	- level: Request
	  userGroups:
	  - system:nodes
	  verbs:
	  - update
	  - patch
	  resources:
	  - group: ""
		resources:
		- nodes/status
		- pods/status
	  omitStages:
	  - RequestReceived
	- level: Request
	  users:
	  - system:serviceaccount:kube-system:namespace-controller
	  verbs:
	  - deletecollection
	  omitStages:
	  - RequestReceived
	- level: Metadata
	  resources:
	  - group: ""
		resources:
		- secrets
		- configmaps
	  - group: authentication.k8s.io
		resources:
		- tokenreviews
	  omitStages:
	  - RequestReceived
	- level: Request
	  verbs:
	  - get
	  - list
	  - watch
	  resources:
	  - group: ""
	  - group: admissionregistration.k8s.io
	  - group: apiextensions.k8s.io
	  - group: apiregistration.k8s.io
	  - group: apps
	  - group: authentication.k8s.io
	  - group: authorization.k8s.io
	  - group: autoscaling
	  - group: batch
	  - group: certificates.k8s.io
	  - group: extensions
	  - group: metrics.k8s.io
	  - group: networking.k8s.io
	  - group: policy
	  - group: rbac.authorization.k8s.io
	  - group: settings.k8s.io
	  - group: storage.k8s.io
	  omitStages:
	  - RequestReceived
	- level: RequestResponse
	  resources:
	  - group: ""
	  - group: admissionregistration.k8s.io
	  - group: apiextensions.k8s.io
	  - group: apiregistration.k8s.io
	  - group: apps
	  - group: authentication.k8s.io
	  - group: authorization.k8s.io
	  - group: autoscaling
	  - group: batch
	  - group: certificates.k8s.io
	  - group: extensions
	  - group: metrics.k8s.io
	  - group: networking.k8s.io
	  - group: policy
	  - group: rbac.authorization.k8s.io
	  - group: settings.k8s.io
	  - group: storage.k8s.io
	  omitStages:
	  - RequestReceived
	- level: Metadata
	  omitStages:
	  - RequestReceived
	`
)

func TestTkgClient_updateAuditConfig(t *testing.T) {
	c := &TkgClient{}

	auditContentV1Alpha1 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(auditConfigFileString, "audit.k8s.io/v1alpha1")))
	kcp_v1aplha1 := &controlplanev1.KubeadmControlPlane{
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
									Name:      "audit-policy",
									HostPath:  auditConfigFilePath,
									MountPath: auditConfigFilePath,
									ReadOnly:  true,
									PathType:  corev1.HostPathFile,
								},
							},
						},
					},
				},
				Files: []capibootstrapv1.File{
					{
						Path:     auditConfigFilePath,
						Content:  auditContentV1Alpha1,
						Encoding: "base64",
					},
				},
			},
		},
	}

	auditContentV1Beta1 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(auditConfigFileString, "audit.k8s.io/v1beta1")))
	kcp_v1beta1 := &controlplanev1.KubeadmControlPlane{
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
									Name:      "audit-policy",
									HostPath:  auditConfigFilePath,
									MountPath: auditConfigFilePath,
									ReadOnly:  true,
									PathType:  corev1.HostPathFile,
								},
							},
						},
					},
				},
				Files: []capibootstrapv1.File{
					{
						Path:     auditConfigFilePath,
						Content:  auditContentV1Beta1,
						Encoding: "base64",
					},
				},
			},
		},
	}

	auditContentV1 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf(auditConfigFileString, "audit.k8s.io/v1")))
	kcp_v1 := &controlplanev1.KubeadmControlPlane{
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
									Name:      "audit-policy",
									HostPath:  auditConfigFilePath,
									MountPath: auditConfigFilePath,
									ReadOnly:  true,
									PathType:  corev1.HostPathFile,
								},
							},
						},
					},
				},
				Files: []capibootstrapv1.File{
					{
						Path:     auditConfigFilePath,
						Content:  auditContentV1,
						Encoding: "base64",
					},
				},
			},
		},
	}

	kcp_more_cases := &controlplanev1.KubeadmControlPlane{
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
									Name:      "audit-policy",
									HostPath:  auditConfigFilePath,
									MountPath: auditConfigFilePath,
									ReadOnly:  true,
									PathType:  corev1.HostPathFile,
								},
							},
						},
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
			"v1alpha1 replace check",
			kcp_v1aplha1.DeepCopy(),
			kcp_v1.DeepCopy(),
		},
		{
			"v1beta1 replace check",
			kcp_v1beta1.DeepCopy(),
			kcp_v1.DeepCopy(),
		},
		{
			"no-op",
			kcp_v1.DeepCopy(),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.configureAuditVersion(tt.old)
			if err != nil {
				t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() has error %v", err.Error())
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() = %v, want %v", got, tt.want)
			}
		})
	}

	kcp_more_cases.Spec.KubeadmConfigSpec.Files = []capibootstrapv1.File{
		{
			Path:     auditConfigFilePath,
			Content:  "12345",
			Encoding: "base64",
		},
	}
	t.Run("bad audit content", func(t *testing.T) {
		got, err := c.configureAuditVersion(kcp_more_cases)
		if err == nil || got != nil {
			t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() bad content in audit should return nil kcp and give error")
		}
	})

	kcp_more_cases.Spec.KubeadmConfigSpec.Files = []capibootstrapv1.File{
		{
			Path:     auditConfigFilePath,
			Content:  auditContentV1,
			Encoding: "encoding_not_exists",
		},
	}
	t.Run("bad audit content encoding type", func(t *testing.T) {
		got, err := c.configureAuditVersion(kcp_more_cases)
		if err == nil || got != nil {
			t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() bad content encoding type in audit should return nil kcp and give error")
		}
	})

	kcp_more_cases.Spec.KubeadmConfigSpec.Files = []capibootstrapv1.File{
		{
			Path:     auditConfigFilePath,
			Content:  auditContentV1,
			Encoding: "gzip",
		},
	}
	t.Run("gzip audit content encoding", func(t *testing.T) {
		got, err := c.configureAuditVersion(kcp_more_cases)
		if err != nil || got != nil {
			t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() gzip content encoding in audit should return nil kcp and give no error")
		}
	})

	kcp_more_cases.Spec.KubeadmConfigSpec.Files = []capibootstrapv1.File{}
	t.Run("empty files array", func(t *testing.T) {
		got, err := c.configureAuditVersion(kcp_more_cases)
		if err != nil || got != nil {
			t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() empty files array should return nil kcp and give no error")
		}
	})

	kcp_more_cases.Spec.KubeadmConfigSpec.Files = []capibootstrapv1.File{
		{
			Path:     "/etc/kubernetes/anotherfile.yaml",
			Content:  auditContentV1,
			Encoding: "base64",
		},
	}
	t.Run("have files but not audit file", func(t *testing.T) {
		got, err := c.configureAuditVersion(kcp_more_cases)
		if err != nil || got != nil {
			t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() files do not have audit file path should return nil kcp and give no error")
		}
	})

	kcp_more_cases.Spec.KubeadmConfigSpec.Files = []capibootstrapv1.File{
		{
			Path:     "/etc/kubernetes/anotherfile.yaml",
			Content:  auditContentV1,
			Encoding: "encoding_not_exists",
		},
	}
	t.Run("have none audit file with not existed encoding", func(t *testing.T) {
		got, err := c.configureAuditVersion(kcp_more_cases)
		if err != nil || got != nil {
			t.Errorf("TkgClient.TestTkgClient_updateAuditConfig() should skip file content check that is not audit yaml")
		}
	})
}
