// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"

	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/packages/v1alpha1"
)

// resolvePackage takes a package name and package version and returns the corresponding package
// and package version resources. If the resolution be unsuccessful, an error is returned.
func (p *pkgClient) resolvePackage(pkgName, pkgVersion, namespace string) (*kapppkg.Package, *kapppkg.PackageVersion, error) {
	var (
		resolvedPackage *kapppkg.Package
		err             error
	)

	if resolvedPackage, err = p.kappClient.GetPackageByName(pkgName, namespace); err != nil {
		return nil, nil, errors.Wrap(err, "failed to find a package with the specified name")
	}

	packageVersionList, err := p.kappClient.ListPackageVersions(pkgName, namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "failed to list package versions")
	}

	for _, item := range packageVersionList.Items { //nolint:gocritic
		if item.Spec.Version == pkgVersion {
			return resolvedPackage, &item, nil
		}
	}

	return nil, nil, errors.Errorf(fmt.Sprintf("failed to resolve package %s %s", pkgName, pkgVersion))
}

// validateRepository ensures that another repository (with the same name or same OCI registry URL) does not already exist in the cluster
func (p *pkgClient) validateRepository(repositoryName, repositoryImg string) error {
	repositoryList, err := p.kappClient.ListPackageRepositories()
	if err != nil {
		return errors.Wrap(err, "failed to list current package repository")
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
