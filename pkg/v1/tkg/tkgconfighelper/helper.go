// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgconfighelper

import (
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgconfigreaderwriter"
)

// IsCustomRepository is custom image repository
func IsCustomRepository(imageRepo string) bool {
	return !strings.Contains(imageRepo, constants.TKGRegistryContains)
}

// SkipImageReferenceUpdateOnUpgrade returns true if environment variable is set
func SkipImageReferenceUpdateOnUpgrade() bool {
	return os.Getenv("TKG_SKIP_IMAGE_REFERENCE_UPDATE_ON_UPGRADE") != ""
}

// GetIntegerVariableFromConfig returns integer variable from config file
func GetIntegerVariableFromConfig(variable string, tkgConfigReaderWriter tkgconfigreaderwriter.TKGConfigReaderWriter) (int, error) {
	varValue, err := tkgConfigReaderWriter.Get(variable)
	if err != nil {
		return 0, errors.Errorf("%s variable is not set", variable)
	}

	intValue, err := strconv.Atoi(varValue)
	if err != nil {
		return 0, errors.Errorf("invalid %s", variable)
	}
	if intValue == 0 {
		return 0, errors.Errorf("%s cannot be 0", variable)
	}
	return intValue, nil
}
