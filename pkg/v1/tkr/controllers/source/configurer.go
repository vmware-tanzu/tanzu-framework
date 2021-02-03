// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package source

import (
	"context"
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/constants"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

const (
	configMapName     = "tkr-controller-config"
	caCertsKey        = "caCerts"
	registryCertsFile = "registry_certs"
)

func (r *reconciler) Configure() error {
	configMap := &corev1.ConfigMap{}
	err := r.client.Get(context.Background(), types.NamespacedName{Namespace: constants.TKRNamespace, Name: configMapName}, configMap)
	// Not configure anything if the ConfigMap is not found
	if k8serr.IsNotFound(err) {
		return nil
	}

	if err != nil {
		return errors.Wrapf(err, "unable to find the ConfigMap %s", configMapName)
	}
	err = addTrustedCerts(configMap.Data[caCertsKey])
	if err != nil {
		return errors.Wrap(err, "failed to add certs")
	}

	return nil
}

func addTrustedCerts(certChain string) (err error) {
	if certChain == "" {
		return nil
	}

	filePath, err := getRegistryCertFile()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filePath, []byte(certChain), 0644)
}
func getRegistryCertFile() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "could not locate local tanzu dir")
	}
	return path.Join(home, registryCertsFile), nil
}
