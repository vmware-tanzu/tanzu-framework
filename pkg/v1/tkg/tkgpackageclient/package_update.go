// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) UpdatePackageInstall(o *tkgpackagedatamodel.PackageInstalledOptions) error {
	pkgInstall, err := p.kappClient.GetPackageInstall(o.PkgInstallName, o.Namespace)
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}

	if pkgInstall == nil {
		// package is not installed yet and install flag is present, install the package
		if o.Install {
			if o.PackageName == "" {
				return errors.New("package-name is required when install flag is declared")
			}
			if err := p.InstallPackage(o); err != nil {
				return err
			}
			return nil
		}
		return errors.New(fmt.Sprintf("package '%s' is not among the list of installed packages in namespace '%s'", o.PkgInstallName, o.Namespace))
	}

	// update installed package with a different version
	if o.Version != pkgInstall.Status.Version {
		// check if user provided version is valid
		if _, _, err := p.GetPackage(pkgInstall.Spec.PackageRef.RefName, o.Version, o.Namespace); err != nil {
			return err
		}

		if pkgInstall.Spec.PackageRef == nil || pkgInstall.Spec.PackageRef.VersionSelection == nil {
			return errors.New(fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
		}

		pkgInstallToUpdate := pkgInstall.DeepCopy()
		pkgInstallToUpdate.Spec.PackageRef.VersionSelection.Constraints = o.Version
		if err = p.kappClient.UpdatePackageInstall(pkgInstallToUpdate); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
		}
	}

	return nil
}
