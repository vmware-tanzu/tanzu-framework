// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/region"
)

// GetRegions return list of management clusters
func (t *tkgctl) GetRegions(managementClusterName string) ([]region.RegionContext, error) {
	return t.tkgClient.GetRegionContexts(managementClusterName)
}
