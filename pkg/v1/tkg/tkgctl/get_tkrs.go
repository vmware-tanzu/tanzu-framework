// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"

// GetTanzuKubernetesReleases returns the available TanzuKubernetesReleases
func (t *tkgctl) GetTanzuKubernetesReleases(tkrName string) ([]runv1alpha1.TanzuKubernetesRelease, error) {
	tkrs, err := t.tkgClient.GetTanzuKubernetesReleases(tkrName)
	if err != nil {
		return nil, err
	}

	return tkrs, nil
}
