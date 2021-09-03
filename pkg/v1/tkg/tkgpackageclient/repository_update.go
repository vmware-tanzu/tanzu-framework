// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) UpdateRepository(o *tkgpackagedatamodel.RepositoryOptions) error {
	existingRepository, err := p.kappClient.GetPackageRepository(o.RepositoryName, o.Namespace)
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}

	if existingRepository != nil {
		repositoryToUpdate := existingRepository.DeepCopy()
		repositoryToUpdate.Spec.Fetch.ImgpkgBundle.Image = o.RepositoryURL

		if err := p.kappClient.UpdatePackageRepository(repositoryToUpdate); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace))
		}
	} else if o.CreateRepository {
		if err := p.AddRepository(o); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to create package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace))
		}
	} else {
		return errors.Wrap(err, fmt.Sprintf("failed to find package repository '%s' in namespace '%s'", o.RepositoryName, o.Namespace))
	}

	return nil
}
