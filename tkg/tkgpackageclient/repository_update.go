// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"

	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	versions "github.com/vmware-tanzu/carvel-vendir/pkg/vendir/versions/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions, progress *tkgpackagedatamodel.PackageProgress, operationType tkgpackagedatamodel.OperationType) {
	p.updateRepository(o, progress, operationType)
}

func (p *pkgClient) UpdateRepositorySync(o *tkgpackagedatamodel.RepositoryOptions, operationType tkgpackagedatamodel.OperationType) error {
	pp := newPackageProgress()

	go p.updateRepository(o, pp, operationType)

	initialMsg := fmt.Sprintf("Updating package repository '%s'", o.RepositoryName)
	if err := DisplayProgress(initialMsg, pp); err != nil {
		if err.Error() == tkgpackagedatamodel.ErrRepoNotExists {
			log.Warningf("package repository '%s' does not exist in namespace '%s'", o.RepositoryName, o.Namespace)
			return nil
		}
		return err
	}
	log.Infof("Updated package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace)
	return nil
}

func (p *pkgClient) updateRepository(o *tkgpackagedatamodel.RepositoryOptions, progress *tkgpackagedatamodel.PackageProgress, operationType tkgpackagedatamodel.OperationType) {
	var (
		existingRepository *kappipkg.PackageRepository
		err                error
		tag                string
	)

	defer func() {
		if err != nil {
			progress.Err <- err
		}
		if operationType == tkgpackagedatamodel.OperationTypeUpdate {
			close(progress.ProgressMsg)
			close(progress.Done)
		}
	}()

	progress.ProgressMsg <- fmt.Sprintf("Getting package repository '%s'", o.RepositoryName)
	existingRepository, err = p.kappClient.GetPackageRepository(o.RepositoryName, o.Namespace)
	if err != nil {
		if k8serror.IsNotFound(err) {
			err = nil
		} else {
			return
		}
	}

	if existingRepository != nil {
		repositoryToUpdate := existingRepository.DeepCopy()
		progress.ProgressMsg <- "Validating provided settings for the package repository"
		if err = p.validateRepositoryUpdate(o.RepositoryName, o.RepositoryURL, o.Namespace); err != nil {
			return
		}

		_, tag, err = ParseRegistryImageURL(o.RepositoryURL)
		if err != nil {
			err = errors.Wrap(err, "failed to parse OCI registry URL")
			return
		}

		repositoryToUpdate.Spec = kappipkg.PackageRepositorySpec{
			Fetch: &kappipkg.PackageRepositoryFetch{
				ImgpkgBundle: &kappctrl.AppFetchImgpkgBundle{Image: o.RepositoryURL},
			},
		}

		if tag == "" {
			repositoryToUpdate.Spec.Fetch.ImgpkgBundle.TagSelection = &versions.VersionSelection{
				Semver: &versions.VersionSelectionSemver{
					Constraints: tkgpackagedatamodel.DefaultRepositoryImageTagConstraint,
				},
			}
		}

		progress.ProgressMsg <- "Updating package repository resource"
		if err = p.kappClient.UpdatePackageRepository(repositoryToUpdate); err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to update package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace))
			return
		}

		if o.Wait {
			if err = p.waitForResourceInstallation(o.RepositoryName, o.Namespace, o.PollInterval, o.PollTimeout, progress.ProgressMsg, tkgpackagedatamodel.ResourceTypePackageRepository); err != nil {
				return
			}
		}
	} else if o.CreateRepository {
		p.AddRepository(o, progress, tkgpackagedatamodel.OperationTypeUpdate)
	} else {
		err = &tkgpackagedatamodel.PackagePluginNonCriticalError{Reason: tkgpackagedatamodel.ErrRepoNotExists}
	}
}

// validateRepositoryUpdate ensures that another repository (with the same OCI registry URL) does not already exist in the cluster
func (p *pkgClient) validateRepositoryUpdate(repositoryName, repositoryImg, namespace string) error {
	repositoryList, err := p.kappClient.ListPackageRepositories(namespace)
	if err != nil {
		return errors.Wrap(err, "failed to list package repositories")
	}

	for _, repository := range repositoryList.Items { //nolint:gocritic
		// This stops the update validation to compare with itself
		if repository.Name == repositoryName {
			continue
		}

		if repository.Spec.Fetch != nil && repository.Spec.Fetch.ImgpkgBundle != nil &&
			repository.Spec.Fetch.ImgpkgBundle.Image == repositoryImg {
			return errors.New("repository with the same OCI registry URL already exists")
		}
	}

	return nil
}
