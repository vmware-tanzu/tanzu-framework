// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"bytes"
	"strings"

	"gopkg.in/yaml.v2"

	"github.com/helloeave/json"
	"github.com/jedib0t/go-pretty/table"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/log"
)

const indentation = `  `

const (
	renderOutputTypeJSON = "json"
	renderOutputTypeYAML = "yaml"
)

// RenderOutput renders output
func RenderOutput(data interface{}, renderType string) error {
	var err error
	switch strings.ToLower(renderType) {
	case renderOutputTypeJSON:
		err = renderJSON(data)
	case renderOutputTypeYAML:
		err = renderYAML(data)
	default:
		err = errors.Errorf("Invalid output format: %v", renderType)
	}
	return err
}

// renderJSON prints output as json
func renderJSON(data interface{}) error {
	bytesJSON, err := json.MarshalSafeCollections(data)
	if err != nil {
		return err
	}
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, bytesJSON, "", indentation)
	if err != nil {
		return err
	}
	log.Outputf("%v", prettyJSON.String())
	return nil
}

// renderYAML prints output as yaml
func renderYAML(data interface{}) error {
	yamlInBytes, err := yaml.Marshal(data)
	if err != nil {
		return err
	}
	log.Outputf("%v", string(yamlInBytes))
	return nil
}

// CreateTableWriter create a new table writer with default options
func CreateTableWriter() table.Writer {
	t := table.NewWriter()
	t.Style().Options.DrawBorder = false
	t.Style().Options.SeparateRows = false
	t.Style().Options.SeparateColumns = false
	t.Style().Options.SeparateFooter = false
	t.Style().Options.SeparateHeader = false
	return t
}
