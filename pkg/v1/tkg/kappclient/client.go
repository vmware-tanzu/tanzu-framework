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

	kappipkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/installpackage/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	kapppkg "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/packages/v1alpha1"

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

// CreateInstalledPackage creates a InstalledPackage CR
func (c *client) CreateInstalledPackage(installedPackage *kappipkg.InstalledPackage, isPkgPluginCreatedSvcAccount, isPkgPluginCreatedSecret bool) error {
	installedPkg := installedPackage.DeepCopy()
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

// GetInstalledPackage gets the InstalledPackage CR for the provided package name
func (c *client) GetInstalledPackage(installedPackageName, namespace string) (*kappipkg.InstalledPackage, error) {
	installedPkg := &kappipkg.InstalledPackage{}
	if err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: installedPackageName, Namespace: namespace}, installedPkg); err != nil {
		return nil, err
	}

	return installedPkg, nil
}

// GetPackageByName gets the package with the specified name
func (c *client) GetPackageByName(packageName, namespace string) (*kapppkg.Package, error) {
	pkg := &kapppkg.Package{}
	if err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: packageName, Namespace: namespace}, pkg); err != nil {
		return nil, err
	}

	return pkg, nil
}

// ListInstalledPackages gets the list of InstalledPackage CR in the specified namespace.
// If no namespace be provided, it returns the list of installed packages across all namespaces
func (c *client) ListInstalledPackages(namespace string) (*kappipkg.InstalledPackageList, error) {
	var selectors []crtclient.ListOption
	installedPackageList := &kappipkg.InstalledPackageList{}

	if namespace != "" {
		selectors = []crtclient.ListOption{crtclient.InNamespace(namespace)}
	}

	if err := c.client.List(context.Background(), installedPackageList, selectors...); err != nil {
		return nil, err
	}

	return installedPackageList, nil
}

// ListPackageVersions gets the list of PackageVersion CR
func (c *client) ListPackageVersions(packageName, namespace string) (*kapppkg.PackageVersionList, error) {
	var selectors []crtclient.ListOption
	packageVersionList := &kapppkg.PackageVersionList{}

	if packageName != "" {
		selectors = []crtclient.ListOption{
			crtclient.MatchingFields(map[string]string{"spec.packageName": packageName}),
			crtclient.InNamespace(namespace),
		}
	}

	if err := c.client.List(context.Background(), packageVersionList, selectors...); err != nil {
		return nil, err
	}

	return packageVersionList, nil
}
