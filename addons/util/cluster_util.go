package util

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	bomv1alpha1 "github.com/vmware-tanzu-private/core/apis/bom/v1alpha1"
	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	controlplanev1alpha3 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1alpha3"
	clusterapisecretutil "sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

// GetOwnerCluster returns the Cluster object owning the current resource.
func GetOwnerCluster(ctx context.Context, c client.Client, obj metav1.ObjectMeta) (*clusterv1alpha3.Cluster, error) {
	for _, ref := range obj.OwnerReferences {
		if ref.Kind != "Cluster" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if gv.Group == clusterv1alpha3.GroupVersion.Group {
			return GetClusterByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// GetClusterByName finds and return a Cluster object using the specified params.
func GetClusterByName(ctx context.Context, c client.Client, namespace, name string) (*clusterv1alpha3.Cluster, error) {
	cluster := &clusterv1alpha3.Cluster{}
	key := client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	if err := c.Get(ctx, key, cluster); err != nil {
		return nil, err
	}

	return cluster, nil
}

// GetKCPForCluster returns control plane for the cluster
func GetKCPForCluster(ctx context.Context, c client.Client, cluster *clusterv1alpha3.Cluster) (*controlplanev1alpha3.KubeadmControlPlane, error) {
	if c == nil || cluster == nil {
		return nil, nil
	}

	if cluster.Spec.ControlPlaneRef == nil {
		return nil, nil
	}

	kcp := &controlplanev1alpha3.KubeadmControlPlane{}
	kcpObjectKey := client.ObjectKey{
		Namespace: cluster.Spec.ControlPlaneRef.Namespace,
		Name:      cluster.Spec.ControlPlaneRef.Name,
	}
	if err := c.Get(ctx, kcpObjectKey, kcp); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	return kcp, nil
}

func getKCPByK8sVersion(ctx context.Context, c client.Client, k8sVersion string) ([]*controlplanev1alpha3.KubeadmControlPlane, error) {
	var kcps []*controlplanev1alpha3.KubeadmControlPlane

	if k8sVersion == "" {
		return nil, nil
	}

	if !strings.HasPrefix(k8sVersion, "v") {
		k8sVersion = fmt.Sprintf("v%s", k8sVersion)
	}

	kcpList := &controlplanev1alpha3.KubeadmControlPlaneList{}
	if err := c.List(context.TODO(), kcpList); err != nil {
		return nil, err
	}

	for _, kcp := range kcpList.Items {
		if kcp.Spec.Version != k8sVersion {
			continue
		}
		kcps = append(kcps, &kcp)
	}

	return kcps, nil
}

// GetClustersByTKR gets the clusters using this TKR
func GetClustersByTKR(ctx context.Context, c client.Client, tkr *runtanzuv1alpha1.TanzuKubernetesRelease) ([]*clusterv1alpha3.Cluster, error) {
	var clusters []*clusterv1alpha3.Cluster

	if c == nil || tkr == nil {
		return nil, nil
	}

	kcps, err := getKCPByK8sVersion(ctx, c, tkr.Spec.KubernetesVersion)
	if err != nil {
		return nil, err
	}

	for _, kcp := range kcps {
		cluster, err := GetOwnerCluster(context.TODO(), c, kcp.ObjectMeta)
		if err != nil {
			return nil, err
		}

		if cluster != nil {
			clusters = append(clusters, cluster)
		}
	}

	return clusters, nil

}

// GetTKRForCluster gets the TKR for cluster
func GetTKRForCluster(ctx context.Context, c client.Client, cluster *clusterv1alpha3.Cluster) (*runtanzuv1alpha1.TanzuKubernetesRelease, error) {
	if c == nil || cluster == nil {
		return nil, nil
	}

	kcp, err := GetKCPForCluster(ctx, c, cluster)
	if err != nil {
		return nil, err
	}

	if kcp == nil {
		return nil, nil
	}

	k8sVersion := kcp.Spec.Version

	clusterTkr, err := GetTKRByK8sVersion(ctx, c, k8sVersion)
	if err != nil {
		return nil, err
	}

	return clusterTkr, nil
}

// GetBOMForCluster gets the bom associated with the cluster
func GetBOMForCluster(ctx context.Context, c client.Client, cluster *clusterv1alpha3.Cluster) (*bomv1alpha1.BomConfig, error) {

	tkr, err := GetTKRForCluster(ctx, c, cluster)
	if err != nil {
		return nil, err
	}

	if tkr == nil {
		return nil, nil
	}

	bomConfig, err := GetBOMByTKRName(ctx, c, tkr.Name)
	if err != nil {
		return nil, err
	}

	return bomConfig, nil
}

type ClusterKubeconfigSecretDetails struct {
	Name      string
	Namespace string
	Key       string
}

// GetClusterKubeconfigSecretDetails returns the name, namespace and key of the cluster's kubeconfig secret
func GetClusterKubeconfigSecretDetails(cluster *clusterv1alpha3.Cluster) *ClusterKubeconfigSecretDetails {

	return &ClusterKubeconfigSecretDetails{
		Name:      clusterapisecretutil.Name(cluster.Name, clusterapisecretutil.Kubeconfig),
		Namespace: cluster.Namespace,
		Key:       clusterapisecretutil.KubeconfigDataName,
	}
}
