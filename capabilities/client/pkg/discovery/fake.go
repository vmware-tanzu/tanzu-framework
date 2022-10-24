// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	openapi_v2 "github.com/google/gnostic/openapiv2"
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

// fakeWithSchema is the same as fakediscovery.FakeDiscovery except it has a
// different OPenAPISchema() method.
type fakeWithSchema struct {
	*fakediscovery.FakeDiscovery
}

func (fws fakeWithSchema) OpenAPISchema() (*openapi_v2.Document, error) {
	schema := `swagger: '2.0'
info:
  title: 'example schema for test'
  version: '1.3'
paths:
  - '/test/path'
  - '/another/path'
`
	return openapi_v2.ParseDocument([]byte(schema))
}

// NewFakeClusterQueryClient returns a fake ClusterQueryClient for use in tests.
func NewFakeClusterQueryClientWithSchema(resources []*metav1.APIResourceList, scheme *runtime.Scheme, objs []runtime.Object) (*ClusterQueryClient, error) {
	fakeDynamicClient := dynamicFake.NewSimpleDynamicClient(scheme, objs...)

	fakeDiscWithSchema := fakeWithSchema{
		&fakediscovery.FakeDiscovery{
			Fake: &k8stesting.Fake{
				Resources: nil,
			},
		},
	}
	fakeDiscoveryClient := &fakeDiscWithSchema
	return NewClusterQueryClient(fakeDynamicClient, fakeDiscoveryClient)
}
