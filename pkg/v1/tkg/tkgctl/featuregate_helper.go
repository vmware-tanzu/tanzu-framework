// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"context"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/sdk/features/featuregate"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
)

type featureGateHelper struct {
	clusterClientOptions clusterclient.Options
	contextName          string
	kubeconfig           string
}

type FeatureGateHelper interface {
	FeatureActivatedInNamespace(reqContext context.Context, feature, namespace string) (bool, error)
}

func newFeatureGateHelper(options *clusterclient.Options, contextName, kubeconfig string) FeatureGateHelper {
	return &featureGateHelper{
		clusterClientOptions: *options,
		contextName:          contextName,
		kubeconfig:           kubeconfig,
	}
}

// FeatureActivatedInNamespace gets and returns the status of the feature in namespace
func (fg *featureGateHelper) FeatureActivatedInNamespace(reqContext context.Context, feature, namespace string) (bool, error) {
	clusterClient, _ := clusterclient.NewClient(fg.kubeconfig, fg.contextName, fg.clusterClientOptions)
	scheme := clusterClient.GetClientSet().Scheme()
	_ = v1alpha1.AddToScheme(scheme)
	return featuregate.FeatureActivatedInNamespace(reqContext, clusterClient.GetClientSet(), namespace, feature)
}
