// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package propagation

import (
	"fmt"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v2/object-propagation/config"
)

func NewConfig(configEntry *config.Entry) *Config {
	sourceObject := &unstructured.Unstructured{}

	sourceObject.SetAPIVersion(configEntry.Source.APIVersion)
	sourceObject.SetKind(configEntry.Source.Kind)
	sourceObjectList := &unstructured.UnstructuredList{}
	sourceObjectList.SetAPIVersion(configEntry.Source.APIVersion)
	sourceObjectList.SetKind(fmt.Sprintf("%sList", configEntry.Source.Kind))

	propagationConfig := Config{
		SourceNamespace: configEntry.Source.Namespace,
		ObjectType:      sourceObject,
		ObjectListType:  sourceObjectList,
	}

	for _, s := range []struct {
		str      string
		selector *labels.Selector
	}{
		{configEntry.Source.LabelSelector, &propagationConfig.SourceSelector},
		{configEntry.Target.NamespaceLabelSelector, &propagationConfig.TargetNSSelector},
	} {
		var err error
		if *s.selector, err = labels.Parse(s.str); err != nil {
			panic(errors.Wrapf(err, "Error parsing selector '%s'", s.str))
		}
	}

	return &propagationConfig
}

func Configs(configEntries []*config.Entry) []*Config {
	var result []*Config
	for _, configEntry := range configEntries {
		result = append(result, NewConfig(configEntry))
	}
	return result
}
