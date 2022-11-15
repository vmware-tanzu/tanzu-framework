// Copyright 2022 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

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

const (
	defaultDataValueNameSpace          = "kube-system"
	defaultDataValueDeploymentReplicas = 3
)

// getOwnerCluster verifies that the AwsEbsCSIConfig has a cluster as its owner reference,
// and returns the cluster. It tries to read the cluster name from the AwsEbsCSIConfig's owner reference objects.
// If not there, we assume the owner cluster and AwsEbsCSIConfig always has the same name.
func (r *AwsEbsCSIConfigReconciler) getOwnerCluster(ctx context.Context,
	awsebsCSIConfig *csiv1alpha1.AwsEbsCSIConfig) (*clusterapiv1beta1.Cluster, error) {

	logger := log.FromContext(ctx)
	cluster := &clusterapiv1beta1.Cluster{}
	// TODO donot support using addon name as cluster name
	clusterName := awsebsCSIConfig.Name

	// retrieve the owner cluster for the AwsEbsCSIConfig object
	for _, ownerRef := range awsebsCSIConfig.GetOwnerReferences() {
		if strings.EqualFold(ownerRef.Kind, constants.ClusterKind) {
			clusterName = ownerRef.Name
			break
		}
	}
	if err := r.Client.Get(ctx, types.NamespacedName{Namespace: awsebsCSIConfig.Namespace, Name: clusterName}, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("Cluster resource '%s/%s' not found", awsebsCSIConfig.Namespace, clusterName))
			return nil, nil
		}
		logger.Error(err, fmt.Sprintf("Unable to fetch cluster '%s/%s'", awsebsCSIConfig.Namespace, clusterName))
		return nil, err
	}

	return cluster, nil
}

// mapAwsEbsCSIConfigToDataValues maps AwsEbsCSIConfig CR to data values
func (r *AwsEbsCSIConfigReconciler) mapAwsEbsCSIConfigToDataValues(ctx context.Context,
	awsEbsCSIConfig *csiv1alpha1.AwsEbsCSIConfig,
	cluster *clusterapiv1beta1.Cluster) (*DataValues, error) {

	dvs := &DataValues{
		AwsEbsCSI: &DataValuesAwsEbsCSI{
			Namespace:          defaultDataValueNameSpace,
			DeploymentReplicas: defaultDataValueDeploymentReplicas,
			HTTPProxy:          "",
			HTTPSProxy:         "",
			NoProxy:            "",
		},
	}

	// TODO do we need to check namespace's existense here ? leave it not checked, in case user may create namespacd asyncly
	if awsEbsCSIConfig.Spec.AwsEbsCSI.Namespace != "" {
		dvs.AwsEbsCSI.Namespace = awsEbsCSIConfig.Spec.AwsEbsCSI.Namespace
	}

	dvs.AwsEbsCSI.HTTPProxy = awsEbsCSIConfig.Spec.AwsEbsCSI.HTTPProxy
	dvs.AwsEbsCSI.HTTPSProxy = awsEbsCSIConfig.Spec.AwsEbsCSI.HTTPSProxy
	dvs.AwsEbsCSI.NoProxy = awsEbsCSIConfig.Spec.AwsEbsCSI.NoProxy

	if awsEbsCSIConfig.Spec.AwsEbsCSI.DeploymentReplicas != nil {
		dvs.AwsEbsCSI.DeploymentReplicas = *awsEbsCSIConfig.Spec.AwsEbsCSI.DeploymentReplicas
	}

	return dvs, nil
}
