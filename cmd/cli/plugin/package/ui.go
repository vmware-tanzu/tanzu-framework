// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"io"
	"strings"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/spf13/cobra"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/component"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
)

type AdapterUI struct {
	ui.WriterUI
	outWriter    io.Writer
	outputFormat string
}

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

func (adapterUI *AdapterUI) PrintTable(table uitable.Table) {
	keys := []string{}
	for _, h := range table.Header {
		if !h.Hidden {
			keys = append(keys, strings.ToUpper(strings.ReplaceAll(h.Title, " ", "-")))
		}
	}

	if table.Transpose {
		if adapterUI.outputFormat != string(component.JSONOutputType) && adapterUI.outputFormat != string(component.YAMLOutputType) {
			// For table output, we want to force the list table format for this part
			adapterUI.outputFormat = string(component.ListTableOutputType)
		}
	}

	t := component.NewOutputWriter(adapterUI.outWriter, adapterUI.outputFormat, keys...)

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

func setOutputFormat(cmd *cobra.Command, adapterUI *AdapterUI) {
	for _, c := range cmd.Commands() {
		if len(c.Commands()) > 1 {
			setOutputFormat(c, adapterUI)
		}
		var output string
		if _, ok := c.Annotations["table"]; ok {
			c.Flags().StringVarP(&output, "output", "o", "", "Output format (yaml|json|table), optional")
			c.PreRun = func(_ *cobra.Command, args []string) { adapterUI.SetOutputFormat(output) }
		}
	}
}
