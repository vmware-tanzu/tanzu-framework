package util

import (
	"context"
	"github.com/vmware-tanzu-private/core/addons/constants"
	addonsv1alpha1 "github.com/vmware-tanzu-private/core/apis/addons/v1alpha1"
	kappctrl "github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	clusterapiv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// GetAddonsInCluster returns
func GetAddonsInCluster(ctx context.Context, c client.Client) ([]string, error) {
	var addons []string

	apps := &kappctrl.AppList{}
	if err := c.List(ctx, apps, client.InNamespace(constants.TKG_ADDONS_APP_NAMESPACE)); err != nil {
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
