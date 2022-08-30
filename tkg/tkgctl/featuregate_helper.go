// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"context"

	"github.com/pkg/errors"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/featuregate"
	"github.com/vmware-tanzu/tanzu-framework/tkg/clusterclient"
	"github.com/vmware-tanzu/tanzu-framework/tkg/constants"
)

type featureGateHelper struct {
	clusterClientOptions clusterclient.Options
	contextName          string
	kubeconfig           string
}

type FeatureGateHelper interface {
	FeatureActivatedInNamespace(reqContext context.Context, feature, namespace string) (bool, error)
}

func NewFeatureGateHelper(options *clusterclient.Options, contextName, kubeconfig string) FeatureGateHelper {
	return &featureGateHelper{
		clusterClientOptions: *options,
		contextName:          contextName,
		kubeconfig:           kubeconfig,
	}
}

// FeatureActivatedInNamespace gets and returns the status of the feature in namespace
func (fg *featureGateHelper) FeatureActivatedInNamespace(reqContext context.Context, feature, namespace string) (bool, error) {
	clusterClient, err := clusterclient.NewClient(fg.kubeconfig, fg.contextName, fg.clusterClientOptions)
	if err != nil {
		return false, err
	}
	scheme := clusterClient.GetClientSet().Scheme()
	_ = v1alpha1.AddToScheme(scheme)

	isPacific, err := clusterClient.IsPacificRegionalCluster()
	if err != nil {
		return false, errors.Wrap(err, "error determining Tanzu Kubernetes Cluster service for vSphere management cluster ")
	}

	if isPacific {
		// If requested feature is `ClusterClassFeature` or `TKCAPIFeature` it gets decided on different criteria at the moment
		// This is because we are unable to query Featuregates for the devops users on the TKGS clusters because of the permission restrictions
		// TODO: Fix this logic to rely on actual featuregates once permission issue for the devops user has been resolved
		switch feature {
		case constants.ClusterClassFeature:
			return fg.isClusterClassFeatureActivated(clusterClient, feature, namespace)
		case constants.TKCAPIFeature:
			return fg.isTKCFeatureActivated(clusterClient, feature, namespace)
		}
	}

	return featuregate.FeatureActivatedInNamespace(reqContext, clusterClient.GetClientSet(), namespace, feature)
}

// isClusterClassFeatureActivated is decided based on the FeatureGate CRD is present or not
// Currently, we are unable to query the featuregate for devops users because of the permission
// restrictions on the clusters. Hence for TKGS cluster, we are relying on existence of Featuregate
// API to decide ClusterClass feature is supported or not.
// Assumption is only vSphere8 based TKGS clusters will have this featuregate API as well as support for
// ClusterClass feature
func (fg *featureGateHelper) isClusterClassFeatureActivated(clusterClient clusterclient.Client, feature, namespace string) (bool, error) {
	isFeaturegateCRDExists, err := clusterClient.VerifyExistenceOfCRD("featuregates", configv1alpha1.GroupVersion.Group)
	if err != nil {
		return false, err
	}
	return isFeaturegateCRDExists, nil
}

// isTKCFeatureActivated returns true always for TKGS based cluster
// Assumption is both vSphere7 and vSphere8 based TKGS will continue to support TKC based cluster creation
func (fg *featureGateHelper) isTKCFeatureActivated(clusterClient clusterclient.Client, feature, namespace string) (bool, error) {
	return true, nil
}
