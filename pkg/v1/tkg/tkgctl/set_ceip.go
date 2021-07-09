// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// SetCeip sets CEIP to the management cluster
func (t *tkgctl) SetCeip(ceipOptIn, isProd, labels string) error {
	ceipOptInBool, err := strconv.ParseBool(ceipOptIn)
	if err != nil {
		return errors.Errorf("cannot parse provided boolean argument, '%s'. Expected 'true' or 'false'", ceipOptIn)
	}
	if err := isValidLabels(labels); err != nil {
		return err
	}
	err = t.tkgClient.SetCEIPParticipation(ceipOptInBool, isProd, labels)
	if err != nil {
		return err
	}
	return nil
}

func isValidLabels(labels string) error {
	if labels == "" {
		return nil
	}

	re := regexp.MustCompile("^[a-zA-Z0-9]*$")
	individualLabels, err := convertLabelsToMap(labels)
	if err != nil {
		return err
	}

	if val, ok := individualLabels["entitlement-account-number"]; ok {
		if !re.MatchString(val) {
			return errors.Errorf("entitlement-account-number: %s cannot contain special characters", val)
		}
	}

	if val, ok := individualLabels["env-type"]; ok {
		supportedEnvironmentTypes := getSupportedEnvironmentTypes()
		if _, ok := supportedEnvironmentTypes[val]; !ok {
			return errors.Errorf("Invalid error type %s, environment type can be production, development, or test", val)
		}
	}

	return nil
}

func convertLabelsToMap(labels string) (map[string]string, error) {
	labelArr := strings.Split(labels, ",")
	permittedNumberOfLabels := 2
	if len(labelArr) > permittedNumberOfLabels {
		return nil, errors.Errorf("There are more labels provided than are currently supported. The supported labels are entitlement-account-number,and env-type")
	}
	individualLabels := make(map[string]string)
	for i := range labelArr {
		keyVal := strings.Split(labelArr[i], "=")
		if len(keyVal) != permittedNumberOfLabels {
			return nil, errors.Errorf("The individual labels are formed incorrectly, use -h to add them correctly")
		}
		if keyVal[0] != "entitlement-account-number" && keyVal[0] != "env-type" {
			return nil, errors.Errorf("Incorrect key provided, the currently supported keys are entitlement-account-number, and env-type")
		}
		individualLabels[keyVal[0]] = keyVal[1]
	}
	return individualLabels, nil
}

func getSupportedEnvironmentTypes() map[string]bool {
	supportedEnvironmentTypes := make(map[string]bool)
	supportedEnvironmentTypes["production"] = true
	supportedEnvironmentTypes["development"] = true
	supportedEnvironmentTypes["test"] = true

	return supportedEnvironmentTypes
}
