// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

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

// equalScalars returns true if two scalar nodes has same value
func equalScalars(left, right *yaml.Node) (bool, error) {
	if left.Kind == yaml.ScalarNode && right.Kind == yaml.ScalarNode {
		return left.Value == right.Value, nil
	}
	return false, errors.New("equals on non-scalars not implemented")
}

func checkErrors(src, dst *yaml.Node) error {
	if src.Kind != dst.Kind {
		return ErrDifferentArgumentsTypes
	}
	return nil
}
