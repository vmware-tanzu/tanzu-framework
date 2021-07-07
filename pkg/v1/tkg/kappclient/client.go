// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package kappclient

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	crtclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	k8sconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"

	"github.com/vmware-tanzu-private/core/pkg/v1/tkg/tkgpackagedatamodel"
)

type client struct {
	client crtclient.Client
}

// GetClient returns k8s client filed of the client
func (c *client) GetClient() crtclient.Client {
	return c.client
}

// NewKappClient returns a new kapp client
func NewKappClient(kubeCfgPath string) (Client, error) {
	var (
		restConfig *rest.Config
		err        error
	)

	scheme := runtime.NewScheme()
	if err := corev1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := rbacv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := kappipkg.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := kapppkg.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := kappctrl.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if kubeCfgPath == "" {
		if restConfig, err = k8sconfig.GetConfig(); err != nil {
			return nil, err
		}
	} else {
		config, err := clientcmd.LoadFromFile(kubeCfgPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load kubeconfig from %s", kubeCfgPath)
		}
		rawConfig, err := clientcmd.Write(*config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to write config")
		}
		if restConfig, err = clientcmd.RESTConfigFromKubeConfig(rawConfig); err != nil {
			return nil, errors.Wrap(err, "Unable to set up rest config")
		}
	}

	mapper, err := apiutil.NewDynamicRESTMapper(restConfig, apiutil.WithLazyDiscovery)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to set up rest mapper")
	}
	crtClient, err := crtclient.New(restConfig, crtclient.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return nil, errors.Wrap(err, "Unable to create cluster client")
	}
	return &client{client: crtClient}, nil
}

func (c *client) addAnnotations(meta *v1.ObjectMeta, isPkgPluginCreatedSvcAccount, isPkgPluginCreatedSecret bool) {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if isPkgPluginCreatedSvcAccount {
		meta.Annotations[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-"+tkgpackagedatamodel.KindClusterRole] = fmt.Sprintf(tkgpackagedatamodel.ClusterRoleName, meta.Name, meta.Namespace)
		meta.Annotations[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-"+tkgpackagedatamodel.KindClusterRoleBinding] = fmt.Sprintf(tkgpackagedatamodel.ClusterRoleBindingName, meta.Name, meta.Namespace)
		meta.Annotations[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-"+tkgpackagedatamodel.KindServiceAccount] = fmt.Sprintf(tkgpackagedatamodel.ServiceAccountName, meta.Name, meta.Namespace)
	}
	if isPkgPluginCreatedSecret {
		meta.Annotations[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-"+tkgpackagedatamodel.KindSecret] = fmt.Sprintf(tkgpackagedatamodel.SecretName, meta.Name, meta.Namespace)
	}
}

// CreatePackageRepository creates a PackageRepository CR
func (c *client) CreatePackageRepository(repository *kappipkg.PackageRepository) error {
	if err := c.client.Create(context.Background(), repository); err != nil {
		return err
	}

	return nil
}

// DeletePackageRepository deletes the provided PackageRepository CR
func (c *client) DeletePackageRepository(repository *kappipkg.PackageRepository) error {
	if err := c.client.Delete(context.Background(), repository); err != nil {
		return err
	}

	return nil
}

// CreatePackageInstall creates a PackageInstall CR
func (c *client) CreatePackageInstall(packageInstall *kappipkg.PackageInstall, isPkgPluginCreatedSvcAccount, isPkgPluginCreatedSecret bool) error {
	installedPkg := packageInstall.DeepCopy()
	c.addAnnotations(&installedPkg.ObjectMeta, isPkgPluginCreatedSvcAccount, isPkgPluginCreatedSecret)

	if err := c.client.Create(context.Background(), installedPkg); err != nil {
		return err
	}

	return nil
}

// GetAppCR gets the App CR
func (c *client) GetAppCR(appName, namespace string) (*kappctrl.App, error) {
	app := &kappctrl.App{}
	if err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: appName, Namespace: namespace}, app); err != nil {
		return nil, err
	}

	return app, nil
}

// GetPackageInstall gets the PackageInstall CR for the provided package name
func (c *client) GetPackageInstall(packageInstallName, namespace string) (*kappipkg.PackageInstall, error) {
	installedPkg := &kappipkg.PackageInstall{}
	if err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: packageInstallName, Namespace: namespace}, installedPkg); err != nil {
		return nil, err
	}

	return installedPkg, nil
}

