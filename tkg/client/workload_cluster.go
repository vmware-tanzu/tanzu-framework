// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package client

import (
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/tools/clientcmd"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"

	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
	"github.com/vmware-tanzu/tanzu-framework/tkg/log"
	"github.com/vmware-tanzu/tanzu-framework/tkg/utils"
)

// GetWorkloadClusterCredentialsOptions contains options supported by GetWorkloadClusterCredntials
type GetWorkloadClusterCredentialsOptions struct {
	ClusterName string
	Namespace   string
	ExportFile  string
}

// DeleteWorkloadClusterOptions contains options supported by DeleteWorkloadCluster
type DeleteWorkloadClusterOptions struct {
	ClusterName string
	Namespace   string
}

// GetWorkloadClusterCredentials gets and saves workload cluster credentials
func (c *TkgClient) GetWorkloadClusterCredentials(options GetWorkloadClusterCredentialsOptions) (string, string, error) {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return "", "", errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to get cluster client while getting workload cluster credentials")
	}

	log.V(3).Infof("Retrieving credentials for workload cluster %s \n", options.ClusterName)
	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}
	lock, err := utils.GetFileLockWithTimeOut(filepath.Join(c.tkgConfigDir, constants.LocalTanzuFileLock), utils.DefaultLockTimeout)
	if err != nil {
		return "", "", errors.Wrap(err, "cannot acquire lock for merging workload cluster kubeconfig")
	}

	defer func() {
		if err := lock.Unlock(); err != nil {
			log.Warningf("cannot release lock for merging workload cluster kubeconfig, reason: %v", err)
		}
	}()
	getKubeconfigPollOptions := &clusterclient.PollOptions{Interval: time.Second, Timeout: 3 * time.Second}
	kubeconfig, err := clusterClient.GetKubeConfigForCluster(options.ClusterName, options.Namespace, getKubeconfigPollOptions)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to get cluster kubeconfig. Make sure the cluster name and namespace is correct")
	}

	log.V(3).Infof("Merging credentials into kubeconfig file %s \n", clusterClient.GetCurrentKubeconfigFile())
	err = MergeKubeConfigWithoutSwitchContext(kubeconfig, options.ExportFile)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to merge cluster kubeconfig into the current kubeconfig path")
	}

	newConfig, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return "", "", errors.Wrap(err, "unable to get context from kubeconfig")
	}

	if options.ExportFile != "" {
		return newConfig.CurrentContext, options.ExportFile, nil
	}

	return newConfig.CurrentContext, "", nil
}

// DeleteWorkloadCluster deletes workload cluster
func (c *TkgClient) DeleteWorkloadCluster(options DeleteWorkloadClusterOptions) error {
	currentRegion, err := c.GetCurrentRegionContext()
	if err != nil {
		return errors.Wrap(err, "cannot get current management cluster context")
	}
	clusterclientOptions := clusterclient.Options{
		GetClientInterval: 1 * time.Second,
		GetClientTimeout:  3 * time.Second,
	}
	clusterClient, err := clusterclient.NewClient(currentRegion.SourceFilePath, currentRegion.ContextName, clusterclientOptions)
	if err != nil {
		return errors.Wrap(err, "unable to get cluster client while deleting cluster")
	}

	if options.Namespace == "" {
		options.Namespace = constants.DefaultNamespace
	}

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err == nil && !isPacific {
		err = deleteAutoscalerDeploymentIfPresent(clusterClient, options.ClusterName, options.Namespace)
		if err != nil {
			log.Warningf("failed to delete autoscaler resources from management cluster, reason: %v", err)
		}

		err = deleteCRSObjectsIfPresent(clusterClient, options.ClusterName, options.Namespace)
		if err != nil {
			log.Warningf("failed to delete ClusterResourceSet objects from management cluster, reason: %v", err)
		}
	}

	return clusterClient.DeleteCluster(options.ClusterName, options.Namespace)
}

func deleteAutoscalerDeploymentIfPresent(clusterClient clusterclient.Client, clusterName, namespace string) error {
	var autoScalerDeployment appsv1.Deployment
	autoScalerDeploymentName := clusterName + "-cluster-autoscaler"
	err := clusterClient.GetResource(&autoScalerDeployment, autoScalerDeploymentName, namespace, nil, nil)
	if err != nil && apierrors.IsNotFound(err) {
		log.V(4).Infof("autoscaler deployment '%s' is not present on the management cluster", autoScalerDeploymentName)
		return nil
	}
	if err != nil {
		return errors.Wrapf(err, "failed to retrieve the autoscaler deployment '%s' from management cluster", autoScalerDeploymentName)
	}

	err = clusterClient.DeleteResource(&autoScalerDeployment)
	if err != nil {
		return errors.Wrapf(err, "failed to delete autoscaler deployment '%s' from management cluster", autoScalerDeploymentName)
	}

	// delete service account
	autoScalerServiceAccount := &corev1.ServiceAccount{}
	autoScalerServiceAccountName := clusterName + "-autoscaler"
	autoScalerServiceAccount.Name = autoScalerServiceAccountName
	autoScalerServiceAccount.Namespace = namespace

	err = clusterClient.DeleteResource(autoScalerServiceAccount)
	if err != nil {
		return errors.Wrapf(err, "failed to delete autoscaler serviceaccount '%s' from management cluster", autoScalerServiceAccountName)
	}

	// delete clusterrolebinding
	autoScalerManagementRoleBinding := &rbacv1.ClusterRoleBinding{}
	autoScalerManagementRoleBindingName := clusterName + "-autoscaler-management"
	autoScalerManagementRoleBinding.Name = autoScalerManagementRoleBindingName

	err = clusterClient.DeleteResource(autoScalerManagementRoleBinding)
	if err != nil {
		return errors.Wrapf(err, "failed to delete autoscaler clusterrolebinding '%s' from management cluster", autoScalerManagementRoleBindingName)
	}

	autoScalerWorkloadRoleBinding := &rbacv1.ClusterRoleBinding{}
	autoScalerWorkloadRoleBindingName := clusterName + "-autoscaler-workload"
	autoScalerWorkloadRoleBinding.Name = autoScalerWorkloadRoleBindingName

	err = clusterClient.DeleteResource(autoScalerWorkloadRoleBinding)
	if err != nil {
		return errors.Wrapf(err, "failed to delete autoscaler clusterrolebinding '%s' from management cluster", autoScalerWorkloadRoleBindingName)
	}

	log.Infof("successfully deleted autoscaler resources from management cluster")
	return nil
}

func deleteCRSObjectsIfPresent(clusterClient clusterclient.Client, clusterName, namespace string) error {
	clusterResourceSetList := &addonsv1.ClusterResourceSetList{}
	// Get list of all CRS objects for the cluster
	err := clusterClient.GetResourceList(clusterResourceSetList, clusterName, namespace, nil, nil)
	if err != nil {
		return errors.Wrap(err, "unable to get list of ClusterResourceSet for the cluster while deleting CRS objects")
	}

	// Deletes CRS objects for the cluster one by one
	errorList := []error{}
	for i := range clusterResourceSetList.Items {
		err = clusterClient.DeleteResource(clusterResourceSetList.Items[i].DeepCopy())
		if err != nil {
			errorList = append(errorList, errors.Wrapf(err, "unable to delete ClusterResourceSet '%s'", clusterResourceSetList.Items[i].Name))
		}
	}

	// Throw error if any
	if len(errorList) > 0 {
		return kerrors.NewAggregate(errorList)
	}

	log.V(3).Infof("successfully deleted ClusterResourceSet objects associated with cluster '%s'", clusterName)
	return nil
}
