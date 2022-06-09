// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package config provides structs for parsing object-propagation controller configuration.
package config

import (
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/yaml"
)

type Source struct {
	Namespace     string `json:"namespace"`
	APIVersion    string `json:"apiVersion"`
	Kind          string `json:"kind"`
	LabelSelector string `json:"labelSelector"`
}

type Target struct {
	NamespaceLabelSelector string `json:"namespaceLabelSelector"`
}

type Entry struct {
	Source Source `json:"source"`
	Target Target `json:"target"`
}

func Parse(bytes []byte) ([]*Entry, error) {
	var configEntries []*Entry
	if err := yaml.Unmarshal(bytes, &configEntries); err != nil {
		return nil, errors.Wrap(err, "parsing config")
	}
	if len(configEntries) == 0 {
		return nil, errors.New("no config entries parsed")
	}
	for _, entry := range configEntries {
		if err := validate(entry); err != nil {
			return nil, err
		}
	}
	return configEntries, nil
}

func validate(entry *Entry) error {
	if entry == nil {
		return errors.New("nil config entry")
	}
	if entry.Source.APIVersion == "" {
		return errors.New("source.apiVersion is empty")
	}
	if entry.Source.Kind == "" {
		return errors.New("source.kind is empty")
	}
	if entry.Source.Namespace == "" {
		return errors.New("source.namespace is empty")
	}
	if _, err := labels.Parse(entry.Source.LabelSelector); err != nil {
		return errors.Wrap(err, "parsing source.selector")
	}
	if _, err := labels.Parse(entry.Target.NamespaceLabelSelector); err != nil {
		return errors.Wrap(err, "parsing target.namespaceSelector")
	}
	return nil
}
