// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aunum/log"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serror "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

func (p *pkgClient) UpdatePackage(o *packagedatamodel.PackageOptions, progress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType) {
	p.updatePackage(o, progress, operationType)
}

// UpdatePackageSync installs/updates the package and returns an error if any
func (p *pkgClient) UpdatePackageSync(o *packagedatamodel.PackageOptions, operationType packagedatamodel.OperationType) error {
	pp := newPackageProgress()

	go p.updatePackage(o, pp, operationType)

	initialMsg := fmt.Sprintf("Updating installed package '%s'", o.PkgInstallName)
	if err := DisplayProgress(initialMsg, pp); err != nil {
		if err.Error() == packagedatamodel.ErrPackageNotInstalled {
			log.Warningf("package '%s' is not among the list of installed packages in namespace '%s'", o.PkgInstallName, o.Namespace)
			return nil
		}
		return err
	}
	log.Infof("%s", fmt.Sprintf("Updated installed package '%s' in namespace '%s'", o.PkgInstallName, o.Namespace))
	return nil
}

func (p *pkgClient) updatePackage(o *packagedatamodel.PackageOptions, progress *packagedatamodel.PackageProgress, operationType packagedatamodel.OperationType) {
	var (
		pkgInstall                      *kappipkg.PackageInstall
		pkgInstallToUpdate              *kappipkg.PackageInstall
		pkgPluginResourceCreationStatus packagedatamodel.PkgPluginResourceCreationStatus
		err                             error
		changed                         bool
	)

	defer func() {
		if err != nil {
			progress.Err <- err
		}
		if operationType == packagedatamodel.OperationTypeUpdate {
			close(progress.ProgressMsg)
			close(progress.Done)
		}
	}()

	progress.ProgressMsg <- fmt.Sprintf("Getting package install for '%s'", o.PkgInstallName)
	if pkgInstall, err = p.kappClient.GetPackageInstall(o.PkgInstallName, o.Namespace); err != nil {
		if !k8serror.IsNotFound(err) {
			return
		}
		err = nil
	}

	if pkgInstall == nil {
		if !o.Install {
			err = &packagedatamodel.PackagePluginNonCriticalError{Reason: packagedatamodel.ErrPackageNotInstalled}
			return
		}
		progress.ProgressMsg <- fmt.Sprintf("Installing package '%s'", o.PkgInstallName)
		p.InstallPackage(o, progress, packagedatamodel.OperationTypeUpdate)
		return
	}

	if pkgInstallToUpdate, changed, err = p.preparePackageInstallForUpdate(o, pkgInstall); err != nil {
		return
	}

	progress.ProgressMsg <- fmt.Sprintf("Getting package metadata for '%s'", pkgInstallToUpdate.Spec.PackageRef.RefName)
	if _, _, err = p.GetPackage(o); err != nil {
		return
	}

	if pkgPluginResourceCreationStatus.IsSecretCreated, err = p.createOrUpdateValuesSecret(o, pkgInstallToUpdate, progress.ProgressMsg); err != nil {
		return
	}

	if o.ValuesFile == "" && !changed {
		return
	}

	progress.ProgressMsg <- fmt.Sprintf("Updating package install for '%s'", o.PkgInstallName)
	if err = p.kappClient.UpdatePackageInstall(pkgInstallToUpdate, &pkgPluginResourceCreationStatus); err != nil {
		err = errors.Wrap(err, fmt.Sprintf("failed to update package '%s'", o.PkgInstallName))
		return
	}

	if o.Wait {
		if err = p.waitForResourceInstallation(o.PkgInstallName, o.Namespace, o.PollInterval, o.PollTimeout, progress.ProgressMsg, packagedatamodel.ResourceTypePackageInstall); err != nil {
			return
		}
	}
}

