// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/utils"
)

var dataValues = `#@data/values
#@overlay/match-child-defaults missing_ok=True

---
FOO_NAME: bar
`

var template = `---
apiVersion: foo.x-k8s.io/v1alpha3
kind: Foo
metadata:
  name: "${FOO_NAME}"
  labels:
    tkg.tanzu.vmware.com/label: '${FOO_NAME}'
`

var overlay = `#@ load("@ytt:overlay", "overlay")
#@ load("@ytt:data", "data")
#@ load("@ytt:yaml", "yaml")

#@overlay/match by=overlay.subset({"kind":"Foo"})
---
metadata:
  name: #@ data.values.FOO_NAME
  labels:
    tkg.tanzu.vmware.com/label: #@ data.values.FOO_NAME
`

var expectedOutput = `apiVersion: foo.x-k8s.io/v1alpha3
kind: Foo
metadata:
  name: bar
  labels:
    tkg.tanzu.vmware.com/label: bar
`

func Test_ProcessYTTPackage(t *testing.T) {
	assert := assert.New(t)

	testDir := setupTestDir(dataValues, template, overlay)
	defer os.RemoveAll(testDir)

	bytes, err := ProcessYTTPackage(testDir)
	assert.Nil(err)
	assert.NotNil(bytes)
	assert.Equal(expectedOutput, string(bytes))
}

func Test_ProcessYTTPackage_When_Error(t *testing.T) {
	assert := assert.New(t)

	overlayUpdated := strings.ReplaceAll(overlay, "FOO_NAME", "FAKE_NAME")
	testDir := setupTestDir(dataValues, template, overlayUpdated)
	defer os.RemoveAll(testDir)

	_, err := ProcessYTTPackage(testDir)
	assert.NotNil(err)
	assert.Contains(err.Error(), "struct has no .FAKE_NAME field or method ")
}

func setupTestDir(datavalues, overlay, template string) string {
	tempDir, _ := os.MkdirTemp("", "ytt_test")

	_ = utils.SaveFile(filepath.Join(tempDir, "datavalues.yaml"), []byte(datavalues))
	_ = utils.SaveFile(filepath.Join(tempDir, "overlay.yaml"), []byte(overlay))
	_ = utils.SaveFile(filepath.Join(tempDir, "template.yaml"), []byte(template))
	return tempDir
}
