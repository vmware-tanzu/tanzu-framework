// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"reflect"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

var (
	ErrDifferentArgumentsTypes = errors.New("src and dst must be of same type")
	ErrNonPointerArgument      = errors.New("dst must be a pointer")
)

// equalScalars returns true if two scalar nodes has same value
func equalScalars(left, right *yaml.Node) (bool, error) {
	if left.Kind == yaml.ScalarNode && right.Kind == yaml.ScalarNode {
		return left.Value == right.Value, nil
	}
	return false, errors.New("equals on non-scalars not implemented")
}

// MergeNodes to merge two yaml nodes src(source) to dst(destination) node
func MergeNodes(src, dst *yaml.Node) (bool, error) {
	// only replace if the change is not equal to existing
	mergeUnequalObjects, err := NotEqual(src, dst)
	if err != nil {
		return false, err
	}
	if !mergeUnequalObjects {
		return mergeUnequalObjects, nil
	}
	return mergeUnequalObjects, mergeNodes(src, dst)
}

func mergeNodes(src, dst *yaml.Node) error {
	err := checkErrors(src, dst)
	if err != nil {
		return err
	}
	switch src.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(src.Content); i += 2 {
			found := false
			for j := 0; j < len(dst.Content); j += 2 {
				if ok, _ := equalScalars(src.Content[i], dst.Content[j]); ok {
					found = true
					if err := mergeNodes(src.Content[i+1], dst.Content[j+1]); err != nil {
						return errors.New("merge at key " + src.Content[i].Value + ": " + err.Error())
					}
					break
				}
			}
			if !found {
				dst.Content = append(dst.Content, src.Content[i:i+2]...)
			}
		}
	case yaml.SequenceNode:
		setSeqNode(src, dst)
	case yaml.DocumentNode:
		err := mergeNodes(src.Content[0], dst.Content[0])
		if err != nil {
			return errors.New("merge at key " + src.Content[0].Value + ": " + err.Error())
		}
	case yaml.ScalarNode:
		setScalarNode(src, dst)
	default:
		return errors.New("can only merge mapping and sequence nodes")
	}
	return nil
}

// Construct unique sequence nodes for scalar value type
func setSeqNode(src, dst *yaml.Node) {
	if dst.Content[0].Kind == yaml.ScalarNode && src.Content[0].Kind == yaml.ScalarNode {
		dst.Content = append(dst.Content, src.Content...)
		dst.Content = UniqNodes(dst.Content)
	}
}

func setScalarNode(src, dst *yaml.Node) {
	if dst.Value != src.Value {
		dst.Value = src.Value
	}
}

func checkErrors(src, dst *yaml.Node) error {
	if src.Kind != dst.Kind {
		return ErrDifferentArgumentsTypes
	}
	if dst != nil && reflect.ValueOf(dst).Kind() != reflect.Ptr {
		return ErrNonPointerArgument
	}
	return nil
}
