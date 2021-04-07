// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aunum/log"
	"github.com/spf13/afero"
	"github.com/vmware-tanzu-private/tkg-cli/pkg/utils"
)

func Test_CreateClusterCommand(t *testing.T) {

	for _, test := range []struct {
		testcase    string
		stringMatch []string
		preConfig   func()
	}{
		{
			testcase:    "When default tanzu config file does not exist",
			stringMatch: []string{"kind: Cluster", "name: test-cluster"},
		},
		{
			testcase:    "When default tanzu config file exists but current server is not configured",
			stringMatch: []string{"kind: Cluster", "name: test-cluster"},
			preConfig: func() {
				configureTanzuConfig("./testdata/tanzuconfig/config1.yaml")
			},
		},
		{
			testcase:    "When default tanzu config file exists and current server is configured",
			stringMatch: []string{"kind: Cluster", "name: test-cluster"},
			preConfig: func() {
				configureTanzuConfig("./testdata/tanzuconfig/config2.yaml")
			},
		},
	} {
		t.Run(fmt.Sprintf("%s", test.testcase), func(t *testing.T) {
			defer configureHomeDirectory()()
			out := captureStdoutStderr(runCreateClusterCmd)
			for _, str := range test.stringMatch {
				if !strings.Contains(out, str) {
					t.Fatalf("expected \"%s\" to contain \"%s\"", out, str)
				}
			}
		})
	}
}

func runCreateClusterCmd() {
	cmd := createClusterCmd
	cmd.SetArgs([]string{"test-cluster", "-i", "docker:v0.3.10", "-p", "dev", "-d"})
	cmd.Execute()
}

func captureStdoutStderr(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func configureHomeDirectory() func() {
	fs := new(afero.MemMapFs)
	f, err := afero.TempDir(fs, "", "CreateClusterTest")
	if err != nil {
		log.Fatal(err)
	}
	os.Setenv("HOME", f)
	return func() {
		os.Unsetenv("HOME")
	}
}

func configureTanzuConfig(file string) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	tanzuDir := filepath.Join(dirname, ".tanzu")
	err = os.Mkdir(tanzuDir, 0600)
	if err != nil {
		log.Fatal(err)
	}
	err = utils.CopyFile(file, filepath.Join(tanzuDir, "config.yaml"))
	if err != nil {
		log.Fatal(err)
	}
}
