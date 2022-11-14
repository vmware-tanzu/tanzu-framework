// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
)

const (
	tanzuCorePackageRepoName      = "tanzu-core"
	tanzuCorePackageRepoNamespace = "tkg-system"
)

func (c *TkgClient) CleanupCorePackageRepo(clusterClient clusterclient.Client) error {
	found, err := findCorePackageRepo(clusterClient)
	if err != nil {
		return errors.Wrap(err, "unable to find the core package repository")
	}
	if !found {
		log.Info("Core package repository not found, no need to cleanup")
		return nil
	}

	err = deleteCorePackageRepo(clusterClient)
	if err != nil {
		return errors.Wrap(err, "unable to delete the core package repository")
	}
	return nil
}

func findCorePackageRepo(clusterClient clusterclient.Client) (bool, error) {
	packageRepo := &pkgiv1alpha1.PackageRepository{}
	err := clusterClient.GetResource(packageRepo, tanzuCorePackageRepoName, tanzuCorePackageRepoNamespace, nil, nil)
	if err != nil && apierrors.IsNotFound(err) {
		return false, nil
	} else if err != nil {
		return false, errors.Wrap(err, "unable to get PackageRepositories")
	}
	return true, nil
}

func deleteCorePackageRepo(clusterClient clusterclient.Client) error {
	log.Info("Deleting core package repository")
	packageRepo := &pkgiv1alpha1.PackageRepository{}
	packageRepo.Name = tanzuCorePackageRepoName
	packageRepo.Namespace = tanzuCorePackageRepoNamespace
	err := clusterClient.DeleteResource(packageRepo)
	if err != nil && apierrors.IsNotFound(err) {
		return nil
	} else if err != nil {
		return errors.Wrap(err, "failed to delete core package repository")
	}
	return nil
}
