// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) DeleteRepository(o *tkgpackagedatamodel.RepositoryOptions, progress *tkgpackagedatamodel.PackageProgress) {
	p.deleteRepository(o, progress)
}

func (p *pkgClient) DeleteRepositorySync(o *tkgpackagedatamodel.RepositoryOptions) error {
	pp := newPackageProgress()

	go p.deleteRepository(o, pp)

	initialMsg := fmt.Sprintf("Deleting package repository '%s'", o.RepositoryName)
	if err := DisplayProgress(initialMsg, pp); err != nil {
		if err.Error() == tkgpackagedatamodel.ErrRepoNotExists {
			log.Warningf("package repository '%s' does not exist in namespace '%s'", o.RepositoryName, o.Namespace)
			return nil
		}
		return err
	}
	log.Infof("Deleted package repository '%s' from namespace '%s'", o.RepositoryName, o.Namespace)
	return nil
}

func (p *pkgClient) deleteRepository(o *tkgpackagedatamodel.RepositoryOptions, progress *tkgpackagedatamodel.PackageProgress) {
	var (
		packageRepo *kappipkg.PackageRepository
		err         error
	)

	defer func() {
		progressCleanup(err, progress)
	}()

	progress.ProgressMsg <- fmt.Sprintf("Getting package repository '%s'", o.RepositoryName)
	packageRepo, err = p.kappClient.GetPackageRepository(o.RepositoryName, o.Namespace)
	if err != nil {
		if apierrors.IsNotFound(err) {
			err = &tkgpackagedatamodel.PackagePluginNonCriticalError{Reason: tkgpackagedatamodel.ErrRepoNotExists}
		}
		return
	}

	progress.ProgressMsg <- "Deleting package repository resource"
	err = p.kappClient.DeletePackageRepository(packageRepo)
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to delete package repository '%s' from namespace '%s'", o.RepositoryName, o.Namespace))
	}

	if o.Wait {
		if err = p.waitForResourceDeletion(o.RepositoryName, o.Namespace, o.PollInterval, o.PollTimeout, progress.ProgressMsg, tkgpackagedatamodel.ResourceTypePackageRepository); err != nil {
			return
		}
	}
}
