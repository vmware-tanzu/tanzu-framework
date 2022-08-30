// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import "github.com/vmware-tanzu/tanzu-framework/tkg/constants"

func (t *tkgctl) setCEIPOptinBasedOnConfigAndBuildEdition(edition string) string {
	ceipOptIn, err := t.TKGConfigReaderWriter().Get(constants.ConfigVariableEnableCEIPParticipation)
	if err == nil {
		return ceipOptIn
	}

	if edition == TCEBuildEdition {
		return False
	}

	return True
}
