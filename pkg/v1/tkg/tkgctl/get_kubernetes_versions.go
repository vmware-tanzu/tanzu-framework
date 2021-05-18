// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
)

func (t *tkgctl) GetKubernetesVersions() (*client.KubernetesVersionsInfo, error) {
	return t.tkgClient.GetKubernetesVersions()
}
