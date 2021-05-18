// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/client"
)

// GetCEIP returns CEIP status set on management cluster
func (t *tkgctl) GetCEIP() (client.ClusterCeipInfo, error) {
	ceipStatus, err := t.tkgClient.GetCEIPParticipation()
	if err != nil {
		return client.ClusterCeipInfo{}, err
	}
	return ceipStatus, nil
}
