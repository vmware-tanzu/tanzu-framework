// Copyright 2021 VMware, Inc. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package discovery

import (
	"strings"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/testapigroup"
	"k8s.io/apimachinery/pkg/runtime"
	apitest "k8s.io/apimachinery/pkg/test"
	fakediscovery "k8s.io/client-go/discovery/fake"
	dynamicFake "k8s.io/client-go/dynamic/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestClusterDiscoveryTable(t *testing.T) {
	var mySchema = Schema("mytestschema", partialSchema)

	type testCase struct {
		queryTargets    []QueryTarget
		discoveryClient *fakediscovery.FakeDiscovery
		dynamicClient   *dynamicFake.FakeDynamicClient
		err             string
		pass            bool
	}

	var fakeTesting = &k8stesting.Fake{
		Resources: apiResources,
	}
	var fakeBrokenTesting = &k8stesting.Fake{
		Resources: []*metav1.APIResourceList{},
	}
	scheme, _ := apitest.TestScheme()

	// Scheme with the correct data
	var fakeClientCarp = dynamicFake.NewSimpleDynamicClient(scheme, testObjects...)
	var discoveryClientCarp = &fakediscovery.FakeDiscovery{Fake: fakeTesting}

	// Scheme without any correct data
	var fakeClientBroken = dynamicFake.NewSimpleDynamicClient(scheme)
	var discoveryClientBroken = &fakediscovery.FakeDiscovery{Fake: fakeBrokenTesting}

	// Scheme with partially correct data
	var fakeClientHalf = dynamicFake.NewSimpleDynamicClient(scheme, testObjects...)
	var discoveryClientHalf = &fakediscovery.FakeDiscovery{Fake: fakeBrokenTesting}

	var testTable = []testCase{
		{
			discoveryClient: discoveryClientCarp,
			dynamicClient:   fakeClientCarp,
			pass:            true,
			queryTargets:    []QueryTarget{testGVR1, testResource1, mySchema},
		},
		{
			discoveryClient: discoveryClientBroken,
			dynamicClient:   fakeClientBroken,
			pass:            false,
			err:             "not found",
			queryTargets:    []QueryTarget{testGVR1, testResource1, mySchema},
		},
		{
			discoveryClient: discoveryClientHalf,
			dynamicClient:   fakeClientHalf,
			pass:            false,
			err:             "not found",
			queryTargets:    []QueryTarget{testGVR1, testResource1, mySchema},
		},
		{
			discoveryClient: discoveryClientHalf,
			dynamicClient:   fakeClientHalf,
			pass:            true,
			queryTargets:    []QueryTarget{},
		},
	}

	for _, testCase := range testTable {
		c, err := NewClusterQueryClient(testCase.dynamicClient, testCase.discoveryClient)
		if err != nil {
			t.Fatalf("at=new-query-client err=%q", err)
		}
		q := c.Query(testCase.queryTargets...).Prepare()

		ok, err := q()
		if err != nil {
			if testCase.err == "" {
				t.Fatalf("no error string specified but error obtained err=%q", err)
			}

			if !strings.Contains(err.Error(), testCase.err) {
				t.Fatalf("error string doesnt match partial=%q err=%q", testCase.err, err)
			}
			// error expected and obtained
			continue
		}

		if testCase.pass && !ok {
			t.Fatal("query expected to pass but failed")
		}
	}
}

var carp = corev1.ObjectReference{
	Kind:       "carp",
	Name:       "test14",
	Namespace:  "testns",
	APIVersion: "testapigroup.apimachinery.k8s.io/v1",
}

var testAnnotations = map[string]string{
	"cluster.x-k8s.io/provider": "infrastructure-fake",
}

var testResource1 = Object(&carp).WithAnnotations(testAnnotations) //.WithConditions(field)

var testGVR = testapigroup.Resource("carp").WithVersion("v1")

var testGVR1 = GVR(&testGVR) //.WithFields(field)

var testObjects = []runtime.Object{
	&testapigroup.Carp{
		TypeMeta:   metav1.TypeMeta{Kind: "Carp", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: "test14", Namespace: "testns"},
	},
}

var apiResources = []*metav1.APIResourceList{
	{
		GroupVersion: testGVR.GroupVersion().String(),
		APIResources: []metav1.APIResource{
			{
				Name:    "Carps",
				Kind:    "carp",
				Group:   testGVR.Group,
				Version: testGVR.Version,
			},
		},
	},
}

// There doesnt seem to be any support for fake openapiv2 response data at the moment
var partialSchema = `{}`

func FakeDiscoveryClient(t *testing.T) (*ClusterQueryClient, error) {
	fakeTesting := &k8stesting.Fake{
		Resources: apiResources,
	}
	discovery := &fakediscovery.FakeDiscovery{Fake: fakeTesting}
	scheme, _ := apitest.TestScheme()
	fakeClient := dynamicFake.NewSimpleDynamicClient(scheme, testObjects...)
	return NewClusterQueryClient(fakeClient, discovery)
}
