// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"
	kapppkgv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apiserver/apis/datapackaging/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu/tanzu-framework/addons/pkg/types"
	bomtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
)

// GetAddonSecretsForCluster gets the addon secrets belonging to the cluster
func GetAddonSecretsForCluster(ctx context.Context, c client.Client, cluster *clusterapiv1beta1.Cluster) (*corev1.SecretList, error) {
	if cluster == nil {
		return nil, nil
	}

	addonSecrets := &corev1.SecretList{}
	if err := c.List(ctx, addonSecrets, client.InNamespace(cluster.Namespace),
		client.MatchingLabels{addontypes.ClusterNameLabel: cluster.Name}); err != nil {
		return nil, err
	}

	return addonSecrets, nil
}

// GetAddonNameFromAddonSecret gets the addon name from addon secret
func GetAddonNameFromAddonSecret(addonSecret *corev1.Secret) string {
	return addonSecret.Labels[addontypes.AddonNameLabel]
}

// GetClusterNameFromAddonSecret gets the cluster name from addon secret
func GetClusterNameFromAddonSecret(addonSecret *corev1.Secret) string {
	return addonSecret.Labels[addontypes.ClusterNameLabel]
}

// GenerateAppNameFromAddonSecret generates app name given an addon secret
func GenerateAppNameFromAddonSecret(addonSecret *corev1.Secret) string {
	addonName := GetAddonNameFromAddonSecret(addonSecret)
	if addonName == "" {
		return ""
	}

	remoteApp := IsRemoteApp(addonSecret)
	if remoteApp {
		clusterName := GetClusterNameFromAddonSecret(addonSecret)
		if clusterName == "" {
			return ""
		}
		return fmt.Sprintf("%s-%s", clusterName, addonName)
	}
	return addonName
}

// GeneratePackageInstallName is the util function to generate the PackageInstall CR name in a consistent manner.
// clusterName is the name of cluster within which all resources associated with this PackageInstall CR is installed.
// It does not necessarily
// mean the PackageInstall CR will be installed in that cluster. I.e., the kapp-controller PackageInstall CR is installed
// in the management cluster but is named after "<workload-cluster-name>-kapp-controller". It indicates that this kapp-controller
// PackageInstall is for reconciling resources in a cluster named "<workload-cluster-name>".
// addonName is the short name of a Tanzu addon with which the PackageInstall CR is associated.
func GeneratePackageInstallName(clusterName, addonName string) string {
	return fmt.Sprintf("%s-%s", clusterName, strings.Split(addonName, ".")[0])
}

func GetPackageMetadata(ctx context.Context, c client.Client, carvelPkgName, carvelPkgNamespace string) (string, string, error) {
	pkg := &kapppkgv1alpha1.Package{}
	if err := c.Get(ctx, client.ObjectKey{Name: carvelPkgName, Namespace: carvelPkgNamespace}, pkg); err != nil {
		return "", "", err
	}
	return pkg.Spec.RefName, pkg.Spec.Version, nil
}

// ParseStringForLabel parse the package ref name to make it valid for K8S object labels.
// A package ref name could contain some characters that are not allowed as a label value.
// Also the label should not end with any non-alphanumeric characters.
// The regex used for validation is (([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?
func ParseStringForLabel(s string) string {
	// Replace + sign with ---
	safeLabel := strings.ReplaceAll(s, "+", "---")
	if len(safeLabel) <= 63 {
		return strings.TrimRight(safeLabel, "_.-")
	}
	return strings.TrimRight(safeLabel[:63], "_.-")
}

// GenerateAppSecretNameFromAddonSecret generates app secret name from addon secret
func GenerateAppSecretNameFromAddonSecret(addonSecret *corev1.Secret) string {
	return fmt.Sprintf("%s-data-values", GenerateAppNameFromAddonSecret(addonSecret))
}

// GenerateDataValueSecretName generates data value secret name from the cluster and the package name
func GenerateDataValueSecretName(clusterName, carvelPkgRefName string) string {
	return fmt.Sprintf("%s-%s-data-values", clusterName, packageShortName(carvelPkgRefName))
}

// GeneratePackageSecretName generates secret name for a package from the cluster and the package name
func GeneratePackageSecretName(clusterName, carvelPkgRefName string) string {
	return fmt.Sprintf("%s-%s-package", clusterName, packageShortName(carvelPkgRefName))
}

func packageShortName(pkgRefName string) string {
	nameTokens := strings.Split(pkgRefName, ".")
	if len(nameTokens) >= 1 {
		return nameTokens[0]
	}
	return pkgRefName
}

// GenerateAppNamespaceFromAddonSecret generates app namespace from addons secret
func GenerateAppNamespaceFromAddonSecret(addonSecret *corev1.Secret, defaultAddonNamespace string) string {
	remoteApp := IsRemoteApp(addonSecret)
	if remoteApp {
		return addonSecret.Namespace
	}
	return defaultAddonNamespace
}

// GetClientFromAddonSecret gets appropriate cluster client given addon secret
func GetClientFromAddonSecret(addonSecret *corev1.Secret, localClient, remoteClient client.Client) client.Client {
	var clusterClient client.Client
	remoteApp := IsRemoteApp(addonSecret)
	if remoteApp {
		clusterClient = localClient
	} else {
		clusterClient = remoteClient
	}
	return clusterClient
}

