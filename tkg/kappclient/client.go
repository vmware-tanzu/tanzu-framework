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
	secretgenctrl "github.com/vmware-tanzu/carvel-secretgen-controller/pkg/apis/secretgen2/v1alpha1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/tkgpackagedatamodel"
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
	return NewKappClientForContext(kubeCfgPath, "")
}

func NewKappClientForContext(kubeCfgPath, kubeContext string) (Client, error) {
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
	if err := secretgenctrl.AddToScheme(scheme); err != nil {
		return nil, err
	}

	if restConfig, err = GetKubeConfigForContext(kubeCfgPath, kubeContext); err != nil {
		return nil, err
	}

	// As there are many registered resources in the cluster, set the values for the maximum number of
	// queries per second and the maximum burst for throttle to a high value to avoid throttling of messages
	restConfig.QPS = constants.DefaultQPS
	restConfig.Burst = constants.DefaultBurst

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

// GetKubeConfig gets kubeconfig from the provided kubeconfig path. Otherwise, it gets the kubeconfig from "$HOME/.kube/config" if existing
func GetKubeConfig(kubeCfgPath string) (*rest.Config, error) {
	return GetKubeConfigForContext(kubeCfgPath, "")
}

// GetKubeConfigForContext gets kubeconfig from the provided kubeconfig path. Otherwise, it gets the kubeconfig from "$HOME/.kube/config" if existing
func GetKubeConfigForContext(kubeCfgPath, kubeContext string) (*rest.Config, error) {
	var (
		restConfig *rest.Config
		err        error
	)

	if kubeCfgPath == "" {
		if restConfig, err = k8sconfig.GetConfig(); err != nil {
			return nil, err
		}
	} else {
		config, err := clientcmd.LoadFromFile(kubeCfgPath)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to load kubeconfig from %s", kubeCfgPath)
		}
		if kubeContext != "" {
			config.CurrentContext = kubeContext
		}
		rawConfig, err := clientcmd.Write(*config)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to write config")
		}
		if restConfig, err = clientcmd.RESTConfigFromKubeConfig(rawConfig); err != nil {
			return nil, errors.Wrap(err, "Unable to set up rest config")
		}
	}

	return restConfig, nil
}

func (c *client) addAnnotations(meta *v1.ObjectMeta, pkgPluginResourceCreationStatus *tkgpackagedatamodel.PkgPluginResourceCreationStatus) {
	if meta.Annotations == nil {
		meta.Annotations = make(map[string]string)
	}
	if pkgPluginResourceCreationStatus.IsServiceAccountCreated {
		meta.Annotations[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-"+tkgpackagedatamodel.KindClusterRole] = fmt.Sprintf(tkgpackagedatamodel.ClusterRoleName, meta.Name, meta.Namespace)
		meta.Annotations[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-"+tkgpackagedatamodel.KindClusterRoleBinding] = fmt.Sprintf(tkgpackagedatamodel.ClusterRoleBindingName, meta.Name, meta.Namespace)
		meta.Annotations[tkgpackagedatamodel.TanzuPkgPluginAnnotation+"-"+tkgpackagedatamodel.KindServiceAccount] = fmt.Sprintf(tkgpackagedatamodel.ServiceAccountName, meta.Name, meta.Namespace)
	}
	if pkgPluginResourceCreationStatus.IsSecretCreated {
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
func (c *client) CreatePackageInstall(packageInstall *kappipkg.PackageInstall, pkgPluginResourceCreationStatus *tkgpackagedatamodel.PkgPluginResourceCreationStatus) error {
	installedPkg := packageInstall.DeepCopy()
	c.addAnnotations(&installedPkg.ObjectMeta, pkgPluginResourceCreationStatus)

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

// ListRegistrySecrets gets the list of all Secrets of type "kubernetes.io/dockerconfigjson"
func (c *client) ListRegistrySecrets(namespace string) (*corev1.SecretList, error) {
	var selectors []crtclient.ListOption
	secretList := &corev1.SecretList{}

	selectors = []crtclient.ListOption{crtclient.InNamespace(namespace), crtclient.MatchingFields(map[string]string{"type": string(corev1.SecretTypeDockerConfigJson)})}

	err := c.client.List(context.Background(), secretList, selectors...)
	if err != nil {
		return nil, err
	}
	return secretList, nil
}

// ListSecretExports gets the list of all SecretExports
func (c *client) ListSecretExports(namespace string) (*secretgenctrl.SecretExportList, error) {
	var selectors []crtclient.ListOption
	secretExportList := &secretgenctrl.SecretExportList{}

	selectors = []crtclient.ListOption{crtclient.InNamespace(namespace)}

	err := c.client.List(context.Background(), secretExportList, selectors...)
	if err != nil {
		return nil, err
	}
	return secretExportList, nil
}

// GetSecretExport gets the SecretExport having the same name and namespace as specified secret
func (c *client) GetSecretExport(secretName, namespace string) (*secretgenctrl.SecretExport, error) {
	secretExport := &secretgenctrl.SecretExport{}
	err := c.client.Get(context.Background(), crtclient.ObjectKey{Name: secretName, Namespace: namespace}, secretExport)
	if err != nil {
		return nil, err
	}
	return secretExport, nil
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
func (c *client) UpdatePackageInstall(packageInstall *kappipkg.PackageInstall, pkgPluginResourceCreationStatus *tkgpackagedatamodel.PkgPluginResourceCreationStatus) error {
	c.addAnnotations(&packageInstall.ObjectMeta, pkgPluginResourceCreationStatus)

	if err := c.client.Update(context.Background(), packageInstall); err != nil {
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

func (c *client) GetSecretValue(secretName, namespace string) ([]byte, error) {
	var err error

	secret := &corev1.Secret{}
	err = c.client.Get(context.Background(), crtclient.ObjectKey{Name: secretName, Namespace: namespace}, secret)
	if err != nil {
		return nil, err
	}

	var data []byte
	for _, value := range secret.Data {
		stringValue := string(value)
		if len(stringValue) < 3 {
			data = append(data, tkgpackagedatamodel.YamlSeparator...)
			data = append(data, "\n"...)
		}
		if len(stringValue) >= 3 && stringValue[:3] != tkgpackagedatamodel.YamlSeparator {
			data = append(data, tkgpackagedatamodel.YamlSeparator...)
			data = append(data, "\n"...)
		}
		data = append(data, value...)
	}

	return data, nil
}
