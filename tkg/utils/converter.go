// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"strconv"

	"gopkg.in/yaml.v3"
)

func Convert(in string) interface{} {
	convs := []yamlScalarConvertable{
		nullConvertable,
		integerConvertable,
		booleanConvertable,
		floatConvertable,
		structuredConvertable,
	}

	for _, conv := range convs {
		value, convertable := conv(in)
		if convertable {
			return value
		}
	}
	return in
}

type yamlScalarConvertable func(in string) (interface{}, bool)

func structuredConvertable(in string) (interface{}, bool) {
	var result interface{}
	if err := yaml.Unmarshal([]byte(in), &result); err == nil && result != nil {
		return result, true
	}
	return result, false
}

func nullConvertable(in string) (interface{}, bool) {
	return nil, (in == "~" || in == "null")
}

func booleanConvertable(in string) (interface{}, bool) {
	if v, err := strconv.ParseBool(in); err == nil {
		return v, true
	}
	return false, false
}

func integerConvertable(in string) (interface{}, bool) {
	if v, err := strconv.ParseUint(in, 0, 0); err == nil {
		return v, true
	}
	return 0, false
}

func floatConvertable(in string) (interface{}, bool) {
	if v, err := strconv.ParseFloat(in, 64); err == nil {
		return v, true
	}
	return 0, false
}
