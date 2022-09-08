// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package packageclient

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	secretgenctrl "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/packageclients/pkg/packagedatamodel"
)

// ListRegistrySecrets lists all registry secrets of type kubernetes.io/dockerconfigjson across the cluster.
func (p *pkgClient) ListRegistrySecrets(o *packagedatamodel.RegistrySecretOptions) (*corev1.SecretList, error) {
	registrySecretList, err := p.kappClient.ListRegistrySecrets(o.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list existing registry secrets in the cluster")
	}

	return registrySecretList, nil
}

// ListSecretExports lists all SecretExports across the cluster.
func (p *pkgClient) ListSecretExports(o *packagedatamodel.RegistrySecretOptions) (*secretgenctrl.SecretExportList, error) {
	secretExportList, err := p.kappClient.ListSecretExports(o.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to list existing secret exports in the cluster")
	}

	return secretExportList, nil
}

// GetSecretExport gets the SecretExport having the same name and namespace as specified secret.
func (p *pkgClient) GetSecretExport(o *packagedatamodel.RegistrySecretOptions) (*secretgenctrl.SecretExport, error) {
	var err error
	SecretExport, err = p.kappClient.GetSecretExport(o.SecretName, o.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get existing secret export in the cluster")
	}

	return SecretExport, nil
}
