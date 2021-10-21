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

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/constants"

	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/post-deploy/pkg/configure/supervisor"
)

// createOrUpdatePinnipedInfo creates Pinniped information or updates existing data.
func createOrUpdatePinnipedInfo(ctx context.Context, pinnipedInfo supervisor.PinnipedInfo, k8sClientSet kubernetes.Interface) error {
	var err error
	zap.S().Info("Creating the ConfigMap for Pinniped info")
	data, err := json.Marshal(pinnipedInfo)
	if err != nil {
		err = fmt.Errorf("could not marshal Pinniped info into JSON: %w", err)
		return err
	}
	dataMap := make(map[string]string)
	if err = json.Unmarshal(data, &dataMap); err != nil {
		err = fmt.Errorf("could not unmarshal Pinniped info into map[string]string: %w", err)
		return err
	}
	// create configmap under kube-public namespace
	pinnipedConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.PinnipedInfoConfigMapName,
			Namespace: constants.KubePublicNamespace,
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
