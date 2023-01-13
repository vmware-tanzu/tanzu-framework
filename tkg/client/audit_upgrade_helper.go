// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/base64"
	"strings"

	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
)

const (
	auditConfigFilePath = "/etc/kubernetes/audit-policy.yaml"
	apiVersionV1Alpha1  = "apiVersion: audit.k8s.io/v1alpha1"
	apiVersionV1Beta1   = "apiVersion: audit.k8s.io/v1beta1"
	apiVersionV1        = "apiVersion: audit.k8s.io/v1"
)

func (c *TkgClient) configureAuditVersion(old *controlplanev1.KubeadmControlPlane) (*controlplanev1.KubeadmControlPlane, error) {
	for i, f := range old.Spec.KubeadmConfigSpec.Files {
		if f.Path == auditConfigFilePath {
			decoded, err := base64.StdEncoding.DecodeString(f.Content)
			if err != nil {
				return nil, err
			}
			if strings.Contains(string(decoded), apiVersionV1Alpha1) || strings.Contains(string(decoded), apiVersionV1Beta1) {
				kcp := old.DeepCopy()
				newFile := f.DeepCopy()
				newContent := strings.ReplaceAll(string(decoded), apiVersionV1Beta1, apiVersionV1)
				newContent = strings.ReplaceAll(newContent, apiVersionV1Alpha1, apiVersionV1)
				newFile.Content = base64.StdEncoding.EncodeToString([]byte(newContent))
				kcp.Spec.KubeadmConfigSpec.Files[i] = *newFile
				return kcp, nil
			}
		}
	}
	return nil, nil
}
