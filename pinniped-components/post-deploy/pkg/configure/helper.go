// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package configure

import (
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/kubernetes"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/retry"

	"github.com/vmware-tanzu/tanzu-framework/pinniped-components/post-deploy/pkg/constants"

	"github.com/vmware-tanzu/tanzu-framework/pinniped-components/post-deploy/pkg/configure/supervisor"
)

// createOrUpdateManagementClusterPinnipedInfo creates Pinniped information or updates existing data for a management cluster.
func createOrUpdateManagementClusterPinnipedInfo(ctx context.Context, pinnipedInfo supervisor.PinnipedInfo, k8sClientSet kubernetes.Interface, supervisorNamespaceName string) error {
	var err error
	zap.S().Info("Creating the ConfigMap for Pinniped info")

	data, err := json.Marshal(pinnipedInfo)
	if err != nil {
		err = fmt.Errorf("could not marshal Pinniped info into JSON: %w", err)
		zap.S().Error(err)
		return err
	}
	dataMap := make(map[string]string)
	if err = json.Unmarshal(data, &dataMap); err != nil {
		err = fmt.Errorf("could not unmarshal Pinniped info into map[string]string: %w", err)
		zap.S().Error(err)
		return err
	}

	// Get the Supervisor's namespace so we can use it as the ownerRef for the ConfigMap.
	//
	// In TKGm classy management clusters, the pinniped-info ConfigMap is created here by this Job.
	// If the user configures the management cluster back to the default of identity_management_type=none,
	// then the pinniped-supervisor namespace is deleted, but nothing deletes the pinniped-info ConfigMap.
	// To cause the ConfigMap to be deleted in this case, we set its ownerRef to point to the pinniped-supervisor namespace.
	// When the ConfigMap is deleted, the v3_cascade_controller will update the pinniped addon secret of all workload
	// clusters to have the default content of identity_management_type=none.
	//
	// In TKGm legacy management clusters, the pinniped-info ConfigMap is created here by this Job.
	// When the user deletes the pinniped addon secret, the pinniped addon will be deleted, including the
	// pinniped-supervisor namespace. When the ConfigMap is deleted, the v1_cascade_controller will delete the pinniped
	// addon secret of all workload clusters.
	//
	// In a TKGs management clusters, this Job is not used and something else is responsible for creating/updating/deleting
	// the pinniped-info ConfigMap.
	supervisorNamespace, err := k8sClientSet.CoreV1().Namespaces().Get(ctx, supervisorNamespaceName, metav1.GetOptions{})
	if err != nil {
		err = fmt.Errorf("could not get namespace %s: %w", supervisorNamespaceName, err)
		zap.S().Error(err)
		return err
	}

	// create configmap under kube-public namespace
	pinnipedConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PinnipedInfoConfigMapName,
			Namespace: constants.KubePublicNamespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: supervisorNamespace.APIVersion,
					Kind:       supervisorNamespace.Kind,
					Name:       supervisorNamespace.Name,
					UID:        supervisorNamespace.UID,
				},
			},
		},
		Data: dataMap,
	}

	if _, err = k8sClientSet.CoreV1().ConfigMaps(constants.KubePublicNamespace).Get(ctx, constants.PinnipedInfoConfigMapName, metav1.GetOptions{}); err != nil {
		if errors.IsNotFound(err) {
			// create if does not exist
			if _, err = k8sClientSet.CoreV1().ConfigMaps(constants.KubePublicNamespace).Create(ctx, pinnipedConfigMap, metav1.CreateOptions{}); err != nil {
				err = fmt.Errorf("could not create pinniped-info configmap: %w", err)
				zap.S().Error(err)
				return err
			}

			zap.S().Infof("Created the ConfigMap %s/%s for Pinniped info", constants.KubePublicNamespace, constants.PinnipedInfoConfigMapName)
			return nil
		}
		// return err if could not get the configmap due to other errors
		err = fmt.Errorf("could not get pinniped-info configmap: %w", err)
		zap.S().Error(err)
		return err
	}

	// if we have configmap fetched, try to update
	err = retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var e error
		var configMapUpdated *corev1.ConfigMap
		if configMapUpdated, e = k8sClientSet.CoreV1().ConfigMaps(constants.KubePublicNamespace).Get(ctx, constants.PinnipedInfoConfigMapName, metav1.GetOptions{}); e != nil {
			return e
		}
		configMapUpdated.OwnerReferences = pinnipedConfigMap.OwnerReferences
		configMapUpdated.Data = pinnipedConfigMap.Data
		_, e = k8sClientSet.CoreV1().ConfigMaps(constants.KubePublicNamespace).Update(ctx, configMapUpdated, metav1.UpdateOptions{})
		return e
	})
	if err != nil {
		err = fmt.Errorf("could not update pinniped-info configmap: %w", err)
		zap.S().Error(err)
		return err
	}

	zap.S().Infof("Updated the ConfigMap %s/%s for Pinniped info", constants.KubePublicNamespace, constants.PinnipedInfoConfigMapName)
	return nil
}
