// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Target to template files.
type Target struct {
	// Path of the file.
	Filepath string

	// Template to use.
	Template string
}

// Run the target.
func (t Target) Run(rootDir string, data interface{}, dryRun bool) error {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
		"ToLower": strings.ToLower,
	}
	tmplFp := template.Must(template.New("target-fp").Funcs(funcMap).Parse(t.Filepath))
	bufFp := &bytes.Buffer{}
	err := tmplFp.Execute(bufFp, data)
	if err != nil {
		return err
	}
	fp := filepath.Join(rootDir, bufFp.String())

	buf := &bytes.Buffer{}
	tmpl := template.Must(template.New("target").Funcs(funcMap).Parse(t.Template))
	if err := tmpl.Execute(buf, data); err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("-- file: %s --\n\n%s", t.Filepath, buf.String())
		return nil
	}
	dir := filepath.Dir(fp)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}
	err = os.WriteFile(fp, buf.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

// DefaultInitTargets are the default initialization targets.
var DefaultInitTargets = []Target{
	GoMod,
	BuildVersion,
	GitIgnore,
	Makefile,
	Codeowners,
	MainReadMe,
	GolangCIConfig,
	Tools,
	CommonMK,
}

// DefaultPluginTargets are the default plugin targets.
var DefaultPluginTargets = []Target{
	PluginReadMe,
	PluginMain,
	PluginTest,
}
