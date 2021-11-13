// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package carvelhelpers implements wrapper functions to use carvel tooling
package carvelhelpers

import (
	"bytes"
	"os"

	"github.com/cppforlife/go-cli-ui/ui"
	"github.com/k14s/kbld/pkg/kbld/cmd"
	"github.com/pkg/errors"
)

// ResolveImagesInPackage resolves the images using kbld tool
// Implements similar functionality as `kbld -f <file1> -f <file2>`
func ResolveImagesInPackage(files []string) ([]byte, error) {
	var outputBuf, errorBuf bytes.Buffer
	writerUI := ui.NewWriterUI(&outputBuf, &errorBuf, nil)
	kbldResolveOptions := cmd.NewResolveOptions(writerUI)
	kbldResolveOptions.FileFlags = cmd.FileFlags{Files: files}
	kbldResolveOptions.BuildConcurrency = 1

	// backup and reset stderr to avoid kbld to write anything to stderr
	stdErr := os.Stderr
	os.Stderr = nil
	err := kbldResolveOptions.Run()
	os.Stderr = stdErr

	if err != nil {
		return nil, errors.Wrapf(err, "error while resolving images")
	}
	return outputBuf.Bytes(), nil
}
