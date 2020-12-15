package util

import (
	"context"
	"fmt"
	"github.com/vmware-tanzu-private/core/addons/constants"
	addonsv1alpha1 "github.com/vmware-tanzu-private/core/apis/addons/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
)

// GetAddonSecretsForCluster gets the addon secrets belonging to the cluster
func GetAddonSecretsForCluster(ctx context.Context, c client.Client, cluster *clusterapiv1alpha3.Cluster) (*corev1.SecretList, error) {
	if cluster == nil {
		return nil, nil
	}

	addonSecrets := &corev1.SecretList{}
	if err := c.List(ctx, addonSecrets, client.InNamespace(cluster.Namespace),
		client.MatchingLabels{addonsv1alpha1.ClusterNameLabel: cluster.Name}); err != nil {
		return nil, err
	}

	return addonSecrets, nil
}

// GetAddonNameFromAddonSecret gets the addon name from addon secret
func GetAddonNameFromAddonSecret(addonSecret *corev1.Secret) string {
	return addonSecret.Labels[addonsv1alpha1.AddonNameLabel]
}

// GetClusterNameFromAddonSecret gets the cluster name from addon secret
func GetClusterNameFromAddonSecret(addonSecret *corev1.Secret) string {
	return addonSecret.Labels[addonsv1alpha1.ClusterNameLabel]
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
	} else {
		return addonName
	}
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
	} else {
		return constants.TKGAddonsAppNamespace
	}
}

// GetClientFromAddonSecret gets appropriate cluster client given addon secret
func GetClientFromAddonSecret(addonSecret *corev1.Secret, localClient client.Client, remoteClient client.Client) client.Client {
	var clusterClient client.Client
	remoteApp := IsRemoteApp(addonSecret)
	if remoteApp {
		clusterClient = localClient
	} else {
		clusterClient = remoteClient
	}
	return clusterClient
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

// GetAddonsInCluster returns
func GetAddonsInCluster(ctx context.Context, c client.Client) ([]string, error) {
	var addons []string

	apps := &kappctrl.AppList{}
	if err := c.List(ctx, apps, client.InNamespace(constants.TKGAddonsAppNamespace)); err != nil {
		return nil, err
	}

	for _, app := range apps.Items {
		// Filter only those annotations that have addon type annotation. This is to ensure it is a tkg created addon.
		if addonType := app.Annotations[addonsv1alpha1.AddonTypeAnnotation]; addonType != "" {
			addons = append(addons, app.Name)
		}
	}

	return addons, nil
}

// IsRemoteApp returns true if App needs to be remote instead of App being on local cluster
func IsRemoteApp(addonSecret *corev1.Secret) bool {
	remoteApp := addonSecret.Annotations[addonsv1alpha1.AddonRemoteAppAnnotation]
	if remoteApp == "" {
		return false
	}
	isRemoteApp, _ := strconv.ParseBool(remoteApp)
	return isRemoteApp
}

// GetAddonNameFromApp gets the addon name from app
func GetAddonNameFromApp(app *kappctrl.App) string {
	return app.Annotations[addonsv1alpha1.AddonNameAnnotation]
}

// GetAddonNamespaceFromApp gets the addon secret namespace from app
func GetAddonNamespaceFromApp(app *kappctrl.App) string {
	return app.Annotations[addonsv1alpha1.AddonNamespaceAnnotation]
}
