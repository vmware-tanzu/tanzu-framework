// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"fmt"
	"strconv"

	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu-private/core/addons/pkg/constants"
	addontypes "github.com/vmware-tanzu-private/core/addons/pkg/types"
	bomtypes "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
)

// GetAddonSecretsForCluster gets the addon secrets belonging to the cluster
func GetAddonSecretsForCluster(ctx context.Context, c client.Client, cluster *clusterapiv1alpha3.Cluster) (*corev1.SecretList, error) {
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

// GenerateAppNamespaceFromAddonSecret generates app namespace from addons secret
func GenerateAppNamespaceFromAddonSecret(addonSecret *corev1.Secret) string {
	remoteApp := IsRemoteApp(addonSecret)
	if remoteApp {
		return addonSecret.Namespace
	}
	return constants.TKGAddonsAppNamespace
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
func GetImageInfo(addonConfig *bomtypes.Addon, imageRepository string, bom *bomtypes.Bom) ([]byte, error) {
	componentRefs := addonConfig.AddonContainerImages

	addonImageInfo := &addontypes.AddonImageInfo{Info: addontypes.ImageInfo{ImageRepository: imageRepository, ImagePullPolicy: constants.TKGAddonsImagePullPolicy, Images: map[string]addontypes.Image{}}}

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

// GetApp gets the kapp from cluster
func GetApp(ctx context.Context,
	localClient client.Client,
	remoteClient client.Client,
	addonSecret *corev1.Secret) (*kappctrl.App, error) {

	app := &kappctrl.App{}
	appObjectKey := client.ObjectKey{
		Name:      GenerateAppNameFromAddonSecret(addonSecret),
		Namespace: GenerateAppNamespaceFromAddonSecret(addonSecret),
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

// IsAppPresent returns true if app is present on the cluster
func IsAppPresent(ctx context.Context,
	localClient client.Client,
	remoteClient client.Client,
	addonSecret *corev1.Secret) (bool, error) {

	_, err := GetApp(ctx, localClient, remoteClient, addonSecret)
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
