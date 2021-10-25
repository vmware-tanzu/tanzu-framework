// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publish

// TODO: to be implemented as part of https://github.com/vmware-tanzu/tanzu-framework/issues/946

// OCIPublisher defines OCI publisher configuration
type OCIPublisher struct {
	OCIDiscoveryImage              string
	OCIDistributionImageRepository string

	LocalDiscoveryPath string
}

// NewOCIPublisher create new OCI based publisher
func NewOCIPublisher(
	ociDiscoveryImage,
	ociDistributionImageRepository,
	localDiscoveryPath string) Publisher {

	return &OCIPublisher{
		OCIDiscoveryImage:              ociDiscoveryImage,
		OCIDistributionImageRepository: ociDistributionImageRepository,
		LocalDiscoveryPath:             localDiscoveryPath,
	}
}

// PublishPlugin publishes plugin binaries to OCI based distribution directory
func (o *OCIPublisher) PublishPlugin(version, os, arch, plugin, sourcePath string) (string, error) {
	return "", nil
}

// PublishDiscovery publishes the CLIPlugin resources YAML to a OCI based discovery container image
func (o *OCIPublisher) PublishDiscovery() error {
	return nil
}
