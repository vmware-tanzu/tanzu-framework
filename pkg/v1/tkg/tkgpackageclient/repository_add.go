// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

const (
	msgRunPackageRepositoryUpdate = "\n\nPlease consider using 'tanzu package repository update' to update the package repository with correct settings\n"
)

// AddRepository validates the provided input and adds the package repository CR to the cluster
func (p *pkgClient) AddRepository(o *tkgpackagedatamodel.RepositoryOptions, progress *tkgpackagedatamodel.PackageProgress, operationType tkgpackagedatamodel.OperationType) {
	var err error

	defer func() {
		if err != nil {
			progress.Err <- err
		}
		if operationType == tkgpackagedatamodel.OperationTypeInstall {
			close(progress.ProgressMsg)
			close(progress.Done)
		}
	}()

	progress.ProgressMsg <- "Validating provided settings for the package repository"
	if err = p.validateRepository(o.RepositoryName, o.RepositoryURL, o.Namespace); err != nil {
		return
	}

	_, tag, err := ParseImageUrl(o.RepositoryURL)
	if err != nil {
		return errors.Wrap(err, "invalid repository image URL")
	}

	if o.CreateNamespace {
		progress.ProgressMsg <- fmt.Sprintf("Creating namespace '%s'", o.Namespace)
		if err = p.createNamespace(o.Namespace); err != nil {
			return
		}
	}

	newPackageRepo, err := p.newPackageRepository(o.RepositoryName, o.RepositoryURL, tag, o.Namespace)
	if err != nil{
		return err
	}

	progress.ProgressMsg <- "Creating package repository resource"

	if err = p.kappClient.CreatePackageRepository(newPackageRepo); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to create package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace))
		return
	}

	if o.Wait {
		if err = p.waitForResourceInstallation(o.RepositoryName, o.Namespace, o.PollInterval, o.PollTimeout, progress.ProgressMsg, tkgpackagedatamodel.ResourceTypePackageRepository); err != nil {
			log.Warning(msgRunPackageRepositoryUpdate)
			return
		}
	}
}

// newPackageRepository creates a new instance of the PackageRepository object
func (p *pkgClient) newPackageRepository(repositoryName, repositoryImg, tag, namespace string) (*kappipkg.PackageRepository, error) {
	pkgr := &kappipkg.PackageRepository{
		TypeMeta:   metav1.TypeMeta{APIVersion: tkgpackagedatamodel.DefaultAPIVersion, Kind: tkgpackagedatamodel.KindPackageRepository},
		ObjectMeta: metav1.ObjectMeta{Name: repositoryName, Namespace: namespace},
		Spec: kappipkg.PackageRepositorySpec{Fetch: &kappipkg.PackageRepositoryFetch{
			ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: repositoryImg},
		}},
	}

	found, err := checkPackageRepositoryTagSelection()
	if err != nil {
		return nil, errors.Wrap(err, "failed to check package repository resource version")
	}

	if tag == "" && found  {
		pkgr.Spec.Fetch.ImgpkgBundle.TagSelection = &versions.VersionSelection{
			Semver: &versions.VersionSelectionSemver{
				Constraints: defaultImageTagConstraint,
			},
		}
	}
	return pkgr, nil
}

// validateRepository ensures that another repository (with the same name or same OCI registry URL) does not already exist in the cluster
func (p *pkgClient) validateRepository(repositoryName, repositoryImg, namespace string) error {
	repositoryList, err := p.kappClient.ListPackageRepositories(namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list package repositories")
	}

	for _, repository := range repositoryList.Items { //nolint:gocritic
		if repository.Name == repositoryName {
			return errors.New(fmt.Sprintf("package repository name '%s' already exists in namespace '%s'", repositoryName, namespace))
		}

		if repository.Spec.Fetch != nil && repository.Spec.Fetch.ImgpkgBundle != nil &&
			repository.Spec.Fetch.ImgpkgBundle.Image == repositoryImg {
			return errors.New(fmt.Sprintf("package repository URL '%s' already exists in namespace '%s'", repositoryImg, namespace))
		}
	}

	return nil
}
