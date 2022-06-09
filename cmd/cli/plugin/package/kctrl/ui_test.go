// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kctrl

import (
	"bytes"
	"testing"

	"github.com/cppforlife/go-cli-ui/ui"
	uitable "github.com/cppforlife/go-cli-ui/ui/table"
	"github.com/stretchr/testify/assert"
)

func TestUI(t *testing.T) {
	t.Run("PrintLinef", func(t *testing.T) {
		t.Run("prints to outWriter with a trailing newline", func(t *testing.T) {
			uiOutBuffer := bytes.NewBufferString("")
			uiErrBuffer := bytes.NewBufferString("")
			ui := ui.NewWriterUI(uiOutBuffer, uiErrBuffer, ui.NewNoopLogger())
			adapterUI := AdapterUI{WriterUI: *ui, outWriter: uiOutBuffer}

			adapterUI.PrintLinef("fake-line")
			assert.Contains(t, "fake-line\n", uiOutBuffer.String())
			assert.Equal(t, "", uiErrBuffer.String())
		})

		t.Run("doesn't print target cluster information", func(t *testing.T) {
			uiOutBuffer := bytes.NewBufferString("")
			uiErrBuffer := bytes.NewBufferString("")
			ui := ui.NewWriterUI(uiOutBuffer, uiErrBuffer, ui.NewNoopLogger())
			adapterUI := AdapterUI{WriterUI: *ui, outWriter: uiOutBuffer}

			adapterUI.PrintLinef("Target cluster 127.0.0.1")
			assert.Equal(t, "", uiOutBuffer.String())
			assert.Equal(t, "", uiErrBuffer.String())
		})
	})

	t.Run("PrintTable", func(t *testing.T) {
		t.Run("prints table", func(t *testing.T) {
			uiOutBuffer := bytes.NewBufferString("")
			uiErrBuffer := bytes.NewBufferString("")
			ui := ui.NewWriterUI(uiOutBuffer, uiErrBuffer, ui.NewNoopLogger())
			adapterUI := AdapterUI{WriterUI: *ui, outWriter: uiOutBuffer}

			table := uitable.Table{
				Header: []uitable.Header{uitable.NewHeader("Header1"), uitable.NewHeader("Header2")},

				Rows: [][]uitable.Value{
					{uitable.ValueString{S: "r1c1"}, uitable.ValueString{S: "r1c2"}},
					{uitable.ValueString{S: "r2c1"}, uitable.ValueString{S: "r2c2"}},
				},
			}
			adapterUI.PrintTable(table)

			expected := `
  HEADER1  HEADER2  
  r1c1     r1c2     
  r2c1     r2c2     
`

			assert.Equal(t, expected, "\n"+uiOutBuffer.String())
		})

		t.Run("prints table in yaml format", func(t *testing.T) {
			uiOutBuffer := bytes.NewBufferString("")
			uiErrBuffer := bytes.NewBufferString("")
			ui := ui.NewWriterUI(uiOutBuffer, uiErrBuffer, ui.NewNoopLogger())
			adapterUI := AdapterUI{WriterUI: *ui, outWriter: uiOutBuffer}

			adapterUI.SetOutputFormat("yaml")

			table := uitable.Table{
				Header: []uitable.Header{uitable.NewHeader("Header1"), uitable.NewHeader("Header2")},

				Rows: [][]uitable.Value{
					{uitable.ValueString{S: "r1c1"}, uitable.ValueString{S: "r1c2"}},
					{uitable.ValueString{S: "r2c1"}, uitable.ValueString{S: "r2c2"}},
				},
			}
			adapterUI.PrintTable(table)

			expected := `
- header1: r1c1
  header2: r1c2
- header1: r2c1
  header2: r2c2
`

			assert.Equal(t, expected, "\n"+uiOutBuffer.String())
		})
	})
}
