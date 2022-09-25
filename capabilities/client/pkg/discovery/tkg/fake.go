// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package tkg

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrlfake "sigs.k8s.io/controller-runtime/pkg/client/fake"

	runv1alpha1 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha1"
	runv1alpha2 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha2"
	runv1alpha3 "github.com/vmware-tanzu/tanzu-framework/apis/run/v1alpha3"
	"github.com/vmware-tanzu/tanzu-framework/capabilities/client/pkg/discovery"
)

var (
	// tanzuRunAPIResourceList is a list of run.tanzu.vmware.com resources for use in tests.
	tanzuRunAPIResourceList = []*metav1.APIResourceList{
		{
			GroupVersion: runv1alpha1.GroupVersion.String(),
			APIResources: []metav1.APIResource{
				{
					Name:       "tanzukubernetesreleases",
					Kind:       "TanzuKubernetesRelease",
					Group:      runv1alpha1.GroupVersion.Group,
					Version:    runv1alpha1.GroupVersion.Version,
					Namespaced: true,
				},
				{
					Name:       "tanzukubernetesclusters",
					Kind:       "TanzuKubernetesCluster",
					Group:      runv1alpha1.GroupVersion.Group,
					Version:    runv1alpha1.GroupVersion.Version,
					Namespaced: true,
				},
				{
					Name:       "tanzukubernetesreleases",
					Kind:       "TanzuKubernetesRelease",
					Group:      runv1alpha2.GroupVersion.Group,
					Version:    runv1alpha2.GroupVersion.Version,
					Namespaced: true,
				},
				{
					Name:       "tanzukubernetesclusters",
					Kind:       "TanzuKubernetesCluster",
					Group:      runv1alpha2.GroupVersion.Group,
					Version:    runv1alpha2.GroupVersion.Version,
					Namespaced: true,
				},
				{
					Name:       "tanzukubernetesreleases",
					Kind:       "TanzuKubernetesRelease",
					Group:      runv1alpha3.GroupVersion.Group,
					Version:    runv1alpha3.GroupVersion.Version,
					Namespaced: true,
				},
				{
					Name:       "tanzukubernetesclusters",
					Kind:       "TanzuKubernetesCluster",
					Group:      runv1alpha3.GroupVersion.Group,
					Version:    runv1alpha3.GroupVersion.Version,
					Namespaced: true,
				},
			},
		},
	}

	// coreAPIResourceList is a list of corev1 resources for use in tests.
	coreAPIResourceList = []*metav1.APIResourceList{
		{
			GroupVersion: "v1",
			APIResources: []metav1.APIResource{
				{
					Name:    "namespaces",
					Kind:    "Namespace",
					Group:   "",
					Version: "v1",
				},
			},
		},
	}
)

// newFakeDiscoveryClient returns a fake DiscoveryClient for use in tests.
func newFakeDiscoveryClient(resources []*metav1.APIResourceList, scheme *runtime.Scheme, objs []runtime.Object) (*DiscoveryClient, error) {
	fakeK8SClient := ctrlfake.NewClientBuilder().WithScheme(scheme).WithRuntimeObjects(objs...).Build()
	fakeQueryClient, err := discovery.NewFakeClusterQueryClient(resources, scheme, objs)
	if err != nil {
		return nil, err
	}
	return NewDiscoveryClient(fakeK8SClient, fakeQueryClient), nil
}
