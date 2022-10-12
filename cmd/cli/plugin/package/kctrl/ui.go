// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kctrl

import (
	"fmt"
	"io"
	"strings"

	"github.com/aunum/log"
	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/cli/runtime/component"
)

type AdapterUI struct {
	ui.WriterUI
	outWriter io.Writer

	outputFormat string
	tableCount   int
}

// PrintLinef overrides go-cli-ui/ui.PrintLinef
// It is used to print lines, but excludes certain lines based on the pattern
func (adapterUI *AdapterUI) PrintLinef(pattern string, args ...interface{}) {
	if strings.Contains(pattern, "Target cluster") {
		return
	}
	message := fmt.Sprintf(pattern, args...)
	_, err := fmt.Fprintln(adapterUI.outWriter, message)
	if err != nil {
		log.Error(err, "UI.PrintLinef failed (message='%s')", message)
	}
}

func (adapterUI *AdapterUI) SetOutputFormat(outputFormat string) {
	adapterUI.outputFormat = outputFormat
}

// PrintTable overrides go-cli-ui/ui.PrintTable
// It accepts a table and renders it based on the output format
//
//nolint:gocritic // Cannot change the function signature as it is defined in go-cli-ui
func (adapterUI *AdapterUI) PrintTable(table uitable.Table) {
	outputFormat := adapterUI.outputFormat // copy outputFormat so that it doesn't get overwritten in next steps in cases of multiple tables
	keys := []string{}

	adapterUI.tableCount++

	// For json output, we want to print only 1 table to make sure that it doesn't become an invalid json
	if adapterUI.tableCount > 1 && outputFormat == string(component.JSONOutputType) {
		return
	}

	for _, h := range table.Header {
		if !h.Hidden {
			keys = append(keys, strings.ToUpper(strings.ReplaceAll(h.Title, " ", "-")))
		}
	}

	if table.Transpose {
		if outputFormat != string(component.JSONOutputType) && outputFormat != string(component.YAMLOutputType) {
			// For table output, we want to force the list table format for this part
			outputFormat = string(component.ListTableOutputType)
		}
	}

	t := component.NewOutputWriter(adapterUI.outWriter, outputFormat, keys...)

	for _, row := range table.Rows {
		var tRow []interface{}
		for i, cell := range row {
			if !table.Header[i].Hidden {
				tRow = append(tRow, cell.Value())
			}
		}
		t.AddRow(tRow...)
	}

	t.Render()
}

func setOutputFormatFlag(cmd *cobra.Command, adapterUI *AdapterUI) {
	for _, c := range cmd.Commands() {
		if len(c.Commands()) > 1 {
			setOutputFormatFlag(c, adapterUI)
		}
		var output string
		if _, ok := c.Annotations["table"]; ok {
			c.Flags().StringVarP(&output, "output", "o", "", "Output format (yaml|json|table), optional")
			c.PreRun = func(_ *cobra.Command, args []string) { adapterUI.SetOutputFormat(output) }
		}
	}
}
