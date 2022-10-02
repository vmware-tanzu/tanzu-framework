package testutil

import (
	"fmt"

	openapiv2 "github.com/googleapis/gnostic/openapiv2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// FakeDiscovery customize the behavior of fake client-go FakeDiscovery.ServerPreferredResources to return
// a customized APIResourceList.
// The client-go FakeDiscovery.ServerPreferredResources is hardcoded to return nil.
// https://github.com/kubernetes/client-go/blob/master/discovery/fake/discovery.go#L85
type FakeDiscovery struct {
	FakeDiscovery discovery.DiscoveryInterface
	Resources     []*metav1.APIResourceList
	APIGroups     *metav1.APIGroupList
}

var _ discovery.DiscoveryInterface = &FakeDiscovery{}

func (c FakeDiscovery) RESTClient() rest.Interface {
	return c.FakeDiscovery.RESTClient()
}

func (c FakeDiscovery) ServerGroups() (*metav1.APIGroupList, error) {
	return c.APIGroups, nil
}

func (c FakeDiscovery) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	return c.FakeDiscovery.ServerGroupsAndResources()
}

func (c FakeDiscovery) ServerVersion() (*version.Info, error) {
	return c.FakeDiscovery.ServerVersion()
}

func (c FakeDiscovery) OpenAPISchema() (*openapiv2.Document, error) {
	return c.FakeDiscovery.OpenAPISchema()
}

func (c FakeDiscovery) getFakeServerPreferredResources() []*metav1.APIResourceList {
	return c.Resources
}

func (c FakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	for _, res := range c.Resources {
		if res.GroupVersion == groupVersion {
			return res, nil
		}
	}
	return nil, fmt.Errorf("no matching resources")
}

// Having nolint below to get rid of the complaining on the deprecation of ServerResources. We have to have the following
// function to customize the DiscoveryInterface
//
//nolint:staticcheck
func (c FakeDiscovery) ServerResources() ([]*metav1.APIResourceList, error) {
	return c.FakeDiscovery.ServerResources()
}

func (c FakeDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return c.getFakeServerPreferredResources(), nil
}

func (c FakeDiscovery) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return c.FakeDiscovery.ServerPreferredNamespacedResources()
}

type FakeGVRHelper struct {
	DiscoveryClient discovery.DiscoveryInterface
}

func (f *FakeGVRHelper) GetGVR(gk schema.GroupKind) (*schema.GroupVersionResource, error) {
	apiResourceList, err := f.DiscoveryClient.ServerPreferredResources()
	if err != nil {
		return nil, err
	}
	for _, apiResource := range apiResourceList {
		gv, err := schema.ParseGroupVersion(apiResource.GroupVersion)
		if err != nil {
			return nil, err
		}
		if gv.Group == gk.Group {
			for i := 0; i < len(apiResource.APIResources); i++ {
				if apiResource.APIResources[i].Kind == gk.Kind {
					return &schema.GroupVersionResource{Group: gv.Group, Resource: apiResource.APIResources[i].Name, Version: gv.Version}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("unable to find server preferred resource %s/%s", gk.Group, gk.Kind)
}

func (f *FakeGVRHelper) GetDiscoveryClient() discovery.DiscoveryInterface {
	return f.DiscoveryClient
}
