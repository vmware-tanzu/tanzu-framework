// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

// NewFakeClusterQueryClient returns a fake ClusterQueryClient for use in tests.
func NewFakeClusterQueryClient(resources []*metav1.APIResourceList, scheme *runtime.Scheme, objs []runtime.Object) (*ClusterQueryClient, error) {
	fakeDynamicClient := dynamicFake.NewSimpleDynamicClient(scheme, objs...)
	fakeDiscoveryClient := &fakediscovery.FakeDiscovery{
		Fake: &k8stesting.Fake{
			Resources: resources,
		}}
	return NewClusterQueryClient(fakeDynamicClient, fakeDiscoveryClient)
}
