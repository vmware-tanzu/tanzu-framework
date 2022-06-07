// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package source

import "sigs.k8s.io/yaml"

func normalizeYAML(content []byte) string {
	var parsed map[string]interface{}
	err := yaml.Unmarshal(content, &parsed)
	if err != nil {
		panic(err)
	}
	bytes, err := yaml.Marshal(parsed)
	if err != nil {
		panic(err)
	}
	return string(bytes)
}
