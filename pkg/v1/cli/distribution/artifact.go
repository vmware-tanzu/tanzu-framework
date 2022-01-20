// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package distribution

import (
	"github.com/pkg/errors"

	cliv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/cli/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/cli/artifact"
)

// Artifact points to an individual plugin binary specific to a version and
// platform.
type Artifact struct {
	// Image is a fully qualified OCI image for the plugin binary.
	Image string

	// AssetURI is a URI of the plugin binary.
	URI string

	// SHA256 hash of the plugin binary.
	Digest string

	// OS of the plugin binary in `GOOS` format.
	OS string

	// Arch of the plugin binary in `GOARCH` format.
	Arch string
}

// ArtifactList contains an Artifact object for every supported platform of a
// version.
type ArtifactList []Artifact

// Artifacts contains an artifact list for every supported version.
type Artifacts map[string]ArtifactList

// GetArtifact returns Artifact object
func (aMap Artifacts) GetArtifact(version, os, arch string) (Artifact, error) {
	err := errors.Errorf(
		"could not find the artifact for version:%s, os:%s, arch:%s",
		version, os, arch)
	if aMap == nil {
		return Artifact{}, err
	}

	aList, ok := aMap[version]
	if !ok || aList == nil {
		return Artifact{}, err
	}

	for _, a := range aList {
		if a.OS == os && a.Arch == arch {
			return a, nil
		}
	}
	return Artifact{}, err
}

// Fetch the binary for a plugin version.
func (aMap Artifacts) Fetch(version, os, arch string) ([]byte, error) {
	a, err := aMap.GetArtifact(version, os, arch)
	if err != nil {
		return nil, err
	}

	if a.Image != "" {
		return artifact.NewOCIArtifact(a.Image).Fetch()
	}
	if a.URI != "" {
		u, err := artifact.NewURIArtifact(a.URI)
		if err != nil {
			return nil, err
		}
		return u.Fetch()
	}

	return nil, errors.Errorf("invalid artifact for version:%s, os:%s, "+
		"arch:%s", version, os, arch)
}

// GetDigest returns the SHA256 hash of the binary for a plugin version.
func (aMap Artifacts) GetDigest(version, os, arch string) (string, error) {
	a, err := aMap.GetArtifact(version, os, arch)
	if err != nil {
		return "", err
	}

	return a.Digest, nil
}

// DescribeArtifact returns the artifact resource based plugin metadata
func (aMap Artifacts) DescribeArtifact(version, os, arch string) (Artifact, error) {
	return aMap.GetArtifact(version, os, arch)
}

// ArtifactFromK8sV1alpha1 returns Artifact from k8sV1alpha1
func ArtifactFromK8sV1alpha1(a cliv1alpha1.Artifact) Artifact { //nolint:gocritic
	return Artifact{
		Image:  a.Image,
		URI:    a.URI,
		Digest: a.Digest,
		OS:     a.OS,
		Arch:   a.Arch,
	}
}

// ArtifactListFromK8sV1alpha1 returns ArtifactList from k8sV1alpha1
func ArtifactListFromK8sV1alpha1(l cliv1alpha1.ArtifactList) ArtifactList {
	aList := make(ArtifactList, len(l))
	for _, a := range l {
		aList = append(aList, ArtifactFromK8sV1alpha1(a))
	}
	return aList
}

// ArtifactsFromK8sV1alpha1 returns Artifacts from k8sV1alpha1
func ArtifactsFromK8sV1alpha1(m map[string]cliv1alpha1.ArtifactList) Artifacts {
	aMap := make(Artifacts, len(m))
	for v, l := range m {
		aMap[v] = ArtifactListFromK8sV1alpha1(l)
	}
	return aMap
}
