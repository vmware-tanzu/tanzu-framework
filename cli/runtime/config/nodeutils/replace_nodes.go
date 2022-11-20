// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ReplaceNodes replace nodes in dst as per patchStrategy prior performing merge
func ReplaceNodes(src, dst *yaml.Node, opts ...PatchStrategyOpts) (bool, error) {
	// only replace if the change is not equal to existing
	replaceUnequalObjects, err := NotEqual(src, dst)
	if err != nil {
		return false, err
	}
	if !replaceUnequalObjects {
		return replaceUnequalObjects, nil
	}

	options := &PatchStrategyOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return replaceUnequalObjects, replaceNodes(src, dst, options.Key, options.PatchStrategies)
}

func replaceNodes(src, dst *yaml.Node, patchStrategyKey string, patchStrategies map[string]string) error {
	err := checkErrors(src, dst)
	if err != nil {
		return err
	}
	switch dst.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(dst.Content); i += 2 {
			found := false
			key := patchStrategyKey
			for j := 0; j < len(src.Content); j += 2 {
				// If there is a match and equal continue noop
				if ok, _ := equalScalars(dst.Content[i], src.Content[j]); !ok {
					continue
				}

				found = true

				// check for patch strategy before performing deep replace
				key = fmt.Sprintf("%v.%v", key, dst.Content[i].Value)
				if strings.EqualFold(patchStrategies[key], "replace") {
					dst.Content = append(dst.Content[:i], dst.Content[i+1:]...)
					dst.Content = append(dst.Content[:i], dst.Content[i+1:]...)
					i--
					i--
					break
				}

				if err := replaceNodes(src.Content[j+1], dst.Content[i+1], key, patchStrategies); err != nil {
					return errors.New("replace at key " + src.Content[i].Value + ": " + err.Error())
				}
				key = patchStrategyKey
				break
			}
			// if match not found remove the node if it is found in patch strategy
			if !found {
				key = fmt.Sprintf("%v.%v", key, dst.Content[i].Value)
				if strings.EqualFold(patchStrategies[key], "replace") {
					dst.Content = append(dst.Content[:i], dst.Content[i+1:]...)
					dst.Content = append(dst.Content[:i], dst.Content[i+1:]...)
					i--
					i--
				}
			}
		}
	case yaml.ScalarNode:
	case yaml.SequenceNode:
	case yaml.DocumentNode:
		err := replaceNodes(src.Content[0], dst.Content[0], patchStrategyKey, patchStrategies)
		if err != nil {
			return errors.New("replace at key " + src.Content[0].Value + ": " + err.Error())
		}
	default:
		return errors.New("can only merge mapping nodes")
	}
	return nil
}
