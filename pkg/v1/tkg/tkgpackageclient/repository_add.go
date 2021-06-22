// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

// AddRepository validates the provided input and adds the package repository CR to the cluster
func (p *pkgClient) AddRepository(o *tkgpackagedatamodel.RepositoryOptions) error {
	if err := p.validateRepository(o.RepositoryName, o.RepositoryURL, o.Namespace); err != nil {
		return err
	}

	if o.CreateNamespace {
		if err := p.createNamespace(o.Namespace); err != nil {
			return err
		}
	}

	newPackageRepo := p.newPackageRepository(o.RepositoryName, o.RepositoryURL, o.Namespace)

	if err := p.kappClient.CreatePackageRepository(newPackageRepo); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to create package repository '%s'", o.RepositoryName))
	}

	return nil
}

// newPackageRepository creates a new instance of the PackageRepository object
func (p *pkgClient) newPackageRepository(repositoryName, repositoryImg, namespace string) *kappipkg.PackageRepository {
	return &kappipkg.PackageRepository{
		TypeMeta:   metav1.TypeMeta{APIVersion: tkgpackagedatamodel.DefaultAPIVersion, Kind: tkgpackagedatamodel.KindPackageRepository},
		ObjectMeta: metav1.ObjectMeta{Name: repositoryName, Namespace: namespace},
		Spec: kappipkg.PackageRepositorySpec{Fetch: &kappipkg.PackageRepositoryFetch{
			ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: repositoryImg},
		}},
	}
}
