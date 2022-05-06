// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package fetcher

import (
	"context"
	"os"
	"path"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/registry"
)

const (
	configMapName     = "tkr-controller-config"
	caCertsKey        = "caCerts"
	registryCertsFile = "registry_certs"
)

func (f *Fetcher) configure(ctx context.Context) error {
	configMap := &corev1.ConfigMap{}
	if err := f.Client.Get(ctx,
		types.NamespacedName{Namespace: f.Config.TKRNamespace, Name: configMapName},
		configMap); !k8serr.IsNotFound(err) {
		return errors.Wrapf(err, "unable to get the ConfigMap %s", configMapName)
	}

	err := addTrustedCerts(configMap.Data[caCertsKey])
	if err != nil {
		return errors.Wrap(err, "failed to add certs")
	}

	f.registryOps = ctlimg.Opts{
		VerifyCerts: f.Config.VerifyRegistryCert,
		Anon:        true,
	}

	// Add custom CA cert paths only if VerifyCerts is enabled
	if f.registryOps.VerifyCerts {
		registryCertPath, err := getRegistryCertFile()
		if err == nil {
			if _, err = os.Stat(registryCertPath); err == nil {
				f.registryOps.CACertPaths = []string{registryCertPath}
			}
		}
	}

	f.registry, err = registry.New(&f.registryOps)
	return err
}

func addTrustedCerts(certChain string) (err error) {
	if certChain == "" {
		return nil
	}

	filePath, err := getRegistryCertFile()
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, []byte(certChain), 0644)
}
func getRegistryCertFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "could not locate local tanzu dir")
	}
	return path.Join(home, registryCertsFile), nil
}
