// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"fmt"

	"github.com/pkg/errors"
	k8serror "k8s.io/apimachinery/pkg/api/errors"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) UpdatePackage(o *tkgpackagedatamodel.PackageOptions, progress *tkgpackagedatamodel.PackageProgress) {
	var (
		pkgInstall *kappipkg.PackageInstall
		err        error
	)

	defer func() {
		packageProgressCleanup(err, progress)
	}()

	progress.ProgressMsg <- fmt.Sprintf("Getting package install for '%s'", o.PkgInstallName)
	pkgInstall, err = p.kappClient.GetPackageInstall(o.PkgInstallName, o.Namespace)
	if err != nil {
		if k8serror.IsNotFound(err) {
			err = nil
		} else {
			return
		}
	}

	if pkgInstall == nil {
		if !o.Install {
			err = errors.New(fmt.Sprintf("package '%s' is not among the list of installed packages in namespace '%s'", o.PkgInstallName, o.Namespace))
			return
		}
		if o.PackageName == "" {
			err = errors.New("package-name is required when install flag is declared")
			return
		}
		progress.ProgressMsg <- fmt.Sprintf("Installing package '%s'", o.PkgInstallName)
		p.InstallPackage(o, progress, true)
	} else if pkgInstall != nil && o.Version != pkgInstall.Status.Version {
		if pkgInstall.Spec.PackageRef == nil || pkgInstall.Spec.PackageRef.VersionSelection == nil {
			err = errors.New(fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
			return
		}
		progress.ProgressMsg <- fmt.Sprintf("Getting package metadata for '%s'", pkgInstall.Spec.PackageRef.RefName)
		o.PackageName = pkgInstall.Spec.PackageRef.RefName
		if _, _, err = p.GetPackage(o); err != nil {
			return
		}
		pkgInstallToUpdate := pkgInstall.DeepCopy()
		pkgInstallToUpdate.Spec.PackageRef.VersionSelection.Constraints = o.Version
		progress.ProgressMsg <- fmt.Sprintf("Updating package install for '%s'", o.PkgInstallName)
		if err = p.kappClient.UpdatePackageInstall(pkgInstallToUpdate); err != nil {
			err = errors.Wrap(err, fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
			return
		}
	}

	progress.Success <- true
}
