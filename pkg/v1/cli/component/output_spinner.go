// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package component

import (
	"fmt"
	"io"
	"time"

	"github.com/briandowns/spinner"
)

// OutputWriterSpinner is OutputWriter augmented with a spinner.
type OutputWriterSpinner interface {
	OutputWriter
	RenderWithSpinner()
	StopSpinner()
}

// outputwriterspinner is our internal implementation.
type outputwriterspinner struct {
	outputwriter
	spinnerText string
	spinner     *spinner.Spinner
}

// NewOutputWriterWithSpinner returns implementation of OutputWriterSpinner.
func NewOutputWriterWithSpinner(output io.Writer, outputFormat, spinnerText string, startSpinner bool, headers ...string) (OutputWriterSpinner, error) {
	ows := &outputwriterspinner{}
	ows.out = output
	ows.outputFormat = OutputType(outputFormat)
	ows.keys = headers
	if ows.outputFormat != JSONOutputType && ows.outputFormat != YAMLOutputType {
		ows.spinnerText = spinnerText
		ows.spinner = spinner.New(spinner.CharSets[9], 100*time.Millisecond)
		if err := ows.spinner.Color("bgBlack", "bold", "fgWhite"); err != nil {
			return nil, err
		}
		ows.spinner.Suffix = fmt.Sprintf(" %s", spinnerText)
		if startSpinner {
			ows.spinner.Start()
		}
	}
	return ows, nil
}

// RenderWithSpinner will stop spinner and render the output
func (ows *outputwriterspinner) RenderWithSpinner() {
	if ows.spinner != nil && ows.spinner.Active() {
		ows.spinner.Stop()
		fmt.Fprintln(ows.out)
	}
	ows.Render()
}

// stop spinner
func (ows *outputwriterspinner) StopSpinner() {
	if ows.spinner != nil && ows.spinner.Active() {
		ows.spinner.Stop()
		fmt.Fprintln(ows.out)
	}
}
