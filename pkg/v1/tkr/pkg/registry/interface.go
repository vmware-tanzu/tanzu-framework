// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

//go:generate counterfeiter -o ../../fakes/registy.go --fake-name Registry . Registry

// Registry defines the Registry interface
type Registry interface {
	// ListImageTags lists all tags of the given image.
	ListImageTags(imageName string) ([]string, error)
	// GetFile gets the file content bundled in the given image:tag.
	// If filename is empty, it will get the first file.
	GetFile(imageWithTag string, filename string) ([]byte, error)
	// GetFiles get all the files content bundled in the given image:tag.
	GetFiles(imageWithTag string) (map[string][]byte, error)
	// DownloadBundle downloads OCI bundle similar to `imgpkg pull -b` command
	// It is recommended to use this function when downloading imgpkg bundle
	DownloadBundle(imageName, outputDir string) error
}
