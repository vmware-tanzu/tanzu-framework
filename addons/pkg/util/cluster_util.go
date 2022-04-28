// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	capiremote "sigs.k8s.io/cluster-api/controllers/remote"
	clusterapiutil "sigs.k8s.io/cluster-api/util"
	clusterapisecretutil "sigs.k8s.io/cluster-api/util/secret"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	runtanzuv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	bomtypes "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkr/pkg/types"
	vmoperatorv1alpha1 "github.com/vmware-tanzu/vm-operator-api/api/v1alpha1"
)

const (
	defaultClientTimeout = 10 * time.Second
)

// GetOwnerCluster returns the Cluster object owning the current resource.
func GetOwnerCluster(ctx context.Context, c client.Client, obj *metav1.ObjectMeta) (*clusterv1beta1.Cluster, error) {
	for _, ref := range obj.OwnerReferences {
		if ref.Kind != "Cluster" {
			continue
		}
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if gv.Group == clusterv1beta1.GroupVersion.Group {
			return GetClusterByName(ctx, c, obj.Namespace, ref.Name)
		}
	}
	return nil, nil
}

// GetClusterByName finds and return a Cluster object using the specified params.
func GetClusterByName(ctx context.Context, c client.Client, namespace, name string) (*clusterv1beta1.Cluster, error) {
	cluster := &clusterv1beta1.Cluster{}
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
func GetClustersByTKR(ctx context.Context, c client.Client, tkr *runtanzuv1alpha1.TanzuKubernetesRelease) ([]*clusterv1beta1.Cluster, error) {
	var clusters []*clusterv1beta1.Cluster

	if c == nil || tkr == nil {
		return nil, nil
	}

	clustersList := &clusterv1beta1.ClusterList{}

	if err := c.List(ctx, clustersList, client.MatchingLabels{constants.TKRLabel: tkr.Name}); err != nil {
		return nil, err
	}

	for i := range clustersList.Items {
		clusters = append(clusters, &clustersList.Items[i])
	}

	return clusters, nil
}

// GetClusterClient gets cluster's client
func GetClusterClient(ctx context.Context, currentClusterClient client.Client, scheme *runtime.Scheme, cluster client.ObjectKey) (client.Client, error) {
	config, err := capiremote.RESTConfig(ctx, constants.AddonControllerName, currentClusterClient, cluster)
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

// GetBOMForCluster gets the bom associated with the legacy-style TKGm cluster
func GetBOMForCluster(ctx context.Context, c client.Client, cluster *clusterv1beta1.Cluster) (*bomtypes.Bom, error) {
	tkrName := cluster.Labels[constants.TKRLabel]

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
func GetClusterKubeconfigSecretDetails(cluster *clusterv1beta1.Cluster) *ClusterKubeconfigSecretDetails {
	return &ClusterKubeconfigSecretDetails{
		Name:      clusterapisecretutil.Name(cluster.Name, clusterapisecretutil.Kubeconfig),
		Namespace: cluster.Namespace,
		Key:       clusterapisecretutil.KubeconfigDataName,
	}
}

// ClustersToRequests returns a list of Requests for clusters
func ClustersToRequests(clusters []*clusterv1beta1.Cluster, log logr.Logger) []ctrl.Request {
	var requests []ctrl.Request

	for _, cluster := range clusters {
		log.V(4).Info("Adding cluster for reconciliation",
			constants.ClusterNamespaceLogKey, cluster.Namespace, constants.ClusterNameLogKey, cluster.Name)

		requests = append(requests, ctrl.Request{
			NamespacedName: clusterapiutil.ObjectKey(cluster),
		})
	}

	return requests
}

func GetClusterLabel(clusterLabels map[string]string, labelKey string) string {
	if clusterLabels == nil {
		return ""
	}

	labelValue, ok := clusterLabels[labelKey]
	if !ok {
		return ""
	}

	return labelValue
}

// GetInfraProvider get infrastructure kind from cluster spec
func GetInfraProvider(cluster *clusterv1beta1.Cluster) (string, error) {
	var infraProvider string

	if cluster.Spec.InfrastructureRef != nil {
		infraProvider = cluster.Spec.InfrastructureRef.Kind
		switch infraProvider {
		case tkgconstants.InfrastructureRefVSphere:
			return tkgconstants.InfrastructureProviderVSphere, nil
		case tkgconstants.InfrastructureRefAWS:
			return tkgconstants.InfrastructureProviderAWS, nil
		case tkgconstants.InfrastructureRefAzure:
			return tkgconstants.InfrastructureProviderAzure, nil
		case tkgconstants.InfrastructureRefDocker:
			return tkgconstants.InfrastructureProviderDocker, nil
		}
	}

	return "", errors.New("unknown error in getting infraProvider")
}

// IsTKGSCluster checks if the cluster is a TKGS cluster
func IsTKGSCluster(ctx context.Context, c client.Client, cluster *clusterv1beta1.Cluster) (bool, error) {
	// Verify if operating on a TKGS cluster by checking if virtualmachine objects exist for the cluster label
	virtualMachineList := &vmoperatorv1alpha1.VirtualMachineList{}
	listOptions := client.MatchingLabels{
		tkgconstants.CAPVClusterSelectorKey: cluster.Name,
	}

	if err := c.List(ctx, virtualMachineList, listOptions); err != nil {
		// If CRD resource doesn't exist on cluster, it will throw an unKnownError with the error message that contains `no matches for kind "VirtualMachine"`
		if strings.Contains(err.Error(), "no matches for kind \"VirtualMachine\"") || apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return len(virtualMachineList.Items) != 0, nil
}
