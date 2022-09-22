package controllers

import (
	"context"
	"fmt"
	"strings"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/vmware-tanzu/tanzu-framework/addons/pkg/constants"
	csiv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/addonconfigs/csi/v1alpha1"
)

// getOwnerCluster verifies that the AzureDiskCSIConfig has a cluster as its owner reference,
// and returns the cluster. It tries to read the cluster name from the VSphereCSIConfig's owner reference objects.
// If not there, we assume the owner cluster and VSphereCSIConfig always has the same name.
func (r *AzureDiskCSIConfigReconciler) getOwnerCluster(ctx context.Context,
	azureDiskCSIConfig *csiv1alpha1.AzureDiskCSIConfig) (*clusterapiv1beta1.Cluster, error) {

	logger := log.FromContext(ctx)
	cluster := &clusterapiv1beta1.Cluster{}
	clusterName := azureDiskCSIConfig.Name // usually the corresponding 'cluster' shares the same name

	// retrieve the owner cluster for the VSphereCSIConfig object
	for _, ownerRef := range azureDiskCSIConfig.GetOwnerReferences() {
		if strings.EqualFold(ownerRef.Kind, constants.ClusterKind) {
			clusterName = ownerRef.Name
			break
		}
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: azureDiskCSIConfig.Namespace, Name: clusterName}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Cluster resource '%s/%s' not found", azureDiskCSIConfig.Namespace, clusterName))
			return nil, nil
		}
		logger.Error(err, fmt.Sprintf("Unable to fetch cluster '%s/%s'", azureDiskCSIConfig.Namespace, clusterName))
		return nil, err
	}

	return cluster, nil
}

// mapVSphereCSIConfigToDataValues maps VSphereCSIConfig CR to data values
func (r *AzureDiskCSIConfigReconciler) mapAzureDiskCSIConfigToDataValues(ctx context.Context,
	azureDiskCSIConfig *csiv1alpha1.AzureDiskCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {

	dvs := &DataValues{}
	dvs.AzureDiskCSI = &DataValuesAzureDiskCSI{}
	dvs.AzureDiskCSI.Namespace = azureDiskCSIConfig.Spec.AzureDiskCSI.Namespace
	dvs.AzureDiskCSI.HTTPProxy = azureDiskCSIConfig.Spec.AzureDiskCSI.HTTPProxy
	dvs.AzureDiskCSI.HTTPSProxy = azureDiskCSIConfig.Spec.AzureDiskCSI.HTTPSProxy
	dvs.AzureDiskCSI.NoProxy = azureDiskCSIConfig.Spec.AzureDiskCSI.NoProxy
	dvs.AzureDiskCSI.DeploymentReplicas = *azureDiskCSIConfig.Spec.AzureDiskCSI.DeploymentReplicas

	return dvs, nil
}
