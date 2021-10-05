// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"strings"

	dockerParser "github.com/novln/docker-parser"
	"github.com/pkg/errors"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

// parseRegistryImageURL parses the registry image URL to get repository and tag, tag is empty if not specified
func parseRegistryImageURL(imgURL string) (repository, tag string, err error) {
	ref, err := dockerParser.Parse(imgURL)
	if err != nil {
		return "", "", err
	}

	tag = ref.Tag()
	// dockerParser sets the tag to "latest" if not specified, however we want it to be empty
	if tag == tkgpackagedatamodel.DefaultRepositoryImageTag && !strings.HasSuffix(imgURL, ":"+tkgpackagedatamodel.DefaultRepositoryImageTag) {
		tag = ""
	}
	return ref.Repository(), tag, nil
}

// GetCurrentRepositoryAndTagInUse fetches the current tag used by package repository, taking tagselection into account
func GetCurrentRepositoryAndTagInUse(pkgr *kappipkg.PackageRepository) (repository, tag string, err error) {
	if pkgr.Spec.Fetch == nil || pkgr.Spec.Fetch.ImgpkgBundle == nil {
		return "", "", errors.New("failed to find OCI registry URL")
	}

	repository, tag, err = parseRegistryImageURL(pkgr.Spec.Fetch.ImgpkgBundle.Image)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse OCI registry URL")
	}

	if tag == "" {
		tag = tkgpackagedatamodel.LatestReleaseTag
	}

	return repository, tag, nil
}
