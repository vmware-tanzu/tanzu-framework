// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v2"
)

const colWidth = 300
const indentation = `  `

// OutputWriter is an interface for something that can write output.
type OutputWriter interface {
	SetKeys(headerKeys ...string)
	AddRow(items ...interface{})
	Render()
}

// OutputType defines the format of the output desired.
type OutputType string

const (
	// TableOutputType specifies output should be in table format.
	TableOutputType OutputType = "table"
	// YAMLOutputType specifies output should be in yaml format.
	YAMLOutputType OutputType = "yaml"
	// JSONOutputType sepcifies output should be in json format.
	JSONOutputType OutputType = "json"
)

// outputwriter is our internal implementation.
type outputwriter struct {
	out          io.Writer
	keys         []string
	values       [][]string
	outputFormat OutputType
}

// NewOutputWriter gets a new instance of our output writer.
func NewOutputWriter(output io.Writer, outputFormat string, headers ...string) OutputWriter {
	// Initialize the output writer that we use under the covers
	ow := &outputwriter{}
	ow.out = output
	ow.outputFormat = OutputType(outputFormat)
	ow.keys = headers

	return ow
}

// SetKeys sets the values to use as the keys for the output values.
func (ow *outputwriter) SetKeys(headerKeys ...string) {
	// Overwrite whatever was used in initialization
	ow.keys = headerKeys
}

// AddRow appends a new row to our table.
func (ow *outputwriter) AddRow(items ...interface{}) {
	row := []string{}

	// Make sure all values are ultimately strings
	for _, item := range items {
		row = append(row, fmt.Sprintf("%v", item))
	}
	ow.values = append(ow.values, row)
}

// Render emits the generated table to the output once ready
func (ow *outputwriter) Render() {
	switch ow.outputFormat {
	case JSONOutputType:
		renderJSON(ow.out, ow.dataStruct())
	case YAMLOutputType:
		renderYAML(ow.out, ow.dataStruct())
	default:
		renderTable(ow)
	}
}

func (ow *outputwriter) dataStruct() []map[string]string {
	data := []map[string]string{}
	keys := ow.keys
	for i, k := range keys {
		keys[i] = strings.ToLower(strings.ReplaceAll(k, " ", "_"))
	}

	for _, itemValues := range ow.values {
		item := map[string]string{}
		for i, value := range itemValues {
			item[keys[i]] = value
		}
		data = append(data, item)
	}

	return data
}

// objectwriter is our internal implementation.
type objectwriter struct {
	out          io.Writer
	data         interface{}
	outputFormat OutputType
}

// NewObjectWriter gets a new instance of our output writer.
func NewObjectWriter(output io.Writer, outputFormat string, data interface{}) OutputWriter {
	// Initialize the output writer that we use under the covers
	obw := &objectwriter{}
	obw.out = output
	obw.data = data
	obw.outputFormat = OutputType(outputFormat)

	return obw
}

// SetKeys sets the values to use as the keys for the output values.
func (obw *objectwriter) SetKeys(headerKeys ...string) {
	// Object writer does not have the concept of keys
	fmt.Fprintln(obw.out, "Programming error, attempt to add headers to object output")
}

// AddRow appends a new row to our table.
func (obw *objectwriter) AddRow(items ...interface{}) {
	// Object writer does not have the concept of keys
	fmt.Fprintln(obw.out, "Programming error, attempt to add rows to object output")
}

// Render emits the generated table to the output once ready
func (obw *objectwriter) Render() {
	switch obw.outputFormat {
	case JSONOutputType:
		renderJSON(obw.out, obw.data)
	case YAMLOutputType:
		renderYAML(obw.out, obw.data)
	default:
		fmt.Fprintf(obw.out, "Invalid output format: %v\n", obw.outputFormat)
	}
}

// renderJSON prints output as json
func renderJSON(out io.Writer, data interface{}) {
	bytesJSON, err := json.MarshalIndent(data, "", indentation)
	if err != nil {
		fmt.Fprint(out, err)
		return
	}

	fmt.Fprintf(out, "%v", string(bytesJSON))
}

// renderYAML prints output as yaml
func renderYAML(out io.Writer, data interface{}) {
	yamlInBytes, err := yaml.Marshal(data)
	if err != nil {
		fmt.Fprint(out, err)
		return
	}

	fmt.Fprintf(out, "%s", yamlInBytes)
}

// renderTable prints output as a table
func renderTable(ow *outputwriter) {
	table := tablewriter.NewWriter(ow.out)
	table.SetBorder(false)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderLine(false)
	table.SetColWidth(colWidth)
	table.SetTablePadding("\t\t")
	table.SetHeader(ow.keys)
	colors := []tablewriter.Colors{}
	for range ow.keys {
		colors = append(colors, []int{tablewriter.Bold})
	}
	table.SetHeaderColor(colors...)
	table.AppendBulk(ow.values)
	table.Render()
}
