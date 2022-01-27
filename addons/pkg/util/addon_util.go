// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	pkgiv1alpha1 "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/packaging/v1alpha1"

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

// GenerateAppSecretNameFromAddonSecret generates app secret name from addon secret
func GenerateAppSecretNameFromAddonSecret(addonSecret *corev1.Secret) string {
	return fmt.Sprintf("%s-data-values", GenerateAppNameFromAddonSecret(addonSecret))
}

// GenerateDataValueSecretName generates data value secret name from the cluster and the package name
func GenerateDataValueSecretName(clusterName, pkgName string) string {
	return fmt.Sprintf("%s-%s-data-values", clusterName, pkgName)
// GenerateDataValueSecretNameFromAddonAndClusterNames generates data value secret name from addon names
func GenerateDataValueSecretNameFromAddonAndClusterNames(clusterName, addonName string) string {
	return fmt.Sprintf("%s-%s-data-values", clusterName, addonName)
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

// AddFinalizerToCRD adds finalizer to the config CRD if not present and
// returns true if finalizer is added
func AddFinalizerToCRD(
	log logr.Logger,
	addonName string,
	configCRD client.Object) bool {

	var patchAddonSecret bool

	// add finalizer to addon secret
	if !controllerutil.ContainsFinalizer(configCRD, addontypes.AddonFinalizer) {
		log.Info("Adding finalizer to addon secret", constants.AddonNameLogKey, addonName)
		controllerutil.AddFinalizer(configCRD, addontypes.AddonFinalizer)
		patchAddonSecret = true
	}

	return patchAddonSecret
}

// RemoveFinalizerFromCRD removes finalizer from the config CRD if not present and
// returns true if finalizer is removed
func RemoveFinalizerFromCRD(
	log logr.Logger,
	addonName string,
	configCRD client.Object) bool {

	var patchAddonSecret bool

	// add finalizer to addon secret
	if !controllerutil.ContainsFinalizer(configCRD, addontypes.AddonFinalizer) {
		log.Info("Removing finalizer to addon secret", constants.AddonNameLogKey, addonName)
		controllerutil.RemoveFinalizer(configCRD, addontypes.AddonFinalizer)
		patchAddonSecret = true
	}

	return patchAddonSecret
}

// get paths for external CRDs by introspecting versions of the go dependencies
func GetExternalCRDPaths(externalDeps map[string][]string) ([]string, error) {
	var crdPaths []string
	gopath, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return crdPaths, err
	}
	for dep, crdDirs := range externalDeps {
		depPath, err := exec.Command("go", "list", "-m", "-f", "{{ .Path }}@{{ .Version }}", dep).Output()
		if err != nil {
			return crdPaths, err
		}
		for _, crdDir := range crdDirs {
			crdPaths = append(crdPaths, filepath.Join(strings.TrimSuffix(string(gopath), "\n"),
				"pkg", "mod", strings.TrimSuffix(string(depPath), "\n"), crdDir))
		}
	}

	logf.Log.Info("external CRD paths", "crdPaths", crdPaths)
	return crdPaths, nil
}

func GetServiceCIDR(cluster *clusterapiv1beta1.Cluster) (string, error) {
	var serviceCIDR string
	if cluster.Spec.ClusterNetwork != nil && cluster.Spec.ClusterNetwork.Services != nil && len(cluster.Spec.ClusterNetwork.Services.CIDRBlocks) > 0 {
		serviceCIDR = cluster.Spec.ClusterNetwork.Services.CIDRBlocks[0]
	} else {
		return "", errors.New("Unable to get cluster serviceCIDR")
	}

	return serviceCIDR, nil
}

func GetInfraProvider(cluster *clusterapiv1beta1.Cluster) (string, error) {
	var infraProvider string

	if cluster.Spec.InfrastructureRef != nil {

		infraProvider = cluster.Spec.InfrastructureRef.Kind

		switch infraProvider {
		case constants.InfrastructureRefVSphere:
			infraProvider = constants.InfrastructureProviderVSphere
		case constants.InfrastructureRefAWS:
			infraProvider = constants.InfrastructureProviderAWS
		case constants.InfrastructureRefAzure:
			infraProvider = constants.InfrastructureProviderAzure
		case constants.InfrastructureRefDocker:
			infraProvider = constants.InfrastructureProviderDocker
		default:
			infraProvider = constants.InfrastructureProviderVSphere
		}
	} else {
		return "", errors.New("Unable to get infraProvider")
	}

	return infraProvider, nil
}