// GetPackageMetadataByName gets the package with the specified name
func (c *client) GetPackageMetadataByName(packageName, namespace string) (*kapppkg.PackageMetadata, error) {
	pkg := &kapppkg.PackageMetadata{}
	if err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: packageName, Namespace: namespace}, pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

// GetPackageRepository gets the PackageRepository CR
func (c *client) GetPackageRepository(repositoryName, namespace string) (*kappipkg.PackageRepository, error) {
	repository := &kappipkg.PackageRepository{}
	err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: repositoryName, Namespace: namespace}, repository)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

// ListPackageRepositories gets the list of PackageRepository CRs
func (c *client) ListPackageRepositories(namespace string) (*kappipkg.PackageRepositoryList, error) {
	var selectors []crtclient.ListOption
	repositoryList := &kappipkg.PackageRepositoryList{}

	selectors = []crtclient.ListOption{crtclient.InNamespace(namespace)}

	err := c.client.List(context.Background(), repositoryList, selectors...)
	if err != nil {
		return nil, err
	}
	return repositoryList, nil
}

// ListPackageMetadata gets the list of PackageMetadata CRs
func (c *client) ListPackageMetadata(namespace string) (*kapppkg.PackageMetadataList, error) {
	var selectors []crtclient.ListOption
	packageList := &kapppkg.PackageMetadataList{}

	selectors = []crtclient.ListOption{crtclient.InNamespace(namespace)}

	err := c.client.List(context.Background(), packageList, selectors...)
	if err != nil {
		return nil, err
	}
	return packageList, nil
}

// ListPackageInstalls gets the list of PackageInstall CR in the specified namespace.
// If no namespace be provided, it returns the list of installed packages across all namespaces
func (c *client) ListPackageInstalls(namespace string) (*kappipkg.PackageInstallList, error) {
	var selectors []crtclient.ListOption
	packageInstallList := &kappipkg.PackageInstallList{}

	selectors = []crtclient.ListOption{crtclient.InNamespace(namespace)}

	if err := c.client.List(context.Background(), packageInstallList, selectors...); err != nil {
		return nil, err
	}

	return packageInstallList, nil
}

// ListPackages gets the list of Package CRs
func (c *client) ListPackages(packageName, namespace string) (*kapppkg.PackageList, error) {
	var selectors []crtclient.ListOption
	packageVersionList := &kapppkg.PackageList{}

	if packageName != "" {
		selectors = []crtclient.ListOption{
			crtclient.MatchingFields(map[string]string{"spec.refName": packageName}),
			crtclient.InNamespace(namespace),
		}
	}

	if err := c.client.List(context.Background(), packageVersionList, selectors...); err != nil {
		return nil, err
	}

	return packageVersionList, nil
}

// UpdatePackageInstall updates the PackageInstall CR
func (c *client) UpdatePackageInstall(packageInstall *kappipkg.PackageInstall, isPkgPluginCreatedSecret bool) error {
	installedPkg := packageInstall.DeepCopy()
	c.addAnnotations(&installedPkg.ObjectMeta, false, isPkgPluginCreatedSecret)

	if err := c.client.Update(context.Background(), installedPkg); err != nil {
		return err
	}

	return nil
}

// GetPackage gets Package CR
func (c *client) GetPackage(packageName, namespace string) (*kapppkg.Package, error) {
	pkg := &kapppkg.Package{}
	if err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: packageName, Namespace: namespace}, pkg); err != nil {
		return nil, err
	}
	return pkg, nil
}

// UpdatePackageRepository updates a PackageRepository CR
func (c *client) UpdatePackageRepository(repository *kappipkg.PackageRepository) error {
	if err := c.client.Update(context.Background(), repository); err != nil {
		return err
	}

	return nil
}
