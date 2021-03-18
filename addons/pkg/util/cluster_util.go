// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	capiremote "sigs.k8s.io/cluster-api/controllers/remote"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1alpha3 "sigs.k8s.io/cluster-api/api/v1alpha3"
	clusterapisecretutil "sigs.k8s.io/cluster-api/util/secret"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu-private/core/addons/pkg/constants"
	runtanzuv1alpha1 "github.com/vmware-tanzu-private/core/apis/run/v1alpha1"
	bomtypes "github.com/vmware-tanzu-private/core/pkg/v1/tkr/pkg/types"
)

const (
	defaultClientTimeout = 10 * time.Second
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

// GetClustersByTKR gets the clusters using this TKR
func GetClustersByTKR(ctx context.Context, c client.Client, tkr *runtanzuv1alpha1.TanzuKubernetesRelease) ([]*clusterv1alpha3.Cluster, error) {
	var clusters []*clusterv1alpha3.Cluster

	if c == nil || tkr == nil {
		return nil, nil
	}

	clustersList := &clusterv1alpha3.ClusterList{}

	if err := c.List(ctx, clustersList, client.MatchingLabels{constants.TKRLabel: tkr.Name}); err != nil {
		return nil, err
	}

	for _, cluster := range clustersList.Items {
		clusters = append(clusters, &cluster)
	}

	return clusters, nil
}

// GetClusterClient gets cluster's client
func GetClusterClient(ctx context.Context, currentClusterClient client.Client, scheme *runtime.Scheme, cluster client.ObjectKey) (client.Client, error) {
	config, err := capiremote.RESTConfig(ctx, currentClusterClient, cluster)
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching REST client config for remote cluster %q", cluster.String())
	}
	config.Timeout = defaultClientTimeout

	// Create a mapper for it
	mapper, err := apiutil.NewDynamicRESTMapper(config)
	if err != nil {
		return nil, errors.Wrapf(err, "error creating dynamic rest mapper for remote cluster %q", cluster.String())
	}

	// Create the client for the remote cluster
	c, err := client.New(config, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return nil, errors.Wrapf(err, "error creating client for remote cluster %q", cluster.String())
	}

	return c, nil
}

// GetTKRForCluster gets the TKR for cluster
func GetTKRForCluster(ctx context.Context, c client.Client, cluster *clusterv1alpha3.Cluster) (*runtanzuv1alpha1.TanzuKubernetesRelease, error) {
	if c == nil || cluster == nil {
		return nil, nil
	}

	tkrName := GetTKRNameForCluster(ctx, c, cluster)
	if tkrName == "" {
		return nil, nil
	}

	tkr, err := GetTKRByName(ctx, c, tkrName)
	if err != nil {
		return nil, err
	}

	return tkr, nil
}

// GetTKRNameForCluster get the TKR name for the cluster
func GetTKRNameForCluster(ctx context.Context, c client.Client, cluster *clusterv1alpha3.Cluster) string {
	if c == nil || cluster == nil {
		return ""
	}

	return cluster.Labels[constants.TKRLabel]
}

// GetBOMForCluster gets the bom associated with the cluster
func GetBOMForCluster(ctx context.Context, c client.Client, cluster *clusterv1alpha3.Cluster) (*bomtypes.Bom, error) {

	tkrName := GetTKRNameForCluster(ctx, c, cluster)
	if tkrName == "" {
		return nil, nil
	}

	bom, err := GetBOMByTKRName(ctx, c, tkrName)
	if err != nil {
		return nil, err
	}

	return bom, nil
}

// ClusterKubeconfigSecretDetails contains the cluster kubeconfig secret details.
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
