// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"strings"

	"gopkg.in/yaml.v2"

	dockerParser "github.com/novln/docker-parser"
	"github.com/pkg/errors"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

// packageRepositoryStdout is used for unmarshal the package repository stdout to get the tag in use
type packageRepositoryStdout struct {
	ApiVersion  string `yaml:"apiVersion,omitempty"`
	Directories []struct {
		Contents []struct {
			ImgpkgBundle struct {
				Tag string `yaml:"tag,omitempty"`
			} `yaml:"imgpkgBundle,omitempty"`
		} `yaml:"contents,omitempty"`
	} `yaml:"directories,omitempty"`
}

// parseRegistryImageUrl parses the registry image URL to get repository and tag, tag is empty if not specified
func parseRegistryImageUrl(imgUrl string) (repository string, tag string, err error) {
	ref, err := dockerParser.Parse(imgUrl)
	if err != nil {
		return "", "", err
	}

	tag = ref.Tag()
	// dockerParser will default the tag to be latest if not specified, however we want it to be empty
	if tag == tkgpackagedatamodel.DefaultRepositoryImageTag && !strings.HasSuffix(imgUrl, ":"+tkgpackagedatamodel.DefaultRepositoryImageTag) {
		tag = ""
	}
	return ref.Repository(), tag, nil
}

// GetCurrentRepositoryAndTagInUse fetches the current tag used by package repository, taking tagselection into account
func GetCurrentRepositoryAndTagInUse(pkgr *kappipkg.PackageRepository) (repository, tag string, err error) {
	if pkgr.Spec.Fetch == nil || pkgr.Spec.Fetch.ImgpkgBundle == nil {
		return "", "", errors.New("failed to find OCI registry URL")
	}

	repository, tag, err = parseRegistryImageUrl(pkgr.Spec.Fetch.ImgpkgBundle.Image)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to parse OCI registry URL")
	}

	if pkgr.Spec.Fetch.ImgpkgBundle.TagSelection != nil && pkgr.Spec.Fetch.ImgpkgBundle.TagSelection.Semver != nil && pkgr.Status.Fetch != nil {
		/* Unmarshall the tag from stdout
		   example format:
			stdout: |
		      apiVersion: vendir.k14s.io/v1alpha1
		      directories:
		      - contents:
		        - imgpkgBundle:
		            image: projects.registry.vmware.com/tce/main@sha256:984450d3b1367f761da43e443c36428614c8ce9012d9fc1f2149733de0149cf4
		            tag: 0.8.0
		          path: .
		        path: "0"
		      kind: LockConfig
		*/
		m := packageRepositoryStdout{}

		err := yaml.Unmarshal([]byte(pkgr.Status.Fetch.Stdout), &m)
		if err != nil {
			return "", "", err
		}

		if len(m.Directories) > 0 && len(m.Directories[0].Contents) > 0 && m.Directories[0].Contents[0].ImgpkgBundle.Tag != "" {
			tag = m.Directories[0].Contents[0].ImgpkgBundle.Tag
		}
	}

	return repository, tag, nil
}
