// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

// ActivateTanzuKubernetesReleases activates the given TKR
func (t *tkgctl) ActivateTanzuKubernetesReleases(tkrName string) error {
	err := t.tkgClient.ActivateTanzuKubernetesReleases(tkrName)
	if err != nil {
		return err
	}

	return nil
}

// DeactivateTanzuKubernetesReleases deactivates the given TKR
func (t *tkgctl) DeactivateTanzuKubernetesReleases(tkrName string) error {
	err := t.tkgClient.DeactivateTanzuKubernetesReleases(tkrName)
	if err != nil {
		return err
	}

	return nil
}
