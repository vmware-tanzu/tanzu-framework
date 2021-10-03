// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgpackageclient

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/tkgpackagedatamodel"
)

func (p *pkgClient) UpdatePackage(o *tkgpackagedatamodel.PackageOptions, progress *tkgpackagedatamodel.PackageProgress) {
	var (
		pkgInstall    *kappipkg.PackageInstall
		err           error
		secretCreated bool
	)

	defer func() {
		progressCleanup(err, progress)
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
			err = &tkgpackagedatamodel.PackagePluginNonCriticalError{Reason: tkgpackagedatamodel.ErrPackageNotInstalled}
			return
		}
		if o.PackageName == "" {
			err = errors.New("package-name is required when install flag is declared")
			return
		}
		progress.ProgressMsg <- fmt.Sprintf("Installing package '%s'", o.PkgInstallName)
		p.InstallPackage(o, progress, tkgpackagedatamodel.OperationTypeUpdate)
		return
	}

	pkgInstallToUpdate := pkgInstall.DeepCopy()
	if o.Version != pkgInstallToUpdate.Status.Version {
		if pkgInstallToUpdate.Spec.PackageRef == nil || pkgInstallToUpdate.Spec.PackageRef.VersionSelection == nil {
			err = errors.New(fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
			return
		}
		progress.ProgressMsg <- fmt.Sprintf("Getting package metadata for '%s'", pkgInstallToUpdate.Spec.PackageRef.RefName)
		o.PackageName = pkgInstallToUpdate.Spec.PackageRef.RefName
		if _, _, err = p.GetPackage(o); err != nil {
			return
		}
		pkgInstallToUpdate.Spec.PackageRef.VersionSelection.Constraints = o.Version
	}

	if o.ValuesFile != "" {
		if secretCreated, err = p.updateValuesFile(o, pkgInstallToUpdate, progress.ProgressMsg); err != nil {
			return
		}

		pkgInstallToUpdate.Spec.Values = []kappipkg.PackageInstallValues{
			{
				SecretRef: &kappipkg.PackageInstallValuesSecretRef{
					Name: fmt.Sprintf(tkgpackagedatamodel.SecretName, o.PkgInstallName, o.Namespace),
				},
			},
		}
	}

	progress.ProgressMsg <- fmt.Sprintf("Updating package install for '%s'", o.PkgInstallName)
	if err = p.kappClient.UpdatePackageInstall(pkgInstallToUpdate, secretCreated); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
		return
	}

	if o.Wait {
		if err = p.waitForResourceInstallation(o.PkgInstallName, o.Namespace, o.PollInterval, o.PollTimeout, progress.ProgressMsg, tkgpackagedatamodel.ResourceTypePackageInstall); err != nil {
			return
		}
	}
}

// updateValuesFile either creates or updates the values secret depending on whether the corresponding annotation exist or not
func (p *pkgClient) updateValuesFile(o *tkgpackagedatamodel.PackageOptions, pkgInstallToUpdate *kappipkg.PackageInstall, progress chan string) (bool, error) {
	var (
		secretCreated bool
		err           error
	)

	o.SecretName = fmt.Sprintf(tkgpackagedatamodel.SecretName, o.PkgInstallName, o.Namespace)

	if o.SecretName == pkgInstallToUpdate.GetAnnotations()[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-Secret"] {
		progress <- fmt.Sprintf("Updating secret '%s'", o.SecretName)
		if err = p.updateDataValuesSecret(o); err != nil {
			err = errors.Wrap(err, "failed to update secret based on values file")
			return false, err
		}
	} else {
		progress <- fmt.Sprintf("Creating secret '%s'", o.SecretName)
		if secretCreated, err = p.createDataValuesSecret(o); err != nil {
			return secretCreated, errors.Wrap(err, "failed to create secret based on values file")
		}
	}

	return secretCreated, nil
}

// updateDataValuesSecret update a secret object containing the user-provided configuration.
func (p *pkgClient) updateDataValuesSecret(o *tkgpackagedatamodel.PackageOptions) error {
	var err error
	dataValues := make(map[string][]byte)

	if dataValues[filepath.Base(o.ValuesFile)], err = ioutil.ReadFile(o.ValuesFile); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to read from data values file '%s'", o.ValuesFile))
	}
	secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: o.SecretName, Namespace: o.Namespace}, Data: dataValues,
	}

	if err := p.kappClient.GetClient().Update(context.Background(), secret); err != nil {
		return errors.Wrap(err, "failed to update Secret resource")
	}

	return nil
}
