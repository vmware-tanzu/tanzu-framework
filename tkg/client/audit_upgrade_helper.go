// Copyright 2023 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"encoding/base64"
	"strings"

	"github.com/pkg/errors"
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
			decodedstr := ""
			if f.Encoding == "base64" || f.Encoding == "" {
				decoded, err := base64.StdEncoding.DecodeString(f.Content)
				if err != nil {
					return nil, err
				}
				decodedstr = string(decoded)
			} else if f.Encoding == "gzip" || f.Encoding == "gzip+base64" {
				// TODO we only have base64 at the moment in legacy cluster's yaml, do nothing here in case do has different encodings
				return nil, nil
			} else {
				return nil, errors.Errorf("unknown audit content encoding %s", f.Encoding)
			}

			if strings.Contains(decodedstr, apiVersionV1Alpha1) || strings.Contains(decodedstr, apiVersionV1Beta1) {
				kcp := old.DeepCopy()
				newFile := f.DeepCopy()
				newContent := strings.ReplaceAll(decodedstr, apiVersionV1Beta1, apiVersionV1)
				newContent = strings.ReplaceAll(newContent, apiVersionV1Alpha1, apiVersionV1)
				newFile.Content = base64.StdEncoding.EncodeToString([]byte(newContent))
				kcp.Spec.KubeadmConfigSpec.Files[i] = *newFile
				return kcp, nil
			} else {
				return nil, nil
			}
		}
	}
	return nil, nil
}
