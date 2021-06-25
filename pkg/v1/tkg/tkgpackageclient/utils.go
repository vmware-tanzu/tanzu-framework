// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"

	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
)

// GetPackage takes a package name and package version and returns the corresponding PackageMetadata and Package.
// If the resolution is unsuccessful, an error is returned.
func (p *pkgClient) GetPackage(pkgName, pkgVersion, namespace string) (*kapppkg.PackageMetadata, *kapppkg.Package, error) {
	var (
		resolvedPackage *kapppkg.PackageMetadata
		err             error
	)

	if resolvedPackage, err = p.kappClient.GetPackageMetadataByName(pkgName, namespace); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to find a package with name '%s'", pkgName)
	}

	packageVersions, err := p.kappClient.ListPackages(pkgName, namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list package versions")
	}

	for _, item := range packageVersions.Items { //nolint:gocritic
		if item.Spec.Version == pkgVersion {
			return resolvedPackage, &item, nil
		}
	}

	return nil, nil, errors.Errorf(fmt.Sprintf("failed to resolve package '%s' with version '%s'", pkgName, pkgVersion))
}

// validateRepository ensures that another repository (with the same name or same OCI registry URL) does not already exist in the cluster
func (p *pkgClient) validateRepository(repositoryName, repositoryImg, namespace string) error {
	repositoryList, err := p.kappClient.ListPackageRepositories(namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list package repositories")
	}

	for _, repository := range repositoryList.Items { //nolint:gocritic
		if repository.Name == repositoryName {
			return errors.New("repository with the same name already exists")
		}

		if repository.Spec.Fetch != nil && repository.Spec.Fetch.ImgpkgBundle != nil &&
			repository.Spec.Fetch.ImgpkgBundle.Image == repositoryImg {
			return errors.New("repository with the same OCI registry URL already exists")
		}
	}

	return nil
}
