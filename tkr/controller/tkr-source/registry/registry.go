// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

// Package registry provides the Registry interface and implementation.
package registry

import (
	"context"
	"os"
	"path"
	"sync"

	ctlimg "github.com/k14s/imgpkg/pkg/imgpkg/registry"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/tkr/controller/tkr-source/registry/internal/registry"
)

const (
	configMapName     = "tkr-controller-config"
	caCertsKey        = "caCerts"
	registryCertsFile = "registry_certs"
)

// Registry defines the registry interface
type Registry interface {
	// ListImageTags lists all tags of the given image.
	ListImageTags(imageName string) ([]string, error)
	// GetFile gets the file content bundled in the given image:tag.
	// If filename is empty, it will get the first file.
	GetFile(imageWithTag string, filename string) ([]byte, error)
	// GetFiles get all the files content bundled in the given image:tag.
	GetFiles(imageWithTag string) (map[string][]byte, error)
	// DownloadBundle downloads OCI bundle similar to `imgpkg pull -b` command
	// It is recommended to use this function when downloading imgpkg bundle
	DownloadBundle(imageName, outputDir string) error
}

type impl struct {
	Registry

	Client client.Client
	Config Config

	sync.Once
	initDone chan struct{}
}

var _ Registry = (*impl)(nil)

// New returns a new Registry instance.
func New(c client.Client, config Config) *impl { // nolint:revive // unexported-return: *impl implements a public interface
	return &impl{
		Client:   c,
		Config:   config,
		initDone: make(chan struct{}),
	}
}

// Config contains the controller manager context.
type Config struct {
	TKRNamespace       string
	VerifyRegistryCert bool
}

func (r *impl) SetupWithManager(m ctrl.Manager) error {
	return m.Add(r)
}

func (r *impl) Start(ctx context.Context) error {
	var err error
	r.Do(func() {
		err = r.configure(ctx)
		if err == nil {
			close(r.initDone)
		}
	})
	return err
}

func (r *impl) configure(ctx context.Context) error {
	configMap := &corev1.ConfigMap{}
	if err := r.Client.Get(ctx,
		types.NamespacedName{Namespace: r.Config.TKRNamespace, Name: configMapName},
		configMap); err != nil && !k8serr.IsNotFound(err) {
		return errors.Wrapf(err, "unable to get the ConfigMap %s", configMapName)
	}

	err := addTrustedCerts(configMap.Data[caCertsKey])
	if err != nil {
		return errors.Wrap(err, "failed to add certs")
	}

	registryOps := &ctlimg.Opts{
		VerifyCerts: r.Config.VerifyRegistryCert,
		Anon:        true,
	}

	// Add custom CA cert paths only if VerifyCerts is enabled
	if registryOps.VerifyCerts {
		registryCertPath, err := getRegistryCertFile()
		if err == nil {
			if _, err = os.Stat(registryCertPath); err == nil {
				registryOps.CACertPaths = []string{registryCertPath}
			}
		}
	}

	r.Registry, err = registry.New(registryOps)
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

func (r *impl) ListImageTags(imageName string) ([]string, error) {
	<-r.initDone
	return r.Registry.ListImageTags(imageName)
}

func (r *impl) GetFile(imageWithTag, filename string) ([]byte, error) {
	<-r.initDone
	return r.Registry.GetFile(imageWithTag, filename)
}

func (r *impl) GetFiles(imageWithTag string) (map[string][]byte, error) {
	<-r.initDone
	return r.Registry.GetFiles(imageWithTag)
}

func (r *impl) DownloadBundle(imageName, outputDir string) error {
	<-r.initDone
	return r.Registry.DownloadBundle(imageName, outputDir)
}
