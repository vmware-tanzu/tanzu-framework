/*
Copyright 2020 The TKG Contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package registry

//go:generate counterfeiter -o ../fakes/registy.go --fake-name Registry . Registry

// Registry defines the Registry interface
type Registry interface {
	// ListImageTags lists all tags of the given image.
	ListImageTags(imageName string) ([]string, error)
	// GetFile gets the file content bundled in the given image:tag.
	// If filename is empty, it will get the first file.
	GetFile(image string, tag string, filename string) ([]byte, error)
}
