// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package artifact

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
)

// When local artifact directory exists and test directory exists
func TestLocalArtifactWhenArtifactAndTestDirExists(t *testing.T) {
	assert := assert.New(t)

	testPluginPath, err := filepath.Abs(filepath.Join("test", "local", "v0.0.1", "tanzu-plugin-darwin_amd64"))
	assert.NoError(err)
	artifact := NewLocalArtifact(testPluginPath)

	b, err := artifact.Fetch()
	assert.NoError(err)
	assert.Contains(string(b), "plugin binary")

	b, err = artifact.FetchTest()
	assert.NoError(err)
	assert.Contains(string(b), "test plugin binary")
}

// When local artifact doesn't exists and multiple files exists within test directory
func TestLocalArtifactWhenArtifactDoesntExistsAndMultipleFilesUnderTest(t *testing.T) {
	assert := assert.New(t)

	testPluginPath, err := filepath.Abs(filepath.Join("test", "local", "v0.0.2", "plugin_binary_does_not_exists"))
	assert.NoError(err)

	artifact := NewLocalArtifact(testPluginPath)

	// When local artifact doesn't exists
	_, err = artifact.Fetch()
	assert.Error(err)
	assert.ErrorContains(err, "error while reading artifact")

	// When multiple files exists under test directory
	_, err = artifact.FetchTest()
	assert.Error(err)
	assert.ErrorContains(err, "expected only 1 file under the")
}

var _ = Describe("Unit tests for local artifact", func() {
	Context("when input file not exists", func() {
		It("should return current token", func() {
			localArtifact := NewLocalArtifact("file://home/user/")
			Expect(localArtifact).NotTo(BeNil())
			bytes, err1 := localArtifact.Fetch()
			Expect(bytes).To(BeNil())
			Expect(err1.Error()).To(ContainSubstring("error while reading artifact"))
		})
	})
})