// GetImageInfo gets the image Info of an addon
func GetImageInfo(addonConfig *bomtypes.Addon, imageRepository, imagePullPolicy string, bom *bomtypes.Bom) ([]byte, error) {
	componentRefs := addonConfig.AddonContainerImages

	addonImageInfo := &addontypes.AddonImageInfo{Info: addontypes.ImageInfo{ImageRepository: imageRepository, ImagePullPolicy: imagePullPolicy, Images: map[string]addontypes.Image{}}}

	// No Image will be added if componentRefs is empty
	for _, componentRef := range componentRefs {
		for _, imageRef := range componentRef.ImageRefs {
			image, err := bom.GetImageInfo(componentRef.ComponentRef, "", imageRef)
			if err != nil {
				return nil, err
			}
			addonImageInfo.Info.Images[imageRef] = addontypes.Image{ImagePath: image.ImagePath, Tag: image.Tag}
		}
	}

	ImageInfoBytes, err := yaml.Marshal(addonImageInfo)
	if err != nil {
		return nil, err
	}

	outputBytes := append([]byte(constants.TKGDataValueFormatString), ImageInfoBytes...)
	return outputBytes, nil
}

// GetApp gets the app CR from cluster
func GetApp(ctx context.Context,
	localClient client.Client,
	remoteClient client.Client,
	addonSecret *corev1.Secret,
	defaultAddonNamespace string) (*kappctrl.App, error) {

	app := &kappctrl.App{}
	appObjectKey := client.ObjectKey{
		Name:      GenerateAppNameFromAddonSecret(addonSecret),
		Namespace: GenerateAppNamespaceFromAddonSecret(addonSecret, defaultAddonNamespace),
	}

	var clusterClient client.Client
	remoteApp := IsRemoteApp(addonSecret)
	if remoteApp {
		clusterClient = localClient
	} else {
		clusterClient = remoteClient
	}

	if err := clusterClient.Get(ctx, appObjectKey, app); err != nil {
		return nil, err
	}

	return app, nil
}

// GetPackageInstallFromAddonSecret gets the PackageInstall CR from cluster
func GetPackageInstallFromAddonSecret(ctx context.Context,
	remoteClient client.Client,
	addonSecret *corev1.Secret,
	defaultAddonNamespace string) (*pkgiv1alpha1.PackageInstall, error) {

	pkgi := &pkgiv1alpha1.PackageInstall{}
	pkgiObjectKey := client.ObjectKey{
		Name:      GenerateAppNameFromAddonSecret(addonSecret),
		Namespace: GenerateAppNamespaceFromAddonSecret(addonSecret, defaultAddonNamespace),
	}

	if err := remoteClient.Get(ctx, pkgiObjectKey, pkgi); err != nil {
		return nil, err
	}

	return pkgi, nil
}

// IsAppPresent returns true if app is present on the cluster
func IsAppPresent(ctx context.Context,
	localClient client.Client,
	remoteClient client.Client,
	addonSecret *corev1.Secret,
	defaultAddonNamespace string) (bool, error) {

	_, err := GetApp(ctx, localClient, remoteClient, addonSecret, defaultAddonNamespace)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

// IsRemoteApp returns true if App needs to be remote instead of App being on local cluster
func IsRemoteApp(addonSecret *corev1.Secret) bool {
	remoteApp := addonSecret.Annotations[addontypes.AddonRemoteAppAnnotation]
	if remoteApp == "" {
		return false
	}
	isRemoteApp, _ := strconv.ParseBool(remoteApp)
	return isRemoteApp
}

// IsAddonPaused returns true if Addon is paused
func IsAddonPaused(addonSecret *corev1.Secret) bool {
	annotations := addonSecret.GetAnnotations()
	if annotations == nil {
		return false
	}
	_, ok := annotations[addontypes.AddonPausedAnnotation]
	return ok
}

// IsPackageInstallPresent returns true if PackageInstall is present on the cluster
func IsPackageInstallPresent(ctx context.Context,
	localClient client.Client,
	addonSecret *corev1.Secret,
	defaultAddonNamespace string) (bool, error) {

	_, err := GetPackageInstallFromAddonSecret(ctx, localClient, addonSecret, defaultAddonNamespace)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return false, err
		}
		return false, nil
	}

	return true, nil
}

// GetServiceCIDRs returns the Service CIDR blocks for both IPv4 and IPv6 family
// Parse Service CIDRBlocks obtained from the cluster and return the following from the function:
// <IPv4 CIDRs, IPv6 CIDRs, error>
// The first two return parameters should be used only if the function returns error as nil.
// Also, note that when no error is returned, IPv4 and/or IPv6 CIDRs may still be empty depending on the cluster Service CIDR Blocks
func GetServiceCIDRs(cluster *clusterapiv1beta1.Cluster) (string, string, error) {
	var serviceCIDRs []string
	serviceCIDR, serviceCIDRv6 := "", ""
	if cluster.Spec.ClusterNetwork != nil && cluster.Spec.ClusterNetwork.Services != nil && len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) > 0 {
		serviceCIDRs = cluster.Spec.ClusterNetwork.Services.CIDRBlocks
		if len(serviceCIDRs) > 2 {
			return "", "", errors.New("too many CIDRs specified")
		}

		for _, cidr := range serviceCIDRs {
			ip, _, err := net.ParseCIDR(cidr)
			if err != nil {
				return "", "", errors.Errorf("could not parse CIDR: %s", err)
			}
			if ip.To4() != nil {
				serviceCIDR = cidr
			} else {
				if ip.To16() == nil {
					return "", "", errors.New("Unknown IP type in Service CIDR")
				}
				serviceCIDRv6 = cidr
			}
		}
	} else {
		return "", "", errors.New("Unable to get service CIDRBlocks from cluster")
	}

	return serviceCIDR, serviceCIDRv6, nil
}