func (p *pkgClient) preparePackageInstallForUpdate(o *packagedatamodel.PackageOptions, pkgInstall *kappipkg.PackageInstall) (*kappipkg.PackageInstall, bool, error) {
	var changed bool

	if err := p.validateValuesFile(o); err != nil {
		return nil, false, err
	}

	pkgInstallToUpdate := pkgInstall.DeepCopy()

	if pkgInstallToUpdate.Spec.PackageRef == nil || pkgInstallToUpdate.Spec.PackageRef.VersionSelection == nil {
		err := fmt.Errorf("failed to update package '%s' as no existing package reference/version was found in the package install", o.PkgInstallName)
		return nil, false, err
	}

	// If o.PackageName is provided by the user (via --package-name flag), verify that the package name in PackageInstall matches it.
	// This will prevent the users from accidentally overwriting an installed package with another package content due to choosing a pre-existing name for the package isntall.
	// Otherwise if o.PackageName is not provided, fill it from the installed package spec, as the validation logic in GetPackage() needs this field to be set.
	if o.PackageName != "" && pkgInstallToUpdate.Spec.PackageRef.RefName != o.PackageName {
		err := fmt.Errorf("installed package '%s' is already associated with package '%s'", o.PkgInstallName, pkgInstallToUpdate.Spec.PackageRef.RefName)
		return nil, false, err
	}
	o.PackageName = pkgInstallToUpdate.Spec.PackageRef.RefName

	// If o.Version is provided by the user (via --version flag), set the version in PackageInstall to this version
	// Otherwise if o.Version is not provided, fill it from the installed package spec, as the validation logic in GetPackage() needs this field to be set.
	if o.Version != "" {
		if pkgInstallToUpdate.Spec.PackageRef.VersionSelection.Constraints != o.Version {
			changed = true
			pkgInstallToUpdate.Spec.PackageRef.VersionSelection.Constraints = o.Version
		}
	} else {
		o.Version = pkgInstallToUpdate.Spec.PackageRef.VersionSelection.Constraints
	}

	return pkgInstallToUpdate, changed, nil
}

// createOrUpdateValuesSecret either creates or updates the values secret depending on whether the corresponding annotation exists or not
func (p *pkgClient) createOrUpdateValuesSecret(o *packagedatamodel.PackageOptions, pkgInstallToUpdate *kappipkg.PackageInstall, progress chan string) (bool, error) {
	var (
		secretCreated bool
		err           error
	)

	if o.ValuesFile == "" {
		return false, nil
	}

	o.SecretName = fmt.Sprintf(packagedatamodel.SecretName, o.PkgInstallName, o.Namespace)

	if o.SecretName == pkgInstallToUpdate.GetAnnotations()[packagedatamodel.TanzuPkgPluginAnnotation+"-Secret"] {
		progress <- fmt.Sprintf("Updating secret '%s'", o.SecretName)
		if err = p.updateDataValuesSecret(o); err != nil {
			err = errors.Wrap(err, "failed to update secret based on values file")
			return false, err
		}
	} else {
		progress <- fmt.Sprintf("Creating secret '%s'", o.SecretName)
		if secretCreated, err = p.createOrUpdateDataValuesSecret(o); err != nil {
			return secretCreated, errors.Wrap(err, "failed to create secret based on values file")
		}
	}

	pkgInstallToUpdate.Spec.Values = []kappipkg.PackageInstallValues{
		{SecretRef: &kappipkg.PackageInstallValuesSecretRef{Name: o.SecretName}},
	}

	return secretCreated, nil
}

// updateDataValuesSecret update a secret object containing the user-provided configuration.
func (p *pkgClient) updateDataValuesSecret(o *packagedatamodel.PackageOptions) error {
	var err error
	dataValues := make(map[string][]byte)

	if dataValues[filepath.Base(o.ValuesFile)], err = os.ReadFile(o.ValuesFile); err != nil {
		return errors.Wrap(err, fmt.Sprintf("failed to read from data values file '%s'", o.ValuesFile))
	}
	Secret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: o.SecretName, Namespace: o.Namespace}, Data: dataValues,
	}

	if err := p.kappClient.GetClient().Update(context.Background(), Secret); err != nil {
		return errors.Wrap(err, "failed to update Secret resource")
	}

	return nil
}
