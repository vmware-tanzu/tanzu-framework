// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import "gopkg.in/yaml.v3"

func UniqNodes(nodes []*yaml.Node) []*yaml.Node {
	uniq := make([]*yaml.Node, 0, len(nodes))
	mapper := make(map[string]bool)

	for _, node := range nodes {
		if _, ok := mapper[node.Value]; !ok {
			mapper[node.Value] = true
			uniq = append(uniq, node)
		}
	}

	return uniq
}
