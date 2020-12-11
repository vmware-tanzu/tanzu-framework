package util

import (
	"context"
	"fmt"
	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// GetTKRByK8sVersion gets TKR for a given K8s version
func GetTKRByK8sVersion(ctx context.Context, c client.Client, k8sVersion string) (*runtanzuv1alpha1.TanzuKubernetesRelease, error) {
	var clusterTkr *runtanzuv1alpha1.TanzuKubernetesRelease

	tkrList := &runtanzuv1alpha1.TanzuKubernetesReleaseList{}
	if err := c.List(ctx, tkrList); err != nil {
		return nil, err
	}

	for _, tkr := range tkrList.Items {
		var tkrK8sVersion string

		if !strings.HasPrefix(tkr.Spec.KubernetesVersion, "v") {
			tkrK8sVersion = fmt.Sprintf("v%s", tkr.Spec.KubernetesVersion)
		}

		if tkrK8sVersion == k8sVersion {
			clusterTkr = &tkr
			break
		}
	}

	return clusterTkr, nil
}

// GetTKRByName gets TKR object given a TKR name
func GetTKRByName(ctx context.Context, c client.Client, tkrName string) (*runtanzuv1alpha1.TanzuKubernetesRelease, error) {
	if tkrName == "" {
		return nil, nil
	}

	tkr := &runtanzuv1alpha1.TanzuKubernetesRelease{}

	tkrNamespaceName := client.ObjectKey{
		Name: tkrName,
	}

	if err := c.Get(context.Background(), tkrNamespaceName, tkr); err != nil {
		return nil, err
	}

	return tkr, nil
}
