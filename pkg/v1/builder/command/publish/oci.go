// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package publish

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/aunum/log"
	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/cli/core/pkg/common"
)

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
	localDiscoveryPath string) (Publisher, error) {

	if ociDistributionImageRepository == "" {
		return nil, errors.New("OCI distribution image repository cannot be empty")
	}
	if ociDiscoveryImage == "" {
		return nil, errors.New("OCI discovery image cannot be empty")
	}
	if localDiscoveryPath == "" {
		return nil, errors.New("local discovery path cannot be empty")
	}
	// TODO: Add more validation for image repository and image format

	return &OCIPublisher{
		OCIDiscoveryImage:              ociDiscoveryImage,
		OCIDistributionImageRepository: ociDistributionImageRepository,
		LocalDiscoveryPath:             localDiscoveryPath,
	}, nil
}

// PublishPlugin publishes plugin binaries to OCI based distribution directory
func (o *OCIPublisher) PublishPlugin(sourcePath, version, os, arch, plugin string) (string, error) {
	// Create artifactImage with format: `image.registry.com/tanzu-cli-plugins/plugin-os-arch:version`
	artifactImage := fmt.Sprintf("%s/%s-%s-%s:%s", strings.Trim(o.OCIDistributionImageRepository, "/"), plugin, os, arch, version)
	log.Info("Publishing plugin:", plugin, "to", artifactImage)
	// TODO: Use imgpkg library directly instead of directly using CLI
	out, err := exec.Command("imgpkg", "push", "-i", artifactImage, "-f", filepath.Dir(sourcePath), "--file-exclusion", "test").CombinedOutput()
	if err != nil {
		return "", errors.Wrapf(err, "%v", string(out))
	}
	log.Success("Successfully published plugin:", plugin)
	return artifactImage, nil
}

// PublishDiscovery publishes the CLIPlugin resources YAML to a OCI based discovery container image
func (o *OCIPublisher) PublishDiscovery() error {
	log.Info("Publishing discovery image to:", o.OCIDiscoveryImage)
	// TODO: Use imgpkg library directly instead of directly using CLI
	out, err := exec.Command("imgpkg", "push", "-i", o.OCIDiscoveryImage, "-f", o.LocalDiscoveryPath).CombinedOutput() //nolint:gosec
	if err != nil {
		return errors.Wrapf(err, "%v", string(out))
	}
	log.Success("Successfully published CLIPlugin resources to discovery image:", o.OCIDiscoveryImage)
	return nil
}

// Type returns type of publisher
func (o *OCIPublisher) Type() string {
	return common.DistributionTypeOCI
}
