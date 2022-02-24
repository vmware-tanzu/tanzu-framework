// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers

import (
	"io"

	yttui "github.com/vmware-tanzu/carvel-ytt/pkg/cmd/ui"
	"github.com/vmware-tanzu/carvel-ytt/pkg/files"
	"github.com/vmware-tanzu/carvel-ytt/pkg/workspace"
	"github.com/vmware-tanzu/carvel-ytt/pkg/workspace/datavalues"
)

// ProcessYTTPackage processes configuration directory with ytt tool
// Implements similar functionality as `ytt -f <config-dir>`
func ProcessYTTPackage(configDir string) ([]byte, error) {
	yttFiles, err := files.NewSortedFilesFromPaths([]string{configDir}, files.SymlinkAllowOpts{})
	if err != nil {
		return nil, err
	}

	lib := workspace.NewRootLibrary(yttFiles)
	libCtx := workspace.LibraryExecutionContext{Current: lib, Root: lib}
	libExecFact := workspace.NewLibraryExecutionFactory(&NoopUI{}, workspace.TemplateLoaderOpts{})
	loader := libExecFact.New(libCtx)

	valuesDoc, libraryValueDoc, err := loader.Values([]*datavalues.Envelope{}, datavalues.NewNullSchema())
	if err != nil {
		return nil, err
	}
	result, err := loader.Eval(valuesDoc, libraryValueDoc, []*datavalues.SchemaEnvelope{})
	if err != nil {
		return nil, err
	}
	return result.DocSet.AsBytes()
}

// NoopUI implement noop interface for logging used with carvel tooling
type NoopUI struct{}

var _ yttui.UI = NoopUI{}

// Printf noop print
func (ui NoopUI) Printf(str string, args ...interface{}) {}

// Debugf noop debug
func (ui NoopUI) Debugf(str string, args ...interface{}) {}

// Warnf noop warn
func (ui NoopUI) Warnf(str string, args ...interface{}) {}

// DebugWriter noop debug writer
func (ui NoopUI) DebugWriter() io.Writer {
	return noopWriter{}
}

type noopWriter struct{}

func (n noopWriter) Write(p []byte) (int, error) {
	return 0, nil
}

var _ io.Writer = noopWriter{}
