// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package nodeutils

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// ReplaceNodes replace nodes from src to dst as per patchStrategy
func ReplaceNodes(src, dst *yaml.Node, options *PatchStrategyOptions) error {
	return replaceNodes(src, dst, options.Key, options.PatchStrategies)
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
				if ok, _ := equalScalars(dst.Content[i], src.Content[j]); !ok {
					continue
				}
				found = true
				key = fmt.Sprintf("%v.%v", key, dst.Content[i].Value)
				if err := replaceNodes(src.Content[j+1], dst.Content[i+1], key, patchStrategies); err != nil {
					return errors.New("at key " + src.Content[i].Value + ": " + err.Error())
				}
				key = patchStrategyKey
				break
			}
			if !found {
				key = fmt.Sprintf("%v.%v", key, dst.Content[i].Value)
				if patchStrategies[key] == "replace" {
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
	default:
		return errors.New("can only merge mapping nodes")
	}
	return nil
}
