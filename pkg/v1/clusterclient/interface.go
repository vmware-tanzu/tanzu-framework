package clusterclient

import (
	runv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// Client is the cluster client interface
type Client interface {
	GetTanzuKubernetesReleases(tkrName string) ([]runv1alpha1.TanzuKubernetesRelease, error)
	GetBomConfigMap(tkrNameLabel string) (corev1.ConfigMap, error)
	GetClusterInfrastructure() (string, error)
}
