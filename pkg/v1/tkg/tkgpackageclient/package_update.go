// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) UpdatePackage(o *tkgpackagedatamodel.PackageInstalledOptions) error {
	pkg, err := p.kappClient.GetPackageInstall(o.PkgInstallName, o.Namespace)
	if err != nil && !k8serror.IsNotFound(err) {
		return err
	}

	if pkg == nil {
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
	if o.Version != pkg.Status.Version {
		// check if user provided version is valid
		if _, _, err := p.GetPackage(pkg.Spec.PackageRef.RefName, o.Version, o.Namespace); err != nil {
			return err
		}

		if pkg.Spec.PackageRef == nil || pkg.Spec.PackageRef.VersionSelection == nil {
			return errors.New(fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
		}

		pkg.Spec.PackageRef.VersionSelection.Constraints = o.Version
		if err = p.kappClient.UpdatePackageInstall(pkg); err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
		}
	}

	return nil
}
