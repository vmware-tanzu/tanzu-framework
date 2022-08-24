// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkgctl

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	configv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/config/v1alpha1"
	"github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/clusterclient"
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

	return featureActivatedInNamespace(reqContext, clusterClient.GetClientSet(), namespace, feature)
}

// TODO(ragasthya): Remove temp copies of featuregate helper functions.
// featureActivatedInNamespace returns true only if all of the features specified are activated in the namespace.
func featureActivatedInNamespace(ctx context.Context, c client.Client, namespace, feature string) (bool, error) {
	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{Key: "kubernetes.io/metadata.name", Operator: metav1.LabelSelectorOpIn, Values: []string{namespace}},
		},
	}
	return featuresActivatedInNamespacesMatchingSelector(ctx, c, selector, []string{feature})
}

// featuresActivatedInNamespacesMatchingSelector returns true only if all the features specified are activated in every namespace matched by the selector.
func featuresActivatedInNamespacesMatchingSelector(ctx context.Context, c client.Client, namespaceSelector metav1.LabelSelector, features []string) (bool, error) {
	namespaces, err := namespacesMatchingSelector(ctx, c, &namespaceSelector)
	if err != nil {
		return false, err
	}

	// If no namespaces are matched or no features specified, return false.
	if len(namespaces) == 0 || len(features) == 0 {
		return false, nil
	}

	featureGatesList := &configv1alpha1.FeatureGateList{}
	if err := c.List(ctx, featureGatesList); err != nil {
		return false, err
	}

	// Map of namespace to a set of features activated in that namespace.
	namespaceToActivatedFeatures := make(map[string]sets.String)
	for i := range featureGatesList.Items {
		fg := featureGatesList.Items[i]
		for _, namespace := range fg.Status.Namespaces {
			namespaceToActivatedFeatures[namespace] = sets.NewString(fg.Status.ActivatedFeatures...)
		}
	}

	for _, ns := range namespaces {
		activatedFeatures, found := namespaceToActivatedFeatures[ns]
		if !found {
			// Namespace has no features gated.
			return false, nil
		}
		// Feature is not activated in this namespace.
		if !activatedFeatures.HasAll(features...) {
			return false, nil
		}
	}
	return true, nil
}

// namespacesMatchingSelector returns the list of namespaces after applying the NamespaceSelector filter.
// Note that a nil selector selects nothing, while an empty selector selects everything.
// Callers using this function in feature gates context should be sending a pointer to an empty selector instead of nil.
func namespacesMatchingSelector(ctx context.Context, c client.Client, selector *metav1.LabelSelector) ([]string, error) {
	s, err := metav1.LabelSelectorAsSelector(selector)
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from NamespaceSelector: %w", err)
	}

	nsList := &corev1.NamespaceList{}
	err = c.List(ctx, nsList, client.MatchingLabelsSelector{Selector: s})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespaces from NamespaceSelector: %w", err)
	}

	var namespaces []string
	for i := range nsList.Items {
		namespaces = append(namespaces, nsList.Items[i].Name)
	}
	return namespaces, nil
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
