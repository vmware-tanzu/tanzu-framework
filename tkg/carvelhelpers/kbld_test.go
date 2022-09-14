// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package carvelhelpers

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ResolveImagesInPackage_Success(t *testing.T) {
	assert := assert.New(t)

	input := `---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
- annotations:
    kbld.carvel.dev/id: test-tanzu-cli-plugins/foo-darwin-amd64:latest
  image: localhost:5000/tanzu-plugins/standalone-plugins@sha256:443df31fec8f78b55ea25ac1ba55907567fbdf05301a752e9f6adefe3c37e11d
- annotations:
    kbld.carvel.dev/id: test-tanzu-cli-plugins/foo-linux-amd64:latest
  image: localhost:5000/tanzu-plugins/standalone-plugins@sha256:53fb40d45e6a1267713cfdd6b561a65457cd5a575c6813fbdae06380b48e5e1f
kind: ImagesLock
---
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: Foo
metadata:
  name: foo-test
spec:
  artifacts:
    - image: test-tanzu-cli-plugins/foo-darwin-amd64:latest
    - image: test-tanzu-cli-plugins/foo-linux-amd64:latest
  description: contains artifacts
`

	output := `---
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: Foo
metadata:
  name: foo-test
spec:
  artifacts:
  - image: localhost:5000/tanzu-plugins/standalone-plugins@sha256:443df31fec8f78b55ea25ac1ba55907567fbdf05301a752e9f6adefe3c37e11d
  - image: localhost:5000/tanzu-plugins/standalone-plugins@sha256:53fb40d45e6a1267713cfdd6b561a65457cd5a575c6813fbdae06380b48e5e1f
  description: contains artifacts
`

	f, err := os.CreateTemp("", "kbld_test")
	assert.Nil(err)
	defer os.Remove(f.Name())
	err = os.WriteFile(f.Name(), []byte(input), 0644)
	assert.Nil(err)

	bytes, err := ResolveImagesInPackage([]string{f.Name()})
	assert.Nil(err)
	assert.NotNil(bytes)
	assert.Equal(output, string(bytes))
}

func Test_ResolveImagesInPackage_When_Image_Not_Present_In_ImageLock(t *testing.T) {
	assert := assert.New(t)

	input1 := `---
apiVersion: imgpkg.carvel.dev/v1alpha1
images:
- annotations:
    kbld.carvel.dev/id: test-tanzu-cli-plugins/foo-darwin-amd64:latest
  image: localhost:5000/tanzu-plugins/standalone-plugins@sha256:443df31fec8f78b55ea25ac1ba55907567fbdf05301a752e9f6adefe3c37e11d
kind: ImagesLock
---
apiVersion: cli.tanzu.vmware.com/v1alpha1
kind: Foo
metadata:
  name: foo-test
spec:
  artifacts:
    - image: test-tanzu-cli-plugins/foo-darwin-amd64:latest
    - image: test-tanzu-cli-plugins/foo-linux-amd64:latest
  description: contains artifacts
`

	f, err := os.CreateTemp("", "kbld_test")
	assert.Nil(err)
	defer os.Remove(f.Name())
	err = os.WriteFile(f.Name(), []byte(input1), 0644)
	assert.Nil(err)

	bytes, err := ResolveImagesInPackage([]string{f.Name()})
	assert.NotNil(err)
	assert.Contains(err.Error(), "error while resolving images:")
	assert.Nil(bytes)
}
